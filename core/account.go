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
	"fmt"
	"net/http"
	"net/url"
)

// NewAccount initiates the account creation process.
func NewAccount(email string) (string, error) {
	var stripeURL string

	path := fmt.Sprintf("/account/init?email=%s", url.QueryEscape(email))
	if err := request("", "", http.MethodGet, path, nil, &stripeURL); err != nil {
		return "", err
	}
	return stripeURL, nil
}
