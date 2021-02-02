package cimp

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type Processable interface {
	Process() error
}

type Processor interface {
	Process(ProcessableLeaf) error
}

type KV struct {
	KVTree           *ProcessableTree
	Index            map[string]Path
	ArrayValueFormat FileFormat
	GlobalPrefix     string
}

type ProcessableTree struct {
	Tree      map[string]Processable
	LevelName string
	Order     []string // TODO: order not working in marshalling/unmarshalling. Try to make it with Decoder
}

type Path []string

type ProcessableLeaf struct {
	FullKey string
	Value   interface{}
}

const (
	sep                       = "/"
	rootLevelName             = ""
	keyTemplateForMarshalling = "{{key}}"
)

func (pt ProcessableTree) Process() error {
	return nil
}

func (pt *ProcessableLeaf) Process() error {
	return nil
}

func NewProcessableTree(levelName string) *ProcessableTree {
	return &ProcessableTree{
		Tree:      make(map[string]Processable),
		LevelName: levelName,
	}
}

func NewKV(prefix string, arrayValueFormat FileFormat) *KV {
	return &KV{
		KVTree:           NewProcessableTree(rootLevelName),
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

func (kv *KV) Fill(prefix string, rawData map[string]interface{}) error {
	return kv.KVTree.importRecursive(prefix, rawData, []string{}, kv)
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

func (kv *KV) get(path Path) (*ProcessableLeaf, error) {
	if len(path) < 1 {
		return nil, fmt.Errorf("path is empty")
	}
	curLevel := kv.KVTree.Tree
	for i, breadcrumb := range path {
		if _, ok := curLevel[breadcrumb]; !ok {
			return nil, fmt.Errorf("path %v is incorrect", path)
		}
		switch nextLevel := curLevel[breadcrumb].(type) {
		case *ProcessableTree:
			curLevel = nextLevel.Tree
			continue
		case *ProcessableLeaf:
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

func (kv *KV) AddPrefix(prefix string) {
	if len(prefix) > 0 && prefix[len(prefix)-1] != '/' {
		prefix = prefix + "/"
	}
	kv.GlobalPrefix = prefix
}

func (pt *ProcessableTree) importRecursive(prefix string, data map[string]interface{}, path []string, kv *KV) error {
	for key, rawValue := range data {
		pt.Order = append(pt.Order, key)

		newPath := make([]string, len(path)+1)
		copy(newPath, path)
		newPath[len(path)] = key

		fullKey := makeFullKey(prefix, key)

		switch value := rawValue.(type) {
		case map[string]interface{}:
			childTree := NewProcessableTree(key)
			if err := childTree.importRecursive(fullKey, value, newPath, kv); err != nil {
				return err
			}
			pt.Tree[key] = childTree
		case []interface{}:
			marshaledArray, err := MarshalWithFormat(kv.ArrayValueFormat, value)
			if err != nil {
				return fmt.Errorf("marshal array %#v to %q: %w", value, kv.ArrayValueFormat, err)
			}
			pt.Tree[key] = &ProcessableLeaf{
				FullKey: fullKey,
				Value:   string(marshaledArray),
			}
			kv.Index[fullKey] = newPath
		default:
			pt.Tree[key] = &ProcessableLeaf{
				FullKey: fullKey,
				Value:   value,
			}
			kv.Index[fullKey] = newPath
		}
	}

	return nil
}

func (pt *ProcessableTree) exportRecursive(to map[string]interface{}, m *kvMarshaler) {
	for _, key := range pt.Order {
		val := pt.Tree[key]
		if m.toSnakeCase {
			key = ToSnakeCase(key)
		}
		if len(m.keyPrefix) > 0 {
			key = m.keyPrefix + key
		}

		switch value := val.(type) {
		case *ProcessableTree:
			newLevel := make(map[string]interface{})
			value.exportRecursive(newLevel, m)
			to[key] = newLevel
		case *ProcessableLeaf:
			if m.toTemplate {
				to[key] = strings.ReplaceAll(m.template, keyTemplateForMarshalling, value.FullKey)
				continue
			}
			to[key] = value.Value
		}
	}
}

func makeFullKey(prefix string, key string) string {
	key = ToSnakeCase(key)
	if len(prefix) > 0 {
		key = prefix + sep + key
	}

	return key
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
var matchAllSpecSymbols = regexp.MustCompile("[^A-z0-9]")
var matchAllMultipleUnderscore = regexp.MustCompile("[_]{2,}")

func ToSnakeCase(str string) string {
	str = matchAllSpecSymbols.ReplaceAllString(str, "_")
	str = matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	str = matchAllCap.ReplaceAllString(str, "${1}_${2}")
	str = matchAllMultipleUnderscore.ReplaceAllString(str, "_")
	str = strings.Trim(str, "_")

	return strings.ToLower(str)
}
