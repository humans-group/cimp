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
	default:
		return nil, fmt.Errorf("unsupported format: %v", m.format)
	}

	if err != nil {
		return nil, fmt.Errorf("marshal with format %q: %w", m.format, err)
	}

	return bytes, nil
}

func (m *kvMarshaler) Unmarshal(raw []byte) error {
	switch m.format {
	case JSONFormat:
		return fmt.Errorf("isn't available")
	case YAMLFormat:
		return yaml.Unmarshal(raw, &m.kv.KVTree)
	default:
		return fmt.Errorf("unsupported format: %v", m.format)
	}
}

func (pt *MarshalableTree) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("yaml node should have Mapping kind for unmarshal, not %v", node.Kind)
	}

	for i := 0; i < len(node.Content); i += 2 {
		curKey := node.Content[i].Value
		curNode := node.Content[i+1]

		switch curNode.Kind {
		case yaml.ScalarNode:
			leaf := &MarshalableLeaf{
				FullKey: makeFullKey(pt.fullKey, curKey),
				Name:    curKey,
				Value:   curNode.Value,
			}
			pt.add(curKey, leaf)
		case yaml.MappingNode:
			childTree := NewMarshalableTree(curKey, pt.fullKey)
			if err := childTree.UnmarshalYAML(curNode); err != nil {
				return fmt.Errorf("unmarshal %q: %w", curKey, err)
			}
			pt.add(curKey, childTree)
		case yaml.SequenceNode:
			leafValue, err := yaml.Marshal(curNode)
			if err != nil {
				return fmt.Errorf("marshal sequence %q: %w", curKey, err)
			}
			leaf := &MarshalableLeaf{
				FullKey: makeFullKey(pt.fullKey, curKey),
				Value:   string(leafValue),
			}
			pt.add(curKey, leaf)
		default:
			return fmt.Errorf("unprocessable content type of %q: %v", curKey, curNode.Kind)
		}
	}

	return nil
}

func (pl *MarshalableLeaf) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		pl.Value = node.Value
	case yaml.SequenceNode:
		leafValue, err := yaml.Marshal(node)
		if err != nil {
			return fmt.Errorf("marshal sequence %q: %w", pl.FullKey, err)
		}
		pl.Value = string(leafValue)
	default:
		return fmt.Errorf("unprocessable content type `%v` for leaf %q", node.Kind, pl.FullKey)
	}

	return nil
}

func (pt *MarshalableTree) ToYAMLNode() (*yaml.Node, error) {
	var content []*yaml.Node
	for _, leafName := range pt.Order {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode}

		if err := keyNode.Encode(leafName); err != nil {
			return nil, fmt.Errorf("encode node %q: %w", leafName, err)
		}
		content = append(content, keyNode)

		valueNode, err := pt.Tree[leafName].ToYAMLNode()
		if err != nil {
			return nil, fmt.Errorf("convert %q to YAML-node: %w", leafName, err)
		}
		content = append(content, valueNode)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Style:   0,
		Value:   pt.Name,
		Content: content,
		Line:    1,
		Column:  1,
	}, nil
}

func (pl *MarshalableLeaf) ToYAMLNode() (*yaml.Node, error) {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: fmt.Sprint(pl.Value),
	}, nil
}

func (pt *MarshalableTree) UnmarshalJSON([]byte) error {
	return nil
}

func (pt *MarshalableTree) MarshalJSON() ([]byte, error) {
	return nil, nil
}
