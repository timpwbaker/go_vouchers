package env_test

import (
	"os"
	"testing"

	"github.com/deliveroo/rs-devices/pkg/env"
)

func TestIsEnabled(t *testing.T) {
	os.Setenv("TEST_VAR_ENABLED", "true")
	en0 := env.IsEnabled("TEST_VAR_ENABLED")
	if !en0 {
		t.Error("expected the TEST_VAR_ENABLED to be enabled, was disabled")
	}

	os.Setenv("TEST_VAR_DISABLED", "no")
	en1 := env.IsEnabled("TEST_VAR_DISABLED")
	if en1 {
		t.Error("expected the TEST_VAR_DISABLED to be disabled, was enabled")
	}

	en2 := env.IsEnabled("TEST_VAR_MISSING")
	if en2 {
		t.Error("expected the TEST_VAR_MISSING to be disabled, was enabled")
	}

	os.Unsetenv("TEST_VAR_ENABLED")
	os.Unsetenv("TEST_VAR_DISABLED")
}

func TestFetch(t *testing.T) {
	val := env.Fetch("FOOBAR", "default")
	if val != "default" {
		t.Error("expected Fetch() to return fallback for missing variable, but it didn't")
	}

	os.Setenv("FOOBAR", "blabla")
	val2 := env.Fetch("FOOBAR", "default")
	if val2 != "blabla" {
		t.Error("expected Fetch() to return an env var value for a defined var, but it didn't")
	}

	os.Unsetenv("FOOBAR")
}

func TestFetchNested(t *testing.T) {
	os.Setenv("BAR", "bar")

	val := env.Fetch(
		"FOO",
		env.Fetch(
			"BAR",
			env.Fetch("BAZ", "default"),
		))

	if val != "bar" {
		t.Error("expected a nested Fetch() to return the first present value, but it didn't")
	}

	val2 := env.Fetch(
		"FOO",
		env.Fetch(
			"BAR_NOPE",
			env.Fetch("BAZ", "default"),
		))

	if val2 != "default" {
		t.Error("expected a nested Fetch() to return the fallback with no var defined, but it didn't")
	}

	os.Unsetenv("BAR")
}

func TestMap(t *testing.T) {
	os.Setenv("FOO", "foo")
	os.Setenv("BAR", "bar")

	out := env.Map()
	if out["FOO"] != "foo" || out["BAR"] != "bar" {
		t.Error("expected Map() to return a map with the correct env variables")
	}

	os.Unsetenv("FOO")
	os.Unsetenv("BAR")
}

func TestGetAppEnv(t *testing.T) {
	os.Setenv("APP_ENV", "app_environment_present")
	appEnv := env.GetAppEnv()

	if appEnv != "app_environment_present" {
		t.Error("expected GetAppEnv() to equal APP_ENV value")
	}

	os.Setenv("APP_ENV", "")
	appEnv = env.GetAppEnv()

	if appEnv != "development" {
		t.Error("expected GetAppEnv() to equal development")
	}

	os.Unsetenv("APP_ENV")
}

func TestLoadEnvFileIfNeeded(t *testing.T) {
	err := env.LoadEnvFileIfNeeded("test")

	if err != nil {
		t.Error("expected LoadEnvFileIfNeeded('test') to not error")
	}

	if os.Getenv("ENVIRONMENT") != "test" {
		t.Error("expected LoadEnvFileIfNeeded('test') to load test environment")
	}

	os.Unsetenv("ENVIRONMENT")
	err = env.LoadEnvFileIfNeeded("development")

	if err != nil {
		t.Error("expected LoadEnvFileIfNeeded('development') to not error")
	}

	if os.Getenv("ENVIRONMENT") != "development" {
		t.Error("expected LoadEnvFileIfNeeded('development') to load development environment")
	}

	os.Unsetenv("ENVIRONMENT")
	err = env.LoadEnvFileIfNeeded("not an environment")

	if err != nil && err.Error() != "error environment doesn't exist please use 'development' or 'test'" {
		t.Error("expected LoadEnvFileIfNeeded('') to error")
	}

	if os.Getenv("ENVIRONMENT") != "" {
		t.Error("expected LoadEnvFileIfNeeded('') not to load development environment")
	}
}
