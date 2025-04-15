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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// -----------------------------
// EmailCycleCollector
// -----------------------------

type EmailCycleData struct {
	CycleStart    string  `json:"cycle_start"`
	CycleEnd      string  `json:"cycle_end"`
	CycleUsed     float64 `json:"cycle_used"`
	CycleRemaining float64 `json:"cycle_remaining"`
	CycleMax      float64 `json:"cycle_max"`
}

type EmailCycleResponse struct {
	RequestID string          `json:"request_id"`
	Data      EmailCycleData  `json:"data"`
}

type EmailCycleCollector struct {
	mutex      sync.Mutex
	apiURL     string
	apiKey     string
	debug      bool
	namespace  string

	used            prometheus.Gauge
	remaining       prometheus.Gauge
	max             prometheus.Gauge
	remainingSeconds prometheus.Gauge
}

func NewEmailCycleCollector(apiURL, apiKey string, debug bool) *EmailCycleCollector {
	ns := "smtp2go_email_cycle"

	return &EmailCycleCollector{
		apiURL:    apiURL,
		apiKey:    apiKey,
		debug:     debug,
		namespace: ns,
		used: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "used",
			Help:      "Number of emails used in the current cycle",
		}),
		remaining: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "remaining",
			Help:      "Number of emails remaining in the current cycle",
		}),
		max: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "max",
			Help:      "Maximum number of emails allowed in the current cycle",
		}),
		remainingSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "remaining_seconds",
			Help:      "Seconds remaining until the end of the current cycle",
		}),
	}
}

func (c *EmailCycleCollector) Describe(ch chan<- *prometheus.Desc) {
	c.used.Describe(ch)
	c.remaining.Describe(ch)
	c.max.Describe(ch)
	c.remainingSeconds.Describe(ch)
}

func (c *EmailCycleCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	endpoint := c.apiURL + "/stats/email_cycle"

	reqBody, _ := json.Marshal(map[string]string{"api_key": c.apiKey})
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[email_cycle] HTTP request failed:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if c.debug {
		log.Printf("[email_cycle] Raw response: %s\n", string(body))
	}

	var apiResp EmailCycleResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Println("[email_cycle] Failed to parse JSON:", err)
		return
	}

	data := apiResp.Data
	c.used.Set(data.CycleUsed)
	c.remaining.Set(data.CycleRemaining)
	c.max.Set(data.CycleMax)

	endTime, err := time.Parse("2006-01-02 15:04:05-07:00", data.CycleEnd)
	if err != nil {
		log.Println("[email_cycle] Failed to parse cycle_end timestamp:", err)
	} else {
		remainingSeconds := endTime.Sub(time.Now()).Seconds()
		c.remainingSeconds.Set(remainingSeconds)
	}

	c.used.Collect(ch)
	c.remaining.Collect(ch)
	c.max.Collect(ch)
	c.remainingSeconds.Collect(ch)
}

// -----------------------------
// EmailBouncesCollector
// -----------------------------

type EmailBouncesData struct {
	Emails        float64 `json:"emails"`
	Rejects       float64 `json:"rejects"`
	SoftBounces   float64 `json:"softbounces"`
	HardBounces   float64 `json:"hardbounces"`
	BouncePercent string  `json:"bounce_percent"`
}

type EmailBouncesResponse struct {
	RequestID string            `json:"request_id"`
	Data      EmailBouncesData  `json:"data"`
}

type EmailBouncesCollector struct {
	mutex     sync.Mutex
	apiURL    string
	apiKey    string
	debug     bool
	namespace string

	emails        prometheus.Gauge
	rejects       prometheus.Gauge
	softBounces   prometheus.Gauge
	hardBounces   prometheus.Gauge
	bouncePercent prometheus.Gauge
}

func NewEmailBouncesCollector(apiURL, apiKey string, debug bool) *EmailBouncesCollector {
	ns := "smtp2go_email_bounces"

	return &EmailBouncesCollector{
		apiURL:    apiURL,
		apiKey:    apiKey,
		debug:     debug,
		namespace: ns,
		emails: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "emails",
			Help:      "Number of emails processed",
		}),
		rejects: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "rejects",
			Help:      "Number of rejected emails",
		}),
		softBounces: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "softbounces",
			Help:      "Number of soft bounces",
		}),
		hardBounces: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "hardbounces",
			Help:      "Number of hard bounces",
		}),
		bouncePercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "bounce_percent",
			Help:      "Percentage of bounced emails",
		}),
	}
}

func (c *EmailBouncesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.emails.Describe(ch)
	c.rejects.Describe(ch)
	c.softBounces.Describe(ch)
	c.hardBounces.Describe(ch)
	c.bouncePercent.Describe(ch)
}

func (c *EmailBouncesCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	endpoint := c.apiURL + "/stats/email_bounces"

	reqBody, _ := json.Marshal(map[string]string{"api_key": c.apiKey})
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[email_bounces] HTTP request failed:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if c.debug {
		log.Printf("[email_bounces] Raw response: %s\n", string(body))
	}

	var apiResp EmailBouncesResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Println("[email_bounces] Failed to parse JSON:", err)
		return
	}

	data := apiResp.Data
	c.emails.Set(data.Emails)
	c.rejects.Set(data.Rejects)
	c.softBounces.Set(data.SoftBounces)
	c.hardBounces.Set(data.HardBounces)

	percent, err := strconv.ParseFloat(data.BouncePercent, 64)
	if err != nil {
		log.Println("[email_bounces] Failed to parse bounce_percent:", err)
	} else {
		c.bouncePercent.Set(percent)
	}

	c.emails.Collect(ch)
	c.rejects.Collect(ch)
	c.softBounces.Collect(ch)
	c.hardBounces.Collect(ch)
	c.bouncePercent.Collect(ch)
}

func main() {
	apiURL := flag.String("apiURL", "https://api.smtp2go.com/v3", "Base URL of the API (e.g., https://api.smtp2go.com/v3)")
	apiKey := flag.String("apiKey", "", "API key for authentication")
	debug := flag.Bool("debug", false, "Enable debug logging")
	listenAddr := flag.String("listen", ":22112", "Address to expose metrics")

	flag.Parse()

	if *apiURL == "" || *apiKey == "" {
		log.Fatal("Option -apiKey must be provided")
	}

	// Remove trailing slash from base URL
	*apiURL = strings.TrimRight(*apiURL, "/")

	// Register all collectors
	emailCycle := NewEmailCycleCollector(*apiURL, *apiKey, *debug)
	emailBounces := NewEmailBouncesCollector(*apiURL, *apiKey, *debug)

	prometheus.MustRegister(emailCycle)
	prometheus.MustRegister(emailBounces)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting exporter on %s...\n", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
