/*
Copyright 2022-2023 Nutanix, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ndb_client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
)

// NDBClientInterface defines the methods for an NDB client.
type NDBClientHTTPInterface interface {
	NewRequest(method, endpoint string, requestBody interface{}) (*http.Request, error)
	Do(req *http.Request) (*http.Response, error)
}

type NDBClient struct {
	username string
	password string
	url      string
	client   *http.Client
}

func NewNDBClient(username, password, url, caCert string, skipVerify bool) *NDBClient {
	TLSClientConfig := &tls.Config{InsecureSkipVerify: skipVerify}
	if caCert != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		TLSClientConfig.RootCAs = caCertPool
	}
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: TLSClientConfig},
	}
	return &NDBClient{username, password, url, client}
}

func (ndbClient *NDBClient) Get(path string) (*http.Response, error) {
	url := ndbClient.url + "/" + path
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		// fmt.Println(err)
		return nil, err
	}
	req.SetBasicAuth(ndbClient.username, ndbClient.password)
	return ndbClient.client.Do(req)
}

func (ndbClient *NDBClient) Post(path string, body interface{}) (*http.Response, error) {
	url := ndbClient.url + "/" + path
	payload, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		// fmt.Println(err)
		return nil, err
	}
	req.SetBasicAuth(ndbClient.username, ndbClient.password)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	return ndbClient.client.Do(req)
}

func (ndbClient *NDBClient) Delete(path string, body interface{}) (*http.Response, error) {
	url := ndbClient.url + "/" + path
	payload, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(payload))
	if err != nil {
		// fmt.Println(err)
		return nil, err
	}
	req.SetBasicAuth(ndbClient.username, ndbClient.password)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	return ndbClient.client.Do(req)
}

func (ndbClient *NDBClient) NewRequest(method, endpoint string, requestBody interface{}) (*http.Request, error) {

	url := ndbClient.url + "/" + endpoint

	var body io.Reader

	if requestBody != nil {
		// Serialize the request body to JSON.
		payload, err := json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
		// Create a reader from the serialized payload.
		body = bytes.NewReader(payload)
	}

	// Create a new HTTP request with the specified method and URL.
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.SetBasicAuth(ndbClient.username, ndbClient.password)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	return req, nil
}

func (ndbClient *NDBClient) Do(req *http.Request) (*http.Response, error) {
	// Use the HTTP client to send the provided request.
	return ndbClient.client.Do(req)
}
