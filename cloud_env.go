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
	"encoding/json"
	"fmt"
	"github.com/oliveagle/jsonpath"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"strings"
)

const PREFIX_PATTERN_CF = "cloudfoundry"
const PREFIX_PATTERN_ENV = "env"
const PREFIX_PATTERN_FILE = "file"
const PREFIX_PATTERN_USER = "user-provided"

var loadedMappings = make(map[string]interface{})

func Initialize(mappingsFilePath string) string {
	json, err := ioutil.ReadFile(mappingsFilePath)
	if err != nil {
		log.Error(err)
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Error(err)
	}
	mappingsFilePath = dir + mappingsFilePath
	result := gjson.Parse(string(json))
	version := result.Get("version").Int()
	result.ForEach(func(key, value gjson.Result) bool {
		if !result.Get("version").Exists() {
			processMapping(key.String(), value)
		} else if version == 1 {
			processMapping(key.String(), value)
		} else if version == 2 {
			processMappingV2(key.String(), value)
		}
		return true
	})
	return mappingsFilePath
}

func processMapping(mappingName string, config gjson.Result) bool {
	searchPatterns := config.Get("searchPatterns")
	if !searchPatterns.Exists() || len(searchPatterns.Array()) == 0 {
		log.Warnln("No searchPatterns found for mapping", mappingName)
		return false
	}
	searchPatterns.ForEach(func(_, searchPattern gjson.Result) bool {
		value, OK := processSearchPattern(mappingName, searchPattern.String())
		if OK {
			loadedMappings[mappingName] = value
			return false
		} else {
			return true
		}

	})

	return true
}

func processMappingV2(mappingName string, config gjson.Result) {
	config.ForEach(func(key, value gjson.Result) bool {
		searchPatterns := value.Get("searchPatterns")
		if !searchPatterns.Exists() || len(searchPatterns.Array()) == 0 {
			log.Warningln("No credentials found uusing searchPatterns under ", mappingName)
		}

		searchPatterns.ForEach(func(_, searchPattern gjson.Result) bool {
			value, ok := processSearchPattern(fmt.Sprintf("$%s[%s]", mappingName, key.String()), searchPattern.String())
			if ok {
				_, exists := loadedMappings[mappingName]
				if !exists {
					loadedMappings[mappingName] = make(map[string]string)
				}

				loadedMappings[mappingName].(map[string]string)[key.String()] = value
				return false
			} else {
				return true
			}
		})

		return true
	})
}

func processSearchPattern(mappingName string, searchPattern string) (string, bool) {
	patternComponents := strings.Split(searchPattern, ":")
	value := ""
	OK := false
	switch patternComponents[0] {
	case PREFIX_PATTERN_FILE:
		value, OK = processFileSearchPattern(patternComponents)
	case PREFIX_PATTERN_CF:
		value, OK = processCFSearchPattern(patternComponents)
	case PREFIX_PATTERN_ENV:
		value, OK = processEnvSearchPattern(patternComponents)
	case PREFIX_PATTERN_USER:
		value, OK = processUserProvidedSearchPattern(patternComponents)
	default:
		log.Warnln("Unknown searchPattern prefix", patternComponents[0], "Supported prefixes: user-provided, cloudfoundry, env, file")
		return "", false
	}
	if !OK {
		return "", false
	}
	return value, true
}

func processFileSearchPattern(patternComponents []string) (string, bool) {
	filePath, _ := os.Getwd()
	if _, err := os.Stat(filePath); err != nil {
		log.Errorln("File does not exist", filePath)
		return "", false
	} else {
		fullPathName := filePath + patternComponents[1]
		json, err := ioutil.ReadFile(fullPathName)
		if err != nil {
			log.Errorln(err)
			return "", false
		}
		if len(patternComponents) == 3 {
			return processJSONPath(string(json), patternComponents[2])
		} else {
			return string(json), true
		}
	}
}

func processCFSearchPattern(patternComponents []string) (string, bool) {
	vcapServicesString, ok_service := os.LookupEnv("VCAP_SERVICES")
	vcapApplicationString, ok_app := os.LookupEnv("VCAP_APPLICATION")
	if !ok_service && !ok_app {
		return "", false
	} else {
		if patternComponents[1][0] == '$' {
			value, OK := processJSONPath(vcapServicesString, patternComponents[1])
			if OK {
				return value, true
			} else {
				return processJSONPath(vcapApplicationString, patternComponents[1])
			}
		} else {
			// patternComponents[1] is a service instance name, find it in VCAP_SERVICES and return credentials object
			json := gjson.Parse(vcapServicesString)
			res, ok := "", false
			json.ForEach(func(k, v gjson.Result) bool {
				v.ForEach(func(_, item gjson.Result) bool {
					if item.Get("name").String() == patternComponents[1] {
						res, ok = item.Get("credentials").String(), true
						return false
					} else {
						return true
					}
				})
				if ok {
					return false
				} else {
					return true
				}
			})
			return res, ok

		}
	}
}

func processEnvSearchPattern(patternComponents []string) (string, bool) {
	value, OK := os.LookupEnv(patternComponents[1])
	if OK && (len(patternComponents) == 3) {
		return processJSONPath(value, patternComponents[2])
	}
	return value, OK
}

func processUserProvidedSearchPattern(patternComponents []string) (string, bool) {
	vcapServicesString, ok := os.LookupEnv("VCAP_SERVICES")
	if !ok {
		return "", false
	}
	if len(patternComponents) == 3 {
		serviceName := patternComponents[1]
		return processJSONCredentials(vcapServicesString, serviceName, patternComponents[2])
	}
	return "", false
}

func processJSONCredentials(jsonString, servicename, credkey string) (string, bool) {
	if !gjson.Valid(jsonString) {
		log.Errorln("Failed to apply JSONPath", jsonString)
		return "", false
	}
	jsonObj := gjson.Parse(jsonString)
	credArray := jsonObj.Get(PREFIX_PATTERN_USER)
	ret, ok := "", false
	credArray.ForEach(func(_, searchPattern gjson.Result) bool {
		if searchPattern.Get("name").String() == servicename {
			ret, ok = deepSearch(searchPattern, credkey)
			return !ok
		}
		return true
	})
	return ret, ok
}

func deepSearch(current gjson.Result, search string) (string, bool) {

	res := current.Get(search).String()
	ok := false
	if current.Get(search).Exists() {
		return res, true
	}

	_, isMap := current.Value().(map[string]interface{})
	_, isArr := current.Value().([]interface{})
	if isMap || isArr {
		current.ForEach(func(_, v gjson.Result) bool {
			res, ok = deepSearch(v, search)
			return !ok
		})
	}
	return res, ok
}

func processJSONPath(jsonString string, jsonPath string) (string, bool) {
	var json_data interface{}
	err := json.Unmarshal([]byte(jsonString), &json_data)
	if err != nil {
		return "", false
	}
	res, err := jsonpath.JsonPathLookup(json_data, jsonPath)
	_, isMap := res.(map[string]interface{})
	_, isArr := res.([]interface{})

	if isMap || isArr {
		test, _ := json.Marshal(res)
		return string(test), err == nil
	}
	return fmt.Sprintf("%v", res), err == nil
}

func GetCredentialsForService(serviceTag, serviceLabel, credentials string) map[string]string {
	creds_json := gjson.Parse(credentials)
	creds := make(map[string]string)
	key := serviceTag + "_" + serviceLabel + "_"
	if creds_json.Exists() {
		creds_json.ForEach(func(k, v gjson.Result) bool {
			if strings.Index(k.String(), key) == 0 {
				credKey := k.String()[len(key):]
				if credKey == "apikey" && serviceTag == "watson" {
					creds["iam_"+credKey] = v.String()
				} else {
					creds[credKey] = v.String()
				}
			}
			return true
		})
	}
	return creds
}

func GetString(name string) (string, bool) {
	val, ok := loadedMappings[name]
	if !ok {
		return "", false
	}

	_, isStr := val.(string)
	if !isStr {
		bytes, _ := json.Marshal(val)
		return string(bytes), true
	}
	return val.(string), true
}

func GetDictionary(name string) gjson.Result {
	value, ok := GetString(name)
	if !ok {
		log.Warnln(value + " does not exist")
		return gjson.Parse("{\"value\": \"" + value + "\"}")
	} else {
		if gjson.Valid(value) {
			return gjson.Parse(value)
		} else {
			//value is not a valid json object, return object
			return gjson.Parse("{\"value\": \"" + value + "\"}")
		}
	}
}
