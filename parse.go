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
  result.ForEach(func (key, value gjson.Result) bool{
  	processMapping(key.String(),value) 
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
	//never will be hit 
	return true
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



