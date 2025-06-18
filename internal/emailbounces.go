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

type EmailBouncesData struct {
	Emails        float64 `json:"emails"`
	Rejects       float64 `json:"rejects"`
	SoftBounces   float64 `json:"softbounces"`
	HardBounces   float64 `json:"hardbounces"`
	BouncePercent string  `json:"bounce_percent"`
}

type EmailBouncesResponse struct {
	RequestID string           `json:"request_id"`
	Data      EmailBouncesData `json:"data"`
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

	body, err := doPostRequest(c.apiURL, "/stats/email_bounces", c.apiKey, c.debug, "email_bounces")
	if err != nil {
		return
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
