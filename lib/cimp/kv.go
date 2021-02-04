package cimp

import (
	"fmt"
	"io/ioutil"

	"github.com/humans-group/cimp/lib/tree"
)

type KV struct {
	Tree             *tree.Tree
	Index            map[string]tree.Path
	ArrayValueFormat FileFormat
	GlobalPrefix     string
}

const (
	rootLevelName = ""
)

func NewKV(prefix string, arrayValueFormat FileFormat) *KV {
	return &KV{
		Tree:             tree.NewSubTree(rootLevelName, ""),
		Index:            make(map[string]tree.Path),
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

	leaf, err := kv.Tree.Get(path)
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

	leaf, err := kv.Tree.Get(path)
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
