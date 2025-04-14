// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsData struct {
	CycleStart     string  `json:"cycle_start"`
	CycleEnd       string  `json:"cycle_end"`
	CycleUsed      float64 `json:"cycle_used"`
	CycleRemaining float64 `json:"cycle_remaining"`
	CycleMax       float64 `json:"cycle_max"`
}

type ApiResponse struct {
	RequestID string      `json:"request_id"`
	Data      MetricsData `json:"data"`
}

type ApiCollector struct {
	mutex            sync.Mutex
	apiURL           string
	apiKey           string
	debug            bool
	namespace        string
	max              prometheus.Gauge
	remaining        prometheus.Gauge
	used             prometheus.Gauge
	remainingSeconds prometheus.Gauge
}

func NewApiCollector(apiURL, apiKey string, debug bool, namespace string) *ApiCollector {
	ns := "smtp2go_" + namespace

	return &ApiCollector{
		apiURL:    apiURL,
		apiKey:    apiKey,
		debug:     debug,
		namespace: ns,
		max: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "max",
			Help:      "Maximum value from external API",
		}),
		remaining: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "remaining",
			Help:      "Remaining value from external API",
		}),
		used: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "used",
			Help:      "Used value from external API",
		}),
		remainingSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "remaining_seconds",
			Help:      "Seconds remaining until cycle_end",
		}),
	}
}

func (c *ApiCollector) Describe(ch chan<- *prometheus.Desc) {
	c.max.Describe(ch)
	c.remaining.Describe(ch)
	c.used.Describe(ch)
	c.remainingSeconds.Describe(ch)
}

func (c *ApiCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.collectEmailCycleMetrics(ch)

	// Future: add other collectors here with different namespaces
}

func (c *ApiCollector) collectEmailCycleMetrics(ch chan<- prometheus.Metric) {
	endpoint := c.apiURL + "/stats/email_cycle"

	reqBody, err := json.Marshal(map[string]string{
		"api_key": c.apiKey,
	})
	if err != nil {
		log.Println("Failed to build request JSON body:", err)
		return
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println("Failed to create HTTP request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		log.Println("Error making HTTP request to API:", err)
		return
	}
	defer resp.Body.Close()

	if c.debug {
		log.Printf("[email_cycle] HTTP request duration: %v\n", duration)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read API response body:", err)
		return
	}

	if c.debug {
		log.Printf("[email_cycle] Raw JSON response: %s\n", string(body))
	}

	var apiResp ApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Println("Failed to parse API JSON response:", err)
		return
	}

	data := apiResp.Data

	c.max.Set(data.CycleMax)
	c.remaining.Set(data.CycleRemaining)
	c.used.Set(data.CycleUsed)

	endTime, err := time.Parse("2006-01-02 15:04:05-07:00", data.CycleEnd)
	if err != nil {
		log.Println("Failed to parse cycle_end timestamp:", err)
		return
	}

	now := time.Now().UTC()
	diff := endTime.Sub(now).Seconds()
	if diff < 0 {
		diff = 0
	}
	c.remainingSeconds.Set(diff)

	c.max.Collect(ch)
	c.remaining.Collect(ch)
	c.used.Collect(ch)
	c.remainingSeconds.Collect(ch)
}

func main() {
	apiURL := flag.String("api-url", "", "Base API URL (e.g., https://example.com/api)")
	apiKey := flag.String("api-key", "", "API key for authentication")
	listenAddress := flag.String("listen-address", ":2112", "Address to expose metrics (e.g., :2112 or 0.0.0.0:8080)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Trim trailing slash if present
	if len(*apiURL) > 0 && (*apiURL)[len(*apiURL)-1] == '/' {
		*apiURL = (*apiURL)[:len(*apiURL)-1]
	}

	if *apiURL == "" || *apiKey == "" {
		log.Fatal("Both --api-url and --api-key parameters are required")
	}

	log.Printf("Starting exporter with config:\n  API URL: %s/stats/email_cycle\n  Listen: %s\n  Debug: %v\n", *apiURL, *listenAddress, *debug)

	collector := NewApiCollector(*apiURL, *apiKey, *debug, "email_cycle")
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Exporter available at: http://%s/metrics\n", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
