package generator

import (
	"docker-wizard/internal/generator/catalog"
	"docker-wizard/internal/generator/compose"
	"docker-wizard/internal/generator/dockerfile"
	"docker-wizard/internal/generator/preview"
	"docker-wizard/internal/generator/validate"
	"docker-wizard/internal/generator/write"
)

type Language = dockerfile.Language
type LanguageDetails = dockerfile.LanguageDetails
type ServiceSpec = catalog.ServiceSpec
type Output = write.Output
type Preview = preview.Preview
type FilePreview = preview.FilePreview
type FileStatus = preview.FileStatus
type ComposeSelection = compose.ComposeSelection

const (
	LanguageGo      = dockerfile.LanguageGo
	LanguageNode    = dockerfile.LanguageNode
	LanguagePython  = dockerfile.LanguagePython
	LanguageRuby    = dockerfile.LanguageRuby
	LanguagePHP     = dockerfile.LanguagePHP
	LanguageJava    = dockerfile.LanguageJava
	LanguageDotNet  = dockerfile.LanguageDotNet
	LanguageUnknown = dockerfile.LanguageUnknown
)

const (
	ComposeFileName      = write.ComposeFileName
	DockerfileFileName   = write.DockerfileFileName
	DockerignoreFileName = write.DockerignoreFileName
)

const (
	FileStatusNew       = preview.FileStatusNew
	FileStatusSame      = preview.FileStatusSame
	FileStatusDifferent = preview.FileStatusDifferent
	FileStatusExists    = preview.FileStatusExists
)

func DetectLanguage(root string) (LanguageDetails, error) {
	return dockerfile.DetectLanguage(root)
}

func Dockerfile(details LanguageDetails) (string, error) {
	return dockerfile.Dockerfile(details)
}

func Compose(root string, selection ComposeSelection) (string, error) {
	return compose.Compose(root, selection)
}

func SelectableServices(root string) ([]ServiceSpec, error) {
	return catalog.SelectableServices(root)
}

func CatalogMap(root string) (map[string]ServiceSpec, []ServiceSpec, error) {
	return catalog.CatalogMap(root)
}

func PreviewFiles(root string, compose string, dockerfile string) (Preview, error) {
	return preview.PreviewFiles(root, compose, dockerfile)
}

func SelectionWarnings(root string, selection ComposeSelection) ([]string, error) {
	return validate.SelectionWarnings(root, selection)
}

func WriteFiles(root string, compose string, dockerfile string) (Output, error) {
	return write.WriteFiles(root, compose, dockerfile)
}
