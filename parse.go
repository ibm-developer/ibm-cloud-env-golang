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

import ("os"
		"fmt"
		"strings"
  	"github.com/tidwall/gjson"
  	"io/ioutil"
 		"github.com/oliveagle/jsonpath"
 		"encoding/json"
 		log "github.com/sirupsen/logrus"
)

const PREFIX_PATTERN_CF = "cloudfoundry"
const PREFIX_PATTERN_ENV = "env"
const PREFIX_PATTERN_FILE = "file"
const PREFIX_PATTERN_USER = "user-provided"

var loadedMappings = make (map [string]string) 

func check(e error) {
    if e != nil {
    	log.Error(e)
    }
}

func strip_leading_slash(path string) string {
	if (len(path) > 0) && (path[0] == '/' || path[0] == '\\'){
		return path[1:]
	}    
    return path
}
    
func Initialize(mappingsFilePath string) string {
	json, err := ioutil.ReadFile(mappingsFilePath)
  if err != nil {
  	log.Error(err)
  }
	dir, err := os.Getwd()
	if err != nil {
  	log.Error(err)
	}
  mappingsFilePath =  dir + mappingsFilePath;
  result := gjson.Parse(string(json))
  version := result.Get("version").Int()
  result.ForEach(func (key, value gjson.Result) bool{
  	if !result.Get("version").Exists() {
	  	processMapping(key.String(),value) 
	  } else if version == 1 {
	  	processMapping(key.String(),value) 
	  } else if version == 2{
	  	processMappingV2(key.String(),value) 
	  }
  	return true 
  })
	return mappingsFilePath
}

func processMapping(mappingName string, config gjson.Result) bool {
	searchPatterns := config.Get("searchPatterns")
	if !searchPatterns.Exists() || len (searchPatterns.Array()) == 0{
		log.Warnln("No searchPatterns found for mapping", mappingName)
		return false
	}
	searchPatterns.ForEach (func (_, searchPattern gjson.Result) bool {
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

func processMappingV2(mappingName string, config gjson.Result){
	config.ForEach(func (key, value gjson.Result) bool {
		searchPatterns := value.Get("searchPatterns")
		if !searchPatterns.Exists() || len(searchPatterns.Array()) == 0{
			log.Warningln("No credentials found uusing searchPatterns under ", mappingName)
		}

		searchPatterns.ForEach(func(_, searchPattern gjson.Result) bool{
			value, ok := processSearchPattern(fmt.Sprintf("$%s[%s]", mappingName, key.String()), searchPattern.String())
			if ok {
				_, exists := loadedMappings[mappingName]
				jsonAddition := "\""+key.String()+"\" : \"" + value + "\""
				if !exists {
					loadedMappings[mappingName] = "{"+jsonAddition+"}"
				}else{
					jsonStr := loadedMappings[mappingName]
					loadedMappings[mappingName] = jsonStr[:len(jsonStr)-1]+", "+jsonAddition+"}"
				}
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
	value :=""
	OK := false
	switch patternComponents[0]{
	case PREFIX_PATTERN_FILE:
		value, OK = processFileSearchPattern(patternComponents)
		break
	case PREFIX_PATTERN_CF:
		value, OK  = processCFSearchPattern(patternComponents)
		break
	case PREFIX_PATTERN_ENV:
		value, OK  = processEnvSearchPattern(patternComponents)
		break
	case PREFIX_PATTERN_USER:
		value, OK  = processUserProvidedSearchPattern(patternComponents)
		break
	default:
		log.Warnln("Unknown searchPattern prefix", patternComponents[0], "Supported prefixes: cloudfoundry, env, file")
		return "", false
	}
	if !OK {
		return "", false
	}
	return value, true
}


func processFileSearchPattern(patternComponents [] string) (string, bool) {
	filePath, _ := os.Getwd()
	if _, err := os.Stat(filePath); err != nil {
		log.Errorln("File does not exist", filePath)
		return "", false
	} else {
		fullPathName := filePath + patternComponents [1]
		json, err := ioutil.ReadFile(fullPathName)
		if err !=nil{
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

func processCFSearchPattern(patternComponents [] string) (string, bool) {
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
			jsonPath := "$..[?(@.name==\"" + patternComponents[1] + "\")].credentials"
			return processJSONPath(vcapServicesString, jsonPath)
		
		}
	}
}

func processEnvSearchPattern(patternComponents [] string) (string, bool) {
	value, OK := os.LookupEnv(patternComponents[1])
	if OK && (len (patternComponents) == 3) {
		return processJSONPath(value, patternComponents[2])	
	}
	return value, OK
}

func processUserProvidedSearchPattern(patternComponents []string) (string, bool){
	vcapServicesString, ok := os.LookupEnv("VCAP_SERVICES")
	if !ok {
		return "", false
	}
	if(len(patternComponents) == 3){
		serviceName := patternComponents[1]
		return processJSONCredentials(vcapServicesString, serviceName, patternComponents[2])
	}
	return "", false
}

func processJSONCredentials(jsonString, servicename, credkey string) (string, bool){
	if !gjson.Valid(jsonString) {
		log.Errorln("Failed to apply JSONPath", jsonString)
		return "", false
	}
	jsonObj := gjson.Parse(jsonString)
	credArray := jsonObj.Get(PREFIX_PATTERN_USER)
	ret, ok := "", false
	credArray.ForEach (func (_, searchPattern gjson.Result) bool {
		if searchPattern.Get("name").String() == servicename {
			path := "$.." + credkey
			res, err := jsonpath.JsonPathLookup(searchPattern.String(), path)
			_, isMap := res.(map[string]interface{})
			_, isArr := res.([]interface{})
			
			if isMap || isArr {
				bytes, _ := json.Marshal(res)
				res = string(bytes)
				ok = true
			} else {
				ret, ok = fmt.Sprintf("%v", res), err == nil
			}
			return false
		}
		return true
	})
	return ret, ok
}

func processJSONPath(jsonString string, jsonPath string) (string,bool) {
	var json_data interface{}
	json.Unmarshal([]byte(jsonString), &json_data)
	res, err := jsonpath.JsonPathLookup(json_data, jsonPath)
	_, isMap := res.(map[string]interface{})
	_, isArr := res.([]interface{})
	
	if isMap || isArr {
		test, _ := json.Marshal(res)
		return string(test), err == nil
	}
	return fmt.Sprintf("%v", res), err == nil
}

func GetCredentialsForService(serviceTag, serviceLabel string, credentials gjson.Result) map[string]string{
	creds := make(map[string]string)
	key := serviceTag + "_" + serviceLabel + "_"
	if credentials.Exists() {
		credentials.ForEach(func (k, v gjson.Result) bool {
			if strings.Index(k.String(), key) == 0 {
				credKey := k.String()[:len(key)]
				if(credKey == "apikey" && serviceTag == "watson"){
					creds["iam_"+credKey] = v.String()
				}else{
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
	return val, ok
}

func GetDictionary(name string) gjson.Result {
	value, ok := GetString (name)
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



