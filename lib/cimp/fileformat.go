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
	JsonFormat FileFormat = "json"
	YamlFormat FileFormat = "yaml"
	EdnFormat  FileFormat = "edn"
)

func InitFormat(format, path string) (FileFormat, error) {
	if len(format) > 1 {
		switch FileFormat(format) {
		case JsonFormat:
			return JsonFormat, nil
		case YamlFormat:
			return YamlFormat, nil
		case EdnFormat:
			return EdnFormat, nil
		default:
			return "", fmt.Errorf("undefined format: %s", format)
		}
	}

	switch filepath.Ext(path) {
	case ".json":
		return JsonFormat, nil
	case ".yml", ".yaml":
		return YamlFormat, nil
	case ".edn":
		return EdnFormat, nil
	}

	return YamlFormat, nil
}

func UnmarshalWithFormat(format FileFormat, fileData []byte) (map[interface{}]interface{}, error) {
	cfgRaw := make(map[interface{}]interface{})
	var err error

	switch format {
	case JsonFormat:
		err = json.Unmarshal(fileData, &cfgRaw)
	case YamlFormat:
		err = yaml.Unmarshal(fileData, &cfgRaw)
	case EdnFormat:
		err = edn.Unmarshal(fileData, &cfgRaw)
	default:
		return nil, fmt.Errorf("unsupported format: %v", format)
	}

	if err != nil {
		return nil, fmt.Errorf("unmarshal with format %q error: %w", format, err)
	}

	return cfgRaw, nil
}
