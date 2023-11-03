package lib

import (
	"fmt"
	"os"
)

// variableInformation defines a holder for environment variable details
type variableInformation struct {
	required     bool
	defaultValue string
}

// environment stores a map of environment values
var environment = map[string]string{}

// environmentInformation stores a map of the required flag and default value of environment variables
var environmentInformation = map[string]variableInformation{
	clientId:     {required: true},
	clientSecret: {required: true},
	authUrl:      {required: true},
}

// initializeEnvironment initializes the environment map and ensures all required values are set
func initializeEnvironment() error {
	for key, reqType := range environmentInformation {

		// Look up the key, and error out if it's not present
		value, present := os.LookupEnv(key)
		if !present && reqType.required {
			return fmt.Errorf("%s is a required environment variable", key)
		}

		// Store the value if any or the default one otherwise
		if value == "" {
			environment[key] = reqType.defaultValue
		} else {
			environment[key] = value
		}
	}

	return nil
}
