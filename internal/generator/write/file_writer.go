package write

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeManagedFile(root string, path string, pattern string, content string, merge mergeFunc) (WriteStatus, string, error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return "", "", fmt.Errorf("%s is a directory", filepath.Base(path))
		}
		existing, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", "", fmt.Errorf("read %s: %w", filepath.Base(path), readErr)
		}

		targetContent := content
		if merge != nil {
			mergedContent, mergeErr := merge(string(existing), content)
			if mergeErr != nil {
				return "", "", fmt.Errorf("merge %s: %w", filepath.Base(path), mergeErr)
			}
			targetContent = mergedContent
		}

		if string(existing) == targetContent {
			return WriteStatusUnchanged, "", nil
		}

		backupPath := path + ".bak"
		if err := os.WriteFile(backupPath, existing, info.Mode().Perm()); err != nil {
			return "", "", fmt.Errorf("write %s backup: %w", filepath.Base(path), err)
		}

		if err := writeByTemp(root, path, pattern, targetContent); err != nil {
			return "", "", err
		}

		return WriteStatusUpdated, backupPath, nil
	}
	if !os.IsNotExist(err) {
		return "", "", fmt.Errorf("stat %s: %w", filepath.Base(path), err)
	}

	if err := writeByTemp(root, path, pattern, content); err != nil {
		return "", "", err
	}

	return WriteStatusCreated, "", nil
}

func writeByTemp(root string, path string, pattern string, content string) error {
	tempPath, err := writeTempFile(root, pattern, content)
	if err != nil {
		return fmt.Errorf("write %s temp file: %w", filepath.Base(path), err)
	}
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("move %s into place: %w", filepath.Base(path), err)
	}

	return nil
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
	if err := tempFile.Chmod(0o644); err != nil {
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
