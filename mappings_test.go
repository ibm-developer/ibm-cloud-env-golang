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
	"os"
	"github.com/tidwall/gjson"
)

const jsonObject = 
`{
	"service1": [
		{
			"name": "service1-name1",
			"credentials": {
				"username": "service1-username1"
			}
		},
		{
			"name": "service1-name2",
			"credentials": {
				"username": "service1-username2"
			}
		}
	],
	"user-provided": [
		{
			"name": "service2-name1",
			"credentials":{
				"username": "service2-username1"
			}
		}
	]
}` 

const vcap_application = `{"application_name": "test-application"}`
const var_string = `test-12345` 
const credentials = `{"credentials": {
		"username": "env-var-json-username"
	}}`

func setEnvVariable() {
	os.Setenv("VCAP_APPLICATION", vcap_application)
	os.Setenv("VCAP_SERVICES", jsonObject)
	os.Setenv("ENV_VAR_STRING", var_string)
	os.Setenv("ENV_VAR_JSON", credentials)

	Initialize("/invalid-file-name")
	Initialize("server/config/mappings.json")
}


func TestPlainTextFile(t *testing.T) {
	setEnvVariable()
	testString, _ := GetString("file_var1")
	if testString != "plain-text-string" {
		t.Errorf("testString is: " +  testString)
		t.Errorf("can't read " + testString + " from GetString()")
	} 
	result := GetDictionary("file_var1").Get("value")
	if result.String() != "plain-text-string" {
		 t.Errorf("can't read " + result.String() + " text from GetDictionary()")
	} 
}

func TestJsonFileAndPath(t *testing.T){
	setEnvVariable()
	testString, _ := GetString("file_var2")
	if testString != gjson.Parse("{\"level2\":12345}").String() {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "{\"level2\":12345}")
	}

	testString = GetDictionary("file_var2").Get("level2").String()
	if testString != "12345" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "12345")
	}
}

func TestReadVcapsWithJsonPath(t *testing.T){
	setEnvVariable()
	testString, _ := GetString("cf_var2")

	if testString != "service1-username1" {
		t.Errorf("can't read " + testString + " from GetString()")
	}

	result := GetDictionary("cf_var2").Get("value")
	if result.String() != "service1-username1" {
		 t.Errorf("can't read " + result.String() + " text from GetDictionary()")
	} 

	testString2, _ := GetString("cf_var3")
	if testString2 != "test-application" {
		t.Errorf("can't read " + testString2 + " from GetString()")
	}
	result2 := GetDictionary("cf_var3").Get("value")
	if result2.String() != "test-application" {
		 t.Errorf("can't read " + result2.String() + " text from GetDictionary()")
	} 

	//removed this test because there wasn't a cf_var4 in mappings.json
	/*testString3, _ := GetString("cf_var4")
	if testString3 != "service1-username1" {
		t.Errorf("can't read " + testString3 + " from GetString()")
	}
	result3 := GetDictionary("cf_var4").Get("value")
	if result3.String() != "service1-username1" {
		 t.Errorf("can't read " + result3.String() + " text from GetDictionary()")
	}*/
}

func TestSimpleStringFromEnvVar(t *testing.T){
	setEnvVariable()
	testString, _ := GetString("env_var1")
	if testString != "test-12345" {
		t.Errorf("can't read " + testString + " from GetString()")
	}
	result := GetDictionary("env_var1").Get("value")
	if result.String() != "test-12345" {
		 t.Errorf("can't read " + result.String() + " text from GetDictionary()")
	} 
}

func TestJsonFromEnvVar(t *testing.T){
	setEnvVariable()
	testString, _ := GetString("env_var2")
	if testString != credentials {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, credentials)
	}

	testString = GetDictionary("env_var2").Get("credentials").Get("username").String()
	if testString != "env-var-json-username" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "env-var-json-username")
	}
}

func TestJsonPathFromEnvVar(t *testing.T){
	setEnvVariable()
	testString, _ := GetString("env_var3")
	if testString != "env-var-json-username" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "env-var-json-username")
	}

	testString = GetDictionary("env_var3").Get("value").String()
	if testString != "env-var-json-username" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "env-var-json-username")
	}
}