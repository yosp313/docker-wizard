package write

import (
	"fmt"
	"os"
	"path/filepath"

	"docker-wizard/internal/utils"
)

const (
	ComposeFileName      = "docker-compose.yml"
	DockerfileFileName   = "Dockerfile"
	DockerignoreFileName = ".dockerignore"
)

type Output struct {
	ComposePath          string
	ComposeStatus        WriteStatus
	ComposeBackupPath    string
	DockerfilePath       string
	DockerfileStatus     WriteStatus
	DockerfileBackupPath string
	DockerignorePath     string
	DockerignoreStatus   WriteStatus
}

type WriteStatus string

const (
	WriteStatusCreated   WriteStatus = "created"
	WriteStatusUpdated   WriteStatus = "updated"
	WriteStatusUnchanged WriteStatus = "unchanged"
)

func WriteFiles(root string, compose string, dockerfile string) (Output, error) {
	if root == "" {
		return Output{}, fmt.Errorf("root directory is required")
	}

	composePath := filepath.Join(root, ComposeFileName)
	dockerfilePath := filepath.Join(root, DockerfileFileName)
	dockerignorePath := filepath.Join(root, DockerignoreFileName)

	output := Output{
		ComposePath:      composePath,
		DockerfilePath:   dockerfilePath,
		DockerignorePath: dockerignorePath,
	}

	composeStatus, composeBackup, err := writeManagedFile(root, composePath, "docker-compose-*.tmp", compose, MergeCompose)
	if err != nil {
		return Output{}, err
	}
	output.ComposeStatus = composeStatus
	output.ComposeBackupPath = composeBackup

	dockerfileStatus, dockerfileBackup, err := writeManagedFile(root, dockerfilePath, "dockerfile-*.tmp", dockerfile, MergeDockerfile)
	if err != nil {
		return Output{}, err
	}
	output.DockerfileStatus = dockerfileStatus
	output.DockerfileBackupPath = dockerfileBackup

	if !utils.FileExists(dockerignorePath) {
		if err := os.WriteFile(dockerignorePath, []byte(DefaultDockerignore()), 0644); err != nil {
			return Output{}, fmt.Errorf("write dockerignore: %w", err)
		}
		output.DockerignoreStatus = WriteStatusCreated
	} else {
		output.DockerignoreStatus = WriteStatusUnchanged
	}

	return output, nil
}

type mergeFunc func(existing string, generated string) (string, error)
