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
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/raspbeguy/smtp2go_exporter/internal"
)

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
	emailCycle := internal.NewEmailCycleCollector(*apiURL, *apiKey, *debug)
	emailBounces := internal.NewEmailBouncesCollector(*apiURL, *apiKey, *debug)
	emailHistory := internal.NewEmailHistoryCollector(*apiURL, *apiKey, *debug)
	emailSpam := internal.NewEmailSpamCollector(*apiURL, *apiKey, *debug)
	emailUnsubs := internal.NewEmailUnsubsCollector(*apiURL, *apiKey, *debug)

	prometheus.MustRegister(emailCycle)
	prometheus.MustRegister(emailBounces)
	prometheus.MustRegister(emailHistory)
	prometheus.MustRegister(emailSpam)
	prometheus.MustRegister(emailUnsubs)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting exporter on %s...\n", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
