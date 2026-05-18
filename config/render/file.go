package render

import (
	"os"
	"strings"
)

type FileResolver struct{}

func (FileResolver) Source() string {
	return "file"
}

func (FileResolver) Resolve(key string) (string, bool, error) {
	content, err := os.ReadFile(key)
	if err != nil {
		return "", false, err
	}
	return strings.TrimRight(string(content), "\r\n"), true, nil
}
