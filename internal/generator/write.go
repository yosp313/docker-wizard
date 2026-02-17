package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ComposeFileName    = "docker-compose.yml"
	DockerfileFileName = "Dockerfile"
)

type Output struct {
	ComposePath    string
	DockerfilePath string
}

func WriteFiles(root string, compose string, dockerfile string) (Output, error) {
	if root == "" {
		return Output{}, fmt.Errorf("root directory is required")
	}

	composePath := filepath.Join(root, ComposeFileName)
	dockerfilePath := filepath.Join(root, DockerfileFileName)

	if fileExists(composePath) {
		return Output{}, fmt.Errorf("%s already exists", ComposeFileName)
	}
	if fileExists(dockerfilePath) {
		return Output{}, fmt.Errorf("%s already exists", DockerfileFileName)
	}

	if err := os.WriteFile(composePath, []byte(compose), 0644); err != nil {
		return Output{}, fmt.Errorf("write compose: %w", err)
	}
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return Output{}, fmt.Errorf("write dockerfile: %w", err)
	}

	return Output{ComposePath: composePath, DockerfilePath: dockerfilePath}, nil
}
