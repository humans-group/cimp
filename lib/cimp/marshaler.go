package cimp

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
	"olympos.io/encoding/edn"
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
	raw := make(map[string]interface{})
	m.kv.KVTree.exportRecursive(raw, m)

	var (
		err   error
		bytes []byte
	)
	switch m.format {
	case JSONFormat:
		bytes, err = json.Marshal(raw)
	case YAMLFormat:
		bytes, err = yaml.Marshal(raw)
	case EDNFormat:
		bytes, err = edn.Marshal(raw)
	default:
		return nil, fmt.Errorf("unsupported format: %v", m.format)
	}

	if err != nil {
		return nil, fmt.Errorf("marshal with format %q: %w", m.format, err)
	}

	return bytes, nil
}

func (m *kvMarshaler) Unmarshal(raw []byte) error {
	cfgRaw := make(map[string]interface{})
	var err error

	switch m.format {
	case JSONFormat:
		err = json.Unmarshal(raw, &cfgRaw)
	case YAMLFormat:
		err = yaml.Unmarshal(raw, &cfgRaw)
	case EDNFormat:
		err = edn.Unmarshal(raw, &cfgRaw)
	default:
		return fmt.Errorf("unsupported format: %v", m.format)
	}

	if err != nil {
		return fmt.Errorf("unmarshal with format %q: %w", m.format, err)
	}

	return m.kv.KVTree.importRecursive("", cfgRaw, []string{}, m.kv)
}
