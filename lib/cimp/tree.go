package cimp

import "strings"

func NewMarshalableTree(name, prefix string) *MarshalableTree {
	return &MarshalableTree{
		Tree:    make(map[string]Marshalable),
		Name:    name,
		fullKey: makeFullKey(prefix, name),
	}
}

func (pt *MarshalableTree) exportRecursive(to map[string]interface{}, m *kvMarshaler) {
	for _, key := range pt.Order {
		val := pt.Tree[key]
		if m.toSnakeCase {
			key = ToSnakeCase(key)
		}
		if len(m.keyPrefix) > 0 {
			key = m.keyPrefix + key
		}

		switch value := val.(type) {
		case *MarshalableTree:
			newLevel := make(map[string]interface{})
			value.exportRecursive(newLevel, m)
			to[key] = newLevel
		case *MarshalableLeaf:
			if m.toTemplate {
				to[key] = strings.ReplaceAll(m.template, keyTemplateForMarshalling, value.FullKey)
				continue
			}
			to[key] = value.Value
		}
	}
}

func (pt *MarshalableTree) add(key string, value Marshalable) {
	pt.Order = append(pt.Order, key)
	pt.Tree[key] = value
}
