package env

import (
	"os"

	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

// LoadEnvFileIfNeeded allows to load .env files locally,
// based on the current environment.
//
// This is a noop in production and test.
func LoadEnvFileIfNeeded(environment string) error {
	var err error

	if environment == "development" {
		testFile := os.ExpandEnv("$GOPATH/src/github.com/deliveroo/rs-devices/.env.dev")
		err = godotenv.Load(testFile)
	}
	if environment == "test" {
		// Workaround for the tests: https://github.com/joho/godotenv/issues/43
		testFile := os.ExpandEnv("$GOPATH/src/github.com/deliveroo/rs-devices/.env.test")
		err = godotenv.Load(testFile)
	}

	if err != nil {
		return errors.Wrap(err, "error loading the dotenv file")
	}

	return nil
}

// GetAppEnv reads the APP_ENV variable
func GetAppEnv() string {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		return "development"
	}
	return appEnv
}

// IsEnabled checks if a env variable is set and is true.
//
func IsEnabled(key string) bool {
	val := os.Getenv(key)
	return strings.ToLower(val) == "true"
}

// Fetch reads an os env variable and returns a fallback
// if the variable is not present.
// Notice that "" is returned for both unset variables _and_
// variables explicitly set to "". In this case, we don't
// really care about the distinction.
func Fetch(key string, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

// Map returns the os environment as a map.
// This is useful to iterate over the env to search for a value.
func Map() map[string]string {
	dict := make(map[string]string)
	rawVars := os.Environ()

	for _, entry := range rawVars {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			dict[parts[0]] = parts[1]
		}
	}

	return dict
}
