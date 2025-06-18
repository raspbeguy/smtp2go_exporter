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
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
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
