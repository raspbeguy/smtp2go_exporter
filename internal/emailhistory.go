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

	"github.com/prometheus/client_golang/prometheus"
)

type EmailHistoryEntry struct {
	Used         float64 `json:"used"`
	ByteCount    float64 `json:"bytecount"`
	AvgSize      float64 `json:"avgsize"`
	EmailAddress string  `json:"email_address"`
	Bounces      float64 `json:"bounces"`
	Clicks       float64 `json:"clicks"`
	Opens        float64 `json:"opens"`
	Rejects      float64 `json:"rejects"`
	Spam         float64 `json:"spam"`
	Unsubscribes float64 `json:"unsubscribes"`
}

type EmailHistoryResponse struct {
	RequestID string `json:"request_id"`
	Data      struct {
		History []EmailHistoryEntry `json:"history"`
		Count   int                 `json:"count"`
	} `json:"data"`
}

type EmailHistoryCollector struct {
	mutex     sync.Mutex
	apiURL    string
	apiKey    string
	debug     bool
	namespace string

	metrics map[string]*prometheus.GaugeVec
}

func NewEmailHistoryCollector(apiURL, apiKey string, debug bool) *EmailHistoryCollector {
	ns := "smtp2go_email_history"

	labels := []string{"email_address"}

	return &EmailHistoryCollector{
		apiURL:    apiURL,
		apiKey:    apiKey,
		debug:     debug,
		namespace: ns,
		metrics: map[string]*prometheus.GaugeVec{
			"used": prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "used",
				Help:      "Number of emails used per email address",
			}, labels),
			"bytecount": prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "bytecount",
				Help:      "Total size in bytes of emails sent per email address",
			}, labels),
			"avgsize": prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "avgsize",
				Help:      "Average size of emails per email address",
			}, labels),
			"bounces": prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "bounces",
				Help:      "Number of bounces per email address",
			}, labels),
			"clicks": prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "clicks",
				Help:      "Number of clicks per email address",
			}, labels),
			"opens": prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "opens",
				Help:      "Number of opens per email address",
			}, labels),
			"rejects": prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "rejects",
				Help:      "Number of rejected emails per email address",
			}, labels),
			"spam": prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "spam",
				Help:      "Number of spam reports per email address",
			}, labels),
			"unsubscribes": prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: ns,
				Name:      "unsubscribes",
				Help:      "Number of unsubscribes per email address",
			}, labels),
		},
	}
}

func (c *EmailHistoryCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		metric.Describe(ch)
	}
}

func (c *EmailHistoryCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	body, err := doPostRequest(c.apiURL, "/stats/email_history", c.apiKey, c.debug, "email_history")
	if err != nil {
		return
	}

	var apiResp EmailHistoryResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Println("[email_history] Failed to parse JSON:", err)
		return
	}

	// Reset metrics to remove outdated labels
	for _, metric := range c.metrics {
		metric.Reset()
	}

	for _, entry := range apiResp.Data.History {
		labels := prometheus.Labels{"email_address": entry.EmailAddress}
		c.metrics["used"].With(labels).Set(entry.Used)
		c.metrics["bytecount"].With(labels).Set(entry.ByteCount)
		c.metrics["avgsize"].With(labels).Set(entry.AvgSize)
		c.metrics["bounces"].With(labels).Set(entry.Bounces)
		c.metrics["clicks"].With(labels).Set(entry.Clicks)
		c.metrics["opens"].With(labels).Set(entry.Opens)
		c.metrics["rejects"].With(labels).Set(entry.Rejects)
		c.metrics["spam"].With(labels).Set(entry.Spam)
		c.metrics["unsubscribes"].With(labels).Set(entry.Unsubscribes)
	}

	for _, metric := range c.metrics {
		metric.Collect(ch)
	}
}
