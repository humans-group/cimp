package cimp

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/consul/api"
)

type KV map[string]interface{}

func NewKV() KV {
	kv := KV(make(map[string]interface{}))
	return kv
}

func (kv KV) FillFromFile(path string, format FileFormat) error {
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file data: %v", err)
	}

	rawData, err := UnmarshalWithFormat(format, fileData)
	if err != nil {
		return fmt.Errorf("unmarshal %q-file: %w", format, err)
	}

	return kv.fillRecursive("", rawData)
}

func (kv KV) Fill(prefix string, rawData map[interface{}]interface{}) error {
	return kv.fillRecursive(prefix, rawData)
}

func (kv KV) Check(key string) bool {
	_, ok := kv[key]
	return ok
}

func (kv KV) GetString(key string) (string, error) {
	value, ok := kv[key]
	if !ok {
		return "", fmt.Errorf("value %q: %w", key, ErrorNotFoundInKV)
	}

	switch typed := value.(type) {
	case string:
		return typed, nil
	}

	return "", fmt.Errorf("value %q: %w", key, ErrorTypeIncorrect)
}

func (kv KV) AddPair(pair api.KVPair) {
	kv[pair.Key] = pair.Value
}

func (kv KV) AddPrefix(prefix string) {
	newKV := NewKV()
	for key, value := range kv {
		newKV[prefix+key] = value
		delete(kv, key)
	}
	for key, value := range newKV {
		kv[key] = value
		delete(newKV, key)
	}
}

func (kv KV) fillRecursive(prefix string, rawData interface{}) error {
	switch data := rawData.(type) {
	case map[interface{}]interface{}:
		for key := range data {
			stringKey, err := keyToString(prefix, key)
			if err != nil {
				return fmt.Errorf("failed to convert key %#v to string: %w", key, err)
			}
			if err := kv.fillRecursive(stringKey, data[key]); err != nil {
				return err
			}
		}
	case []interface{}:
		yamlData, err := yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal array %#v to yaml: %w", data, err)
		}
		kv[prefix] = string(yamlData)
	default:
		kv[prefix] = rawData
	}

	return nil
}

func keyToString(prefix string, key interface{}) (string, error) {
	var result string
	if len(prefix) > 0 {
		result = prefix + "/"
	}

	switch curKey := key.(type) {
	case string:
		result += curKey
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		result += fmt.Sprintf("%d", curKey)
	case float32, float64:
		result += fmt.Sprintf("%f", curKey)
	case bool:
		result += fmt.Sprintf("%t", curKey)
	default:
		panic(fmt.Sprintf("invalid config key type: %#v", curKey))
	}

	return strings.ToLower(result), nil
}
