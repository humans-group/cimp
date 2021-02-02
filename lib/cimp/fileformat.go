package cimp

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"olympos.io/encoding/edn"
)

type FileFormat string

const (
	JSONFormat FileFormat = "json"
	YAMLFormat FileFormat = "yaml"
	EDNFormat  FileFormat = "edn"
)

func NewFormat(format, path string) (FileFormat, error) {
	if len(format) > 1 {
		switch FileFormat(format) {
		case JSONFormat:
			return JSONFormat, nil
		case YAMLFormat:
			return YAMLFormat, nil
		case EDNFormat:
			return EDNFormat, nil
		default:
			return "", fmt.Errorf("undefined format: %s", format)
		}
	}

	switch filepath.Ext(path) {
	case ".json":
		return JSONFormat, nil
	case ".yml", ".yaml":
		return YAMLFormat, nil
	case ".edn":
		return EDNFormat, nil
	}

	return YAMLFormat, nil
}

func MarshalWithFormat(format FileFormat, smth interface{}) ([]byte, error) {
	var (
		marshaled []byte
		err       error
	)

	switch format {
	case JSONFormat:
		marshaled, err = json.Marshal(smth)
	case YAMLFormat:
		marshaled, err = yaml.Marshal(smth)
	case EDNFormat:
		marshaled, err = edn.Marshal(smth)
	default:
		return nil, fmt.Errorf("unsupported format: %v", format)
	}

	if err != nil {
		return nil, fmt.Errorf("marshal with format %q error: %w", format, err)
	}

	return marshaled, nil
}
