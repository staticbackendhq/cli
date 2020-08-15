/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package core

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const baseURL = "http://localhost:8099"

func request(pubKey, token, method, path string, body io.Reader, v interface{}) error {
	req, err := http.NewRequest(method, baseURL+path, body)
	if err != nil {
		return err
	}

	req.Header.Set("SB-PUBLIC-KEY", pubKey)

	if len(token) > 0 {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error returned by API: %s: %s", res.Status, string(body))
	}

	if v == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(v)
}
