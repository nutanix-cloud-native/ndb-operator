/*
Copyright 2021-2022 Nutanix, Inc.

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

package ndbclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
)

type NDBClient struct {
	username string
	password string
	url      string
	client   *http.Client
}

func NewNDBClient(username, password, url string) *NDBClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	return &NDBClient{username, password, url, client}
}

func (ndbClient *NDBClient) Get(path string) (*http.Response, error) {
	url := ndbClient.url + "/" + path
	req, err := http.NewRequest("GET", url, nil)
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
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
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
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(payload))
	if err != nil {
		// fmt.Println(err)
		return nil, err
	}
	req.SetBasicAuth(ndbClient.username, ndbClient.password)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	return ndbClient.client.Do(req)
}
