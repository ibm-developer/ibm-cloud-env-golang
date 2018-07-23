# IBM Cloud Environment

The `ibm-cloud-env-golang` module allows to abstract environment variables from various Cloud compute providers, such as, but not limited to, CloudFoundry and Kubernetes, so the application could be environment-agnostic.

The module allows to define an array of search patterns that will be executed one by one until required value is found.

### Installation

```bash
go get github.com/ibm-developer/ibm-cloud-env-golang
```
 
### Usage

Create a JSON file containing your mappings and initialize the module

```golang
import "github.com/ibm-developer/ibm-cloud-env-golang"

//in main function 
IBMCloudEnv.init("/path/to/the/mappings/file/relative/to/prject/root")
```
 
#### Supported search patterns types
ibm-cloud-config supports searching for values using three search pattern types - cloudfoundry, env, file. 
- Using `cloudfoundry` allows to search for values in VCAP_SERVICES and VCAP_APPLICATIONS environment variables
- Using `env` allows to search for values in environment variables
- Using `file` allows to search for values in text/json files

#### Example search patterns
- cloudfoundry:service-instance-name - searches through parsed VCAP_SERVICES environment variable and returns the `credentials` object of the matching service instance name
- cloudfoundry:$.JSONPath - searches through parsed VCAP_SERVICES and VCAP_APPLICATION environment variables and returns the value that corresponds to JSONPath
- env:env-var-name - returns environment variable named "env-var-name"
- env:env-var-name:$.JSONPath - attempts to parse the environment variable "env-var-name" and return a value that corresponds to JSONPath
- file:/server/config.text - returns content of /server/config.text file
- file:/server/config.json:$.JSONPath - reads the content of /server/config.json file, tries to parse it, returns the value that corresponds to JSONPath
- user-provided:service:information - attempts to find and return the information (ex. credentials of API key, service name) of the requested variable that the user requests 

#### mappings.json file example
```javascript
{
    "service1-credentials": {
        "searchPatterns": [
            "cloudfoundry:my-service1-instance-name", 
            "env:my-service1-credentials", 
            "file:/localdev/my-service1-credentials.json" 
        ]
    },
    "service2-username": {
        "searchPatterns":[
            "cloudfoundry:$.service2[@.name=='my-service2-instance-name'].credentials.username",
            "env:my-service2-credentials:$.username",
            "file:/localdev/my-service1-credentials.json:$.username" 
        ]
    }
}
```

#### user-provided search pattern example
```javascript
{
    "user_provided_var1":{
        "searchPatterns": [
            "user-provided:servicename1:apikey"
        ]
    }
}
```

### Using the values in application

In your application retrieve the values using below commands

```golang
service1credentials := IBMCloudEnv.getDictionary("service1-credentials") // this will be a dictionary
service2username := IBMCloudEnv.getString("service2-username") // this will be a string
```

Following the above approach your application can be implemented in an runtime-environment agnostic way, abstracting differences in environment variable management introduced by different cloud compute providers.


## Publishing Changes

In order to publish changes, you will need to fork the repository or ask to join the `ibm-developer` org and branch off the `master` branch.

Make sure to follow the [conventional commit specification](https://conventionalcommits.org/) before contributing. To help you with commit a commit template is provide. Run `config.sh` to initialize the commit template to your `.git/config`

Once you are finished with your changes, run `go test` to make sure all tests pass.

Do a pull request against `master`, make sure the build passes. A team member will review and merge your pull request.