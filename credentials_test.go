/*
 * © Copyright IBM Corp. 2018
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package IBMCloudEnv

import (
	"testing"
)

const service_credentials = `{
	"tag_label_creds": "someOtherCreds",
	"watson_discovery_password": "password",
	"watson_conversation_password": "password",
	"watson_conversation_url": "url",
	"watson_conversation_username": "username",
	"watson_conversation_api_key": "api_key",
	"watson_conversation_apikey": "apikey"
}`

var filtered_credentials map[string]string = map[string]string{
	"api_key":    "api_key",
	"iam_apikey": "apikey",
	"password":   "password",
	"url":        "url",
	"username":   "username",
}

func TestMissingParams(t *testing.T) {
	if len(GetCredentialsForService("", "", "")) != 0 {
		t.Errorf("Credentials not empty\n")
	}
	if len(GetCredentialsForService("", "", "{}")) != 0 {
		t.Errorf("Credentials not empty\n")
	}
}

func TestCredentials(t *testing.T) {
	returned_creds := GetCredentialsForService("watson", "conversation", service_credentials)
	if len(returned_creds) != len(filtered_credentials) {
		t.Errorf("Credentials have different sizes: %d, %d\n", len(returned_creds), len(filtered_credentials))
	}

	for k, v := range returned_creds {
		if filtered_credentials[k] != v {
			t.Errorf("Credentials have different values for key %s: %s, %s", k, filtered_credentials[k], v)
		}
	}
}
