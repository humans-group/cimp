package cimp

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Marshaler interface {
	Marshal() ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal([]byte) error
}

type kvMarshaler struct {
	kv          *KV
	format      FileFormat
	toSnakeCase bool   // convert keys to snake case in Marshal()
	keyPrefix   string // add prefix to keys in Marshal()
	toTemplate  bool   // save instead of values templates with keys
	template    string // template with "{{key}}"
}

func NewMarshaler(kv *KV, format FileFormat, marshalToSnakeCase, marshalToTemplate bool, keyPrefix, keyTemplate string) Marshaler {
	return &kvMarshaler{
		kv:          kv,
		format:      format,
		toSnakeCase: marshalToSnakeCase,
		keyPrefix:   keyPrefix,
		toTemplate:  marshalToTemplate,
		template:    keyTemplate,
	}
}

func NewUnmarshaler(kv *KV, format FileFormat) Unmarshaler {
	return &kvMarshaler{
		kv:     kv,
		format: format,
	}
}

func (m *kvMarshaler) Marshal() ([]byte, error) {
	var (
		err      error
		byteList []byte
	)
	switch m.format {
	case JSONFormat:
		byteList, err = json.Marshal(m.kv.tree)
	case YAMLFormat:
		byteList, err = yaml.Marshal(m.kv.tree)
	default:
		return nil, fmt.Errorf("unsupported format: %v", m.format)
	}

	if err != nil {
		return nil, fmt.Errorf("marshal with format %q: %w", m.format, err)
	}

	return byteList, nil
}

func (m *kvMarshaler) Unmarshal(raw []byte) error {
	switch m.format {
	case JSONFormat:
		return fmt.Errorf("isn't available")
	case YAMLFormat:
		return yaml.Unmarshal(raw, &m.kv.tree)
	default:
		return fmt.Errorf("unsupported format: %v", m.format)
	}
}
