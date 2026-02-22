package preview

import (
	"fmt"
	"os"
	"path/filepath"

	"docker-wizard/internal/generator/write"
)

type FileStatus string

const (
	FileStatusNew       FileStatus = "new"
	FileStatusSame      FileStatus = "same"
	FileStatusDifferent FileStatus = "different"
	FileStatusExists    FileStatus = "exists"
)

type FilePreview struct {
	Path    string
	Status  FileStatus
	Content string
}

type Preview struct {
	Compose      FilePreview
	Dockerfile   FilePreview
	Dockerignore FilePreview
}

func PreviewFiles(root string, compose string, dockerfile string) (Preview, error) {
	if root == "" {
		return Preview{}, fmt.Errorf("root directory is required")
	}

	composePath := filepath.Join(root, write.ComposeFileName)
	dockerfilePath := filepath.Join(root, write.DockerfileFileName)
	dockerignorePath := filepath.Join(root, write.DockerignoreFileName)

	composePreview, err := buildFilePreview(composePath, compose)
	if err != nil {
		return Preview{}, err
	}
	dockerfilePreview, err := buildFilePreview(dockerfilePath, dockerfile)
	if err != nil {
		return Preview{}, err
	}

	dockerignorePreview := FilePreview{Path: dockerignorePath, Status: FileStatusExists}
	if !fileExists(dockerignorePath) {
		dockerignorePreview = FilePreview{
			Path:    dockerignorePath,
			Status:  FileStatusNew,
			Content: write.DefaultDockerignore(),
		}
	}

	return Preview{
		Compose:      composePreview,
		Dockerfile:   dockerfilePreview,
		Dockerignore: dockerignorePreview,
	}, nil
}

func buildFilePreview(path string, content string) (FilePreview, error) {
	if !fileExists(path) {
		return FilePreview{Path: path, Status: FileStatusNew, Content: content}, nil
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		return FilePreview{}, fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}

	status := FileStatusDifferent
	if string(existing) == content {
		status = FileStatusSame
	}

	return FilePreview{Path: path, Status: status, Content: content}, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
