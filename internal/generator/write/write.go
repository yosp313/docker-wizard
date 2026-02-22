package write

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ComposeFileName      = "docker-compose.yml"
	DockerfileFileName   = "Dockerfile"
	DockerignoreFileName = ".dockerignore"
)

type Output struct {
	ComposePath      string
	DockerfilePath   string
	DockerignorePath string
}

func WriteFiles(root string, compose string, dockerfile string) (Output, error) {
	if root == "" {
		return Output{}, fmt.Errorf("root directory is required")
	}

	composePath := filepath.Join(root, ComposeFileName)
	dockerfilePath := filepath.Join(root, DockerfileFileName)
	dockerignorePath := filepath.Join(root, DockerignoreFileName)

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

	dockerignoreCreated := false
	if !fileExists(dockerignorePath) {
		if err := os.WriteFile(dockerignorePath, []byte(DefaultDockerignore()), 0644); err != nil {
			return Output{}, fmt.Errorf("write dockerignore: %w", err)
		}
		dockerignoreCreated = true
	}

	output := Output{ComposePath: composePath, DockerfilePath: dockerfilePath}
	if dockerignoreCreated {
		output.DockerignorePath = dockerignorePath
	}
	return output, nil
}

func DefaultDockerignore() string {
	return "" +
		".git\n" +
		".gitignore\n" +
		"node_modules\n" +
		"vendor\n" +
		"bin\n" +
		"dist\n" +
		"build\n" +
		"tmp\n"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
