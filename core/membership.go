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
	"bytes"
	"encoding/json"
	"net/http"
)

func Login(pubKey, email, password string) (string, error) {
	var body = new(struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	})

	body.Email = email
	body.Password = password

	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	var token string
	if err := request(pubKey, "", http.MethodPatch, "/login", bytes.NewReader(b), &token); err != nil {
		return "", err
	}
	return token, nil
}

func SetRole(pubKey, token, email string, role int) error {
	var data = new(struct {
		Email string `json:"email"`
		Role  int    `json:"role"`
	})
	data.Email = email
	data.Role = role

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return request(pubKey, token, http.MethodPost, "/setrole", bytes.NewReader(b), nil)
}
