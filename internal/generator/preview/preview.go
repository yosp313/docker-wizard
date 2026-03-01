package preview

import (
	"fmt"
	"os"
	"path/filepath"

	"docker-wizard/internal/generator/write"
	"docker-wizard/internal/utils"
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

	composePreview, err := buildFilePreview(composePath, compose, write.MergeCompose)
	if err != nil {
		return Preview{}, err
	}
	dockerfilePreview, err := buildFilePreview(dockerfilePath, dockerfile, write.MergeDockerfile)
	if err != nil {
		return Preview{}, err
	}

	dockerignorePreview := FilePreview{Path: dockerignorePath, Status: FileStatusExists}
	if !utils.FileExists(dockerignorePath) {
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

type mergeFunc func(existing string, generated string) (string, error)

func buildFilePreview(path string, content string, merge mergeFunc) (FilePreview, error) {
	if !utils.FileExists(path) {
		return FilePreview{Path: path, Status: FileStatusNew, Content: content}, nil
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		return FilePreview{}, fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}

	targetContent := content
	if merge != nil {
		merged, mergeErr := merge(string(existing), content)
		if mergeErr != nil {
			return FilePreview{}, fmt.Errorf("merge %s: %w", filepath.Base(path), mergeErr)
		}
		targetContent = merged
	}

	status := FileStatusDifferent
	if string(existing) == targetContent {
		status = FileStatusSame
	}

	return FilePreview{Path: path, Status: status, Content: targetContent}, nil
}

// PreviewComposeFile previews only the compose file merge result, without
// touching Dockerfile or .dockerignore. Used by the add subcommand.
func PreviewComposeFile(root string, compose string) (FilePreview, error) {
	if root == "" {
		return FilePreview{}, fmt.Errorf("root directory is required")
	}
	composePath := filepath.Join(root, write.ComposeFileName)
	return buildFilePreview(composePath, compose, write.MergeCompose)
}
