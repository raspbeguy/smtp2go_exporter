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

package internal

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type EmailCycleData struct {
	CycleStart     string  `json:"cycle_start"`
	CycleEnd       string  `json:"cycle_end"`
	CycleUsed      float64 `json:"cycle_used"`
	CycleRemaining float64 `json:"cycle_remaining"`
	CycleMax       float64 `json:"cycle_max"`
}

type EmailCycleResponse struct {
	RequestID string         `json:"request_id"`
	Data      EmailCycleData `json:"data"`
}

type EmailCycleCollector struct {
	mutex     sync.Mutex
	apiURL    string
	apiKey    string
	debug     bool
	namespace string

	used             prometheus.Gauge
	remaining        prometheus.Gauge
	max              prometheus.Gauge
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

	body, err := doPostRequest(c.apiURL, "/stats/email_cycle", c.apiKey, c.debug, "email_cycle")
	if err != nil {
		return
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
		remainingSeconds := time.Until(endTime).Seconds()
		c.remainingSeconds.Set(remainingSeconds)
	}

	c.used.Collect(ch)
	c.remaining.Collect(ch)
	c.max.Collect(ch)
	c.remainingSeconds.Collect(ch)
}
