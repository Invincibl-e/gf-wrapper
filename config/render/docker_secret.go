package render

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const defaultDockerSecretDir = "/run/secrets"

var dockerSecretNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

type DockerSecretResolver struct {
	root string
}

func NewDockerSecretResolver(root string) DockerSecretResolver {
	if root == "" {
		root = defaultDockerSecretDir
	}
	return DockerSecretResolver{root: root}
}

func (r DockerSecretResolver) Source() string {
	return "docker-secret"
}

func (r DockerSecretResolver) Resolve(key string) (string, bool, error) {
	if !dockerSecretNamePattern.MatchString(key) || strings.Contains(key, "..") || strings.ContainsAny(key, `/\\`) {
		return "", false, fmt.Errorf("invalid docker secret name %q", key)
	}
	return FileResolver{}.Resolve(filepath.Join(r.root, key))
}
