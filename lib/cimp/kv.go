package cimp

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// TODO: add the same for JSON if it'll be needed
type Marshalable interface {
	ToYAMLNode() (*yaml.Node, error)
	UnmarshalYAML(value *yaml.Node) error
}

type KV struct {
	KVTree           *MarshalableTree
	Index            map[string]Path
	ArrayValueFormat FileFormat
	GlobalPrefix     string
}

type MarshalableTree struct {
	Tree    map[string]Marshalable
	Name    string
	Order   []string
	fullKey string
}

type MarshalableLeaf struct {
	FullKey string
	Name    string
	Value   interface{}
}

type Path []string

const (
	sep                       = "/"
	rootLevelName             = ""
	keyTemplateForMarshalling = "{{key}}"
)

func NewKV(prefix string, arrayValueFormat FileFormat) *KV {
	return &KV{
		KVTree:           NewMarshalableTree(rootLevelName, ""),
		Index:            make(map[string]Path),
		ArrayValueFormat: arrayValueFormat,
		GlobalPrefix:     prefix,
	}
}

func (kv *KV) FillFromFile(path string, format FileFormat) error {
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file data: %w", err)
	}

	marshaller := NewUnmarshaler(kv, format)
	if err := marshaller.Unmarshal(fileData); err != nil {
		return fmt.Errorf("unmarshal file data: %w", err)
	}

	return nil
}

func (kv *KV) Check(key string) bool {
	_, ok := kv.Index[key]
	return ok
}

func (kv *KV) SetIfExist(key string, value interface{}) error {
	path, ok := kv.Index[key]
	if !ok {
		return nil
	}

	leaf, err := kv.get(path)
	if err != nil {
		return fmt.Errorf("get by path: %w", err)
	}

	leaf.Value = value

	return nil
}

func (kv *KV) GetString(key string) (string, error) {
	path, ok := kv.Index[key]
	if !ok {
		return "", fmt.Errorf("value by key %q: %w", key, ErrorNotFoundInKV)
	}

	leaf, err := kv.get(path)
	if err != nil {
		return "", fmt.Errorf("get by path: %w", err)
	}

	switch value := leaf.Value.(type) {
	case string:
		return value, nil
	default:
		return "", fmt.Errorf("value %q: %w", key, ErrorTypeIncorrect)
	}
}

func (kv *KV) AddPrefix(prefix string) {
	if len(prefix) > 0 && prefix[len(prefix)-1] != '/' {
		prefix = prefix + "/"
	}
	kv.GlobalPrefix = prefix
}

func (kv *KV) get(path Path) (*MarshalableLeaf, error) {
	if len(path) < 1 {
		return nil, fmt.Errorf("path is empty")
	}
	curLevel := kv.KVTree.Tree
	for i, breadcrumb := range path {
		if _, ok := curLevel[breadcrumb]; !ok {
			return nil, fmt.Errorf("path %v is incorrect", path)
		}
		switch nextLevel := curLevel[breadcrumb].(type) {
		case *MarshalableTree:
			curLevel = nextLevel.Tree
			continue
		case *MarshalableLeaf:
			if i != len(path)-1 {
				return nil, fmt.Errorf("path %v is too long", path)
			}
			return nextLevel, nil
		default:
			return nil, fmt.Errorf("tree value type %T is incorrect", nextLevel)
		}
	}

	return nil, ErrorNotFoundInKV
}
