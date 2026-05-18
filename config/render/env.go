package render

import (
	"fmt"
	"os"
	"regexp"
)

var envKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type EnvResolver struct{}

func (EnvResolver) Source() string {
	return "env"
}

func (EnvResolver) Resolve(key string) (string, bool, error) {
	if !envKeyPattern.MatchString(key) {
		return "", false, fmt.Errorf("invalid env key %q", key)
	}
	value, ok := os.LookupEnv(key)
	return value, ok, nil
}
