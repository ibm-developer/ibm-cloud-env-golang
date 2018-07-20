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
	"fmt"
	"github.com/tidwall/gjson"
)

const jsonObjectV1 = 
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
			"credentials":{
				"apikey": "apikey1"
			},
			"name":"servicename1"
			
		},
		{
			"credentials":{
				"writer":{
					"apikey": "apikey2"
				}
			},
			"name":"servicename2"
		},
		{
			"credentials":{
				"apikey": "apikey3", 
				"nestedCreds": {
					"nestedKey1": "nestedValue1",
					"nestedKey2": {
						"nestedKey3": "nestedValue3"
					}
				}
			},
			"name":"servicename3"
		}
	]
}` 

const vcap_applicationV1 = `{"application_name": "test-application"}`
const var_stringV1 = `test-12345` 
const credentialsV1 = `{"credentials": {
		"username": "env-var-json-username"
	}}`

func setEnvVariableV1() {
	os.Setenv("VCAP_APPLICATION", vcap_applicationV1)
	os.Setenv("VCAP_SERVICES", jsonObjectV1)
	os.Setenv("ENV_VAR_STRING", var_stringV1)
	os.Setenv("ENV_VAR_JSON", credentialsV1)

	Initialize("/invalid-file-name")
	Initialize("server/config/v1/mappings.json")
}


func TestPlainTextFileV1(t *testing.T) {
	setEnvVariableV1()
	testString, _ := GetString("file_var1")
	if testString != "plain-text-string" {
		t.Errorf("testString is: " +  testString)
		t.Errorf("can't read " + testString + " from GetString()")
	} 

	fmt.Println("IN V1 result var1 is: ")
	fmt.Println(GetDictionary("file_var1"))

	result := GetDictionary("file_var1").Get("value")
	if result.String() != "plain-text-string" {
		 t.Errorf("can't read " + result.String() + " text from GetDictionary()")
	} 
}

func TestJsonFileAndPathV1(t *testing.T){
	setEnvVariableV1()
	testString, _ := GetString("file_var2")
	if testString != gjson.Parse("{\"level2\":12345}").String() {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "{\"level2\":12345}")
	}

	testString = GetDictionary("file_var2").Get("level2").String()
	if testString != "12345" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "12345")
	}
}

func TestCFServiceCredentialsWithInstanceNameV1 (t *testing.T){
	setEnvVariableV1()
	// testString, _ := GetString("cf_var1")
	// if testString != "{\n\t\"username\": \"service1-username1\"\n}" {
	// 	t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "{\n\t\"username\": \"service1-username1\"\n}")
	// }

	testString := GetDictionary("cf_var1").Get("username").String()
	if testString != "service1-username1" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "{\"username\":\"service1-username1\"}")
	}
}

func TestUserCredentialsWithVcapServicesV1(t *testing.T){
	setEnvVariableV1()
	testString, _ := GetString("user_provided_var1")
	if testString != "apikey1"{
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "apikey1")
	}

	testString2, _ := GetString("user_provided_var2")
	if testString2 != "apikey2"{
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString2, "apikey2")
	}

	testString3, _ := GetString("user_provided_nested1")
	if testString3 != "nestedValue1"{
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString3, "nestedValue1")
	}

	testString4, _ := GetString("user_provided_nested2")
	if testString4 != "nestedValue3"{
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString4, "nestedValue3")
	}
}

func TestReadVcapsWithJsonPathV1(t *testing.T){
	setEnvVariableV1()
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

func TestSimpleStringFromEnvVarV1(t *testing.T){
	setEnvVariableV1()
	testString, _ := GetString("env_var1")
	if testString != "test-12345" {
		t.Errorf("can't read " + testString + " from GetString()")
	}
	result := GetDictionary("env_var1").Get("value")
	if result.String() != "test-12345" {
		 t.Errorf("can't read " + result.String() + " text from GetDictionary()")
	} 
}

func TestJsonFromEnvVarV1(t *testing.T){
	setEnvVariableV1()
	testString, _ := GetString("env_var2")
	if testString != credentials {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, credentials)
	}

	testString = GetDictionary("env_var2").Get("credentials").Get("username").String()
	if testString != "env-var-json-username" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "env-var-json-username")
	}
}

func TestJsonPathFromEnvVarV1(t *testing.T){
	setEnvVariableV1()
	testString, _ := GetString("env_var3")
	if testString != "env-var-json-username" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "env-var-json-username")
	}

	testString = GetDictionary("env_var3").Get("value").String()
	if testString != "env-var-json-username" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "env-var-json-username")
	}
}