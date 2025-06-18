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

type EmailSpamData struct {
	Emails      float64 `json:"emails"`
	Rejects     float64 `json:"rejects"`
	Spams       float64 `json:"spams"`
	SpamPercent string  `json:"spam_percent"`
}

type EmailSpamResponse struct {
	RequestID string        `json:"request_id"`
	Data      EmailSpamData `json:"data"`
}

type EmailSpamCollector struct {
	mutex     sync.Mutex
	apiURL    string
	apiKey    string
	debug     bool
	namespace string

	emails      prometheus.Gauge
	rejects     prometheus.Gauge
	spams       prometheus.Gauge
	spamPercent prometheus.Gauge
}

func NewEmailSpamCollector(apiURL, apiKey string, debug bool) *EmailSpamCollector {
	ns := "smtp2go_email_spam"

	return &EmailSpamCollector{
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
		spams: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "spams",
			Help:      "Number of emails marked as spam",
		}),
		spamPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "spam_percent",
			Help:      "Percentage of spam emails",
		}),
	}
}

func (c *EmailSpamCollector) Describe(ch chan<- *prometheus.Desc) {
	c.emails.Describe(ch)
	c.rejects.Describe(ch)
	c.spams.Describe(ch)
	c.spamPercent.Describe(ch)
}

func (c *EmailSpamCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	body, err := doPostRequest(c.apiURL, "/stats/email_spam", c.apiKey, c.debug, "email_spam")
	if err != nil {
		return
	}

	var apiResp EmailSpamResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Println("[email_spam] Failed to parse JSON:", err)
		return
	}

	data := apiResp.Data
	c.emails.Set(data.Emails)
	c.rejects.Set(data.Rejects)
	c.spams.Set(data.Spams)

	percent, err := strconv.ParseFloat(data.SpamPercent, 64)
	if err != nil {
		log.Println("[email_spam] Failed to parse spam_percent:", err)
	} else {
		c.spamPercent.Set(percent)
	}

	c.emails.Collect(ch)
	c.rejects.Collect(ch)
	c.spams.Collect(ch)
	c.spamPercent.Collect(ch)
}
