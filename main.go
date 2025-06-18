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
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func doPostRequest(apiURL, endpoint, apiKey string, debug bool, logPrefix string) ([]byte, error) {
	fullURL := apiURL + endpoint

	reqBody, _ := json.Marshal(map[string]string{"api_key": apiKey})
	req, _ := http.NewRequest("POST", fullURL, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[%s] HTTP request failed: %v", logPrefix, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if debug {
		log.Printf("[%s] Raw response: %s\n", logPrefix, string(body))
	}

	return body, nil
}

// -----------------------------
// EmailCycleCollector
// -----------------------------

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

// -----------------------------
// EmailHistoryCollector
// -----------------------------

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

// -----------------------------
// EmailSpamCollector
// -----------------------------

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

// -----------------------------
// EmailUnsubsCollector
// -----------------------------

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
	emailHistory := NewEmailHistoryCollector(*apiURL, *apiKey, *debug)
	emailSpam := NewEmailSpamCollector(*apiURL, *apiKey, *debug)
	emailUnsubs := NewEmailUnsubsCollector(*apiURL, *apiKey, *debug)

	prometheus.MustRegister(emailCycle)
	prometheus.MustRegister(emailBounces)
	prometheus.MustRegister(emailHistory)
	prometheus.MustRegister(emailSpam)
	prometheus.MustRegister(emailUnsubs)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting exporter on %s...\n", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
