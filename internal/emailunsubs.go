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
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type EmailUnsubsData struct {
	Emails             float64 `json:"emails"`
	Rejects            float64 `json:"rejects"`
	Unsubscribes       float64 `json:"unsubscribes"`
	UnsubscribePercent string  `json:"unsubscribe_percent"`
}

type EmailUnsubsResponse struct {
	RequestID string          `json:"request_id"`
	Data      EmailUnsubsData `json:"data"`
}

type EmailUnsubsCollector struct {
	mutex     sync.Mutex
	apiURL    string
	apiKey    string
	debug     bool
	namespace string

	emails             prometheus.Gauge
	rejects            prometheus.Gauge
	unsubscribes       prometheus.Gauge
	unsubscribePercent prometheus.Gauge
}

func NewEmailUnsubsCollector(apiURL, apiKey string, debug bool) *EmailUnsubsCollector {
	ns := "smtp2go_email_unsubs"

	return &EmailUnsubsCollector{
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
		unsubscribes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "unsubscribes",
			Help:      "Number of unsubscribes",
		}),
		unsubscribePercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "unsubscribe_percent",
			Help:      "Percentage of unsubscribes",
		}),
	}
}

func (c *EmailUnsubsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.emails.Describe(ch)
	c.rejects.Describe(ch)
	c.unsubscribes.Describe(ch)
	c.unsubscribePercent.Describe(ch)
}

func (c *EmailUnsubsCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	body, err := doPostRequest(c.apiURL, "/stats/email_unsubs", c.apiKey, c.debug, "email_unsubs")
	if err != nil {
		return
	}

	var apiResp EmailUnsubsResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Println("[email_unsubs] Failed to parse JSON:", err)
		return
	}

	data := apiResp.Data
	c.emails.Set(data.Emails)
	c.rejects.Set(data.Rejects)
	c.unsubscribes.Set(data.Unsubscribes)

	percent, err := strconv.ParseFloat(data.UnsubscribePercent, 64)
	if err != nil {
		log.Println("[email_unsubs] Failed to parse unsubscribe_percent:", err)
	} else {
		c.unsubscribePercent.Set(percent)
	}

	c.emails.Collect(ch)
	c.rejects.Collect(ch)
	c.unsubscribes.Collect(ch)
	c.unsubscribePercent.Collect(ch)
}
