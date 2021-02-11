package cimp

import (
	"fmt"
	"path/filepath"
)

type FileFormat string

const (
	JSONFormat FileFormat = "json"
	YAMLFormat FileFormat = "yaml"
)

func NewFormat(format, path string) (FileFormat, error) {
	if len(format) > 1 {
		switch FileFormat(format) {
		case JSONFormat:
			return JSONFormat, nil
		case YAMLFormat:
			return YAMLFormat, nil
		default:
			return "", fmt.Errorf("undefined format: %s", format)
		}
	}

	switch filepath.Ext(path) {
	case ".json":
		return JSONFormat, nil
	case ".yml", ".yaml":
		return YAMLFormat, nil
	}

	return YAMLFormat, nil
}
