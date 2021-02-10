package cimp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/humans-group/cimp/lib/tree"
)

type Marshaler interface {
	Marshal() ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal([]byte) error
}

type kvMarshaler struct {
	kv           *KV
	format       FileFormat
	indentSpaces int
}

func NewMarshaler(kv *KV, format FileFormat, indentSpaces int) Marshaler {
	return &kvMarshaler{
		kv:           kv,
		format:       format,
		indentSpaces: indentSpaces,
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
		rawBuf bytes.Buffer
		err    error
	)

	switch m.format {
	case YAMLFormat:
		yamlEncoder := yaml.NewEncoder(&rawBuf)
		yamlEncoder.SetIndent(m.indentSpaces)
		err = yamlEncoder.Encode(m.kv.tree)
	case JSONFormat:
		jsonEncoder := json.NewEncoder(&rawBuf)
		jsonEncoder.SetIndent("", strings.Repeat("", m.indentSpaces))
		err = jsonEncoder.Encode(m.kv.tree)
	default:
		return nil, fmt.Errorf("unsupported marshal format: %v", m.format)
	}

	if err != nil {
		return nil, fmt.Errorf("%s-marshal of KV: %w", m.format, err)
	}

	return rawBuf.Bytes(), nil
}

func (m *kvMarshaler) Unmarshal(raw []byte) error {
	var err error
	switch m.format {
	case JSONFormat:
		err = json.Unmarshal(raw, &m.kv.tree)
	case YAMLFormat:
		err = yaml.Unmarshal(raw, &m.kv.tree)
	default:
		return fmt.Errorf("unsupported unmarshal format: %v", m.format)
	}

	if err != nil {
		return fmt.Errorf("%s-unmarshal of KV: %w", m.format, err)
	}

	m.kv.idx.clear()
	m.kv.idx.addKeys(m.kv.tree, tree.Path{})

	return nil
}
