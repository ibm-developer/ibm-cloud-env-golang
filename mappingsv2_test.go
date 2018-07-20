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
)

const jsonObjectV2 = 
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

const vcap_applicationV2 = `{"application_name": "test-application"}`
const var_stringV2 = `test-12345` 
const credentialsV2 = `{"credentials": {
		"username": "env-var-json-username"
	}}`

func setEnvVariableV2() {
	os.Setenv("VCAP_APPLICATION", vcap_applicationV2)
	os.Setenv("VCAP_SERVICES", jsonObjectV2)
	os.Setenv("ENV_VAR_STRING", var_stringV2)
	os.Setenv("ENV_VAR_JSON", credentialsV2)

	Initialize("/invalid-file-name")
	Initialize("server/config/v2/mappings.json")
}


func TestPlainTextFileV2(t *testing.T) {
	setEnvVariableV2()
	result := GetDictionary("var1")
	fmt.Println("result var1 is: ")
	fmt.Println(result)

	result2:= result.Get("value")
	fmt.Println("result2 is: ")
	fmt.Println(result2)

	if result2.String() != "plain-text-string" {
		 t.Errorf("can't read " + result2.String() + " text from GetDictionary()")
	} 
}

func TestJsonFileAndPathV2(t *testing.T){
	setEnvVariableV2()
	testString := GetDictionary("var2").Get("level2").String()
	if testString != "12345" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "12345")
	}
}

func TestCFServiceCredentialsWithServiceInstanceV2 (t *testing.T){
	setEnvVariableV2()
	testString := GetDictionary("var1").Get("username").String()
	if testString != "service1-username1" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "service1-username1")
	}
}

func TestReadVcapsWithJsonPathV2(t *testing.T){
	setEnvVariableV2()

	result := GetDictionary("var2").Get("value")
	if result.String() != "service1-username1" {
		 t.Errorf("can't read " + result.String() + " text from GetDictionary()")
	} 

	result2 := GetDictionary("var3").Get("value")
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

func TestSimpleStringFromEnvVarV2(t *testing.T){
	setEnvVariableV2()
	result := GetDictionary("var1").Get("value")
	if result.String() != "test-12345" {
		 t.Errorf("can't read " + result.String() + " text from GetDictionary()")
	} 
}

func TestJsonFromEnvVarV2(t *testing.T){
	setEnvVariableV2()

	testString := GetDictionary("var2").Get("credentials").Get("username").String()
	if testString != "env-var-json-username" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "env-var-json-username")
	}
}

func TestJsonPathFromEnvVarV2(t *testing.T){
	setEnvVariableV2()

	testString := GetDictionary("var3").Get("value").String()
	if testString != "env-var-json-username" {
		t.Errorf("Got: \t%s\n Wanted: \t%s\n", testString, "env-var-json-username")
	}
}