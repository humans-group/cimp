package cimp

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"olympos.io/encoding/edn"
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
}

type Path []string

type ProcessableLeaf struct {
	Key   string
	Value interface{}
}

const sep = "/"

func (pt ProcessableTree) Process() error {
	return nil
}

func (pt *ProcessableLeaf) Process() error {
	return nil
}

func NewProcessableTree() *ProcessableTree {
	return &ProcessableTree{
		Tree:      make(map[string]Processable),
		LevelName: "",
	}
}

func NewKV(prefix string, arrayValueFormat FileFormat) *KV {
	return &KV{
		KVTree:           NewProcessableTree(),
		Index:            make(map[string]Path),
		ArrayValueFormat: arrayValueFormat,
		GlobalPrefix:     prefix,
	}
}

func (kv *KV) FillFromFile(path string, format FileFormat) error {
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file data: %v", err)
	}

	rawData, err := UnmarshalWithFormat(format, fileData)
	if err != nil {
		return fmt.Errorf("unmarshal %q-file: %w", format, err)
	}

	return kv.KVTree.fillRecursive("", rawData, []string{}, kv)
}

func (kv *KV) Fill(prefix string, rawData map[interface{}]interface{}) error {
	return kv.KVTree.fillRecursive(prefix, rawData, []string{}, kv)
}

func (kv *KV) Check(key string) bool {
	_, ok := kv.Index[key]
	return ok
}

func (kv KV) GetString(key string) (string, error) {
	path, ok := kv.Index[key]
	if !ok {
		return "", fmt.Errorf("value %q: %w", key, ErrorNotFoundInKV)
	}

	leaf, err := kv.Get(path)
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

func (kv KV) Get(path Path) (*ProcessableLeaf, error) {
	if len(path) < 1 {
		return nil, fmt.Errorf("path is empty")
	}
	curLevel := kv.KVTree.Tree
	for i, breadcrumb := range path {
		if _, ok := curLevel[breadcrumb]; !ok {
			return nil, fmt.Errorf("path %v is incorrect", path)
		}
		switch nextLevel := curLevel[breadcrumb].(type) {
		case ProcessableTree:
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

func (kv KV) AddPrefix(prefix string) {
	kv.GlobalPrefix = prefix
}

func (pt ProcessableTree) fillRecursive(prefix string, data map[interface{}]interface{}, path []string, kv *KV) error {
	for key, rawValue := range data {
		stringKey, err := keyToString(key)
		if err != nil {
			return fmt.Errorf("failed to convert key %#v to string: %w", key, err)
		}

		newPath := make([]string, len(path)+1)
		copy(path, newPath)
		newPath[len(path)] = stringKey

		fullKey := makeFullKey(prefix, stringKey)
		switch value := rawValue.(type) {
		case map[interface{}]interface{}:
			childTree := NewProcessableTree()
			if err := childTree.fillRecursive(fullKey, value, newPath, kv); err != nil {
				return err
			}
			pt.Tree[stringKey] = childTree
		case []interface{}:
			marshaledArray, err := MarshalWithFormat(kv.ArrayValueFormat, data)
			if err != nil {
				return fmt.Errorf("marshal array %#v to %q: %w", data, kv.ArrayValueFormat, err)
			}
			pt.Tree[stringKey] = &ProcessableLeaf{
				Key:   fullKey,
				Value: string(marshaledArray),
			}
			kv.Index[stringKey] = newPath
		default:
			pt.Tree[stringKey] = &ProcessableLeaf{
				Key:   fullKey,
				Value: value,
			}
			kv.Index[stringKey] = newPath
		}
	}

	return nil
}

func keyToString(key interface{}) (string, error) {
	var result string

	switch curKey := key.(type) {
	case string:
		result = curKey
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		result = fmt.Sprintf("%d", curKey)
	case float32, float64:
		result = fmt.Sprintf("%f", curKey)
	case bool:
		result = fmt.Sprintf("%t", curKey)
	case edn.Keyword:
		strKey := curKey.String()
		if len(strKey) < 1 {
			return "", fmt.Errorf("edn-key %#v is empty or not stringable", key)
		}
		// remove `:` from beginning of the key
		result = fmt.Sprintf("%s", strKey[1:])
	default:
		return "", fmt.Errorf("invalid config key type: %#v", curKey)
	}

	return result, nil
}

func makeFullKey(key string, prefix string) string {
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
