package cimp

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/go-yaml/yaml"
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

func UnmarshalWithFormat(format FileFormat, fileData []byte) (map[interface{}]interface{}, error) {
	cfgRaw := make(map[interface{}]interface{})
	var err error

	switch format {
	case JSONFormat:
		err = json.Unmarshal(fileData, &cfgRaw)
	case YAMLFormat:
		err = yaml.Unmarshal(fileData, &cfgRaw)
	case EDNFormat:
		err = edn.Unmarshal(fileData, &cfgRaw)
	default:
		return nil, fmt.Errorf("unsupported format: %v", format)
	}

	if err != nil {
		return nil, fmt.Errorf("unmarshal with format %q error: %w", format, err)
	}

	return cfgRaw, nil
}
