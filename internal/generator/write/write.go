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

	composeTempPath, err := writeTempFile(root, "docker-compose-*.tmp", compose)
	if err != nil {
		return Output{}, fmt.Errorf("write compose temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(composeTempPath)
	}()

	dockerfileTempPath, err := writeTempFile(root, "dockerfile-*.tmp", dockerfile)
	if err != nil {
		return Output{}, fmt.Errorf("write dockerfile temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(dockerfileTempPath)
	}()

	if err := os.Rename(composeTempPath, composePath); err != nil {
		return Output{}, fmt.Errorf("move compose into place: %w", err)
	}
	if err := os.Rename(dockerfileTempPath, dockerfilePath); err != nil {
		rollbackErr := os.Remove(composePath)
		if rollbackErr != nil && !os.IsNotExist(rollbackErr) {
			return Output{}, fmt.Errorf("move dockerfile into place: %w (rollback compose: %v)", err, rollbackErr)
		}
		return Output{}, fmt.Errorf("move dockerfile into place: %w", err)
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

func writeTempFile(root string, pattern string, content string) (string, error) {
	tempFile, err := os.CreateTemp(root, pattern)
	if err != nil {
		return "", err
	}

	path := tempFile.Name()
	if _, err := tempFile.WriteString(content); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := tempFile.Chmod(0644); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}

	return path, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
