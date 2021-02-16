package cimp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/humans-group/cimp/lib/tree"
)

type KV struct {
	tree         *tree.Tree
	idx          index
	globalPrefix string
}

type index map[string]tree.Path

type treeConverter struct {
	Format FileFormat
	Indent int
}

const consulSep = "/"

func NewKV(t *tree.Tree) *KV {
	idx := index(make(map[string]tree.Path))
	idx.addKeys(t, nil)

	return &KV{
		tree: t,
		idx:  idx,
	}
}

func (kv *KV) SetIfExist(key string, value interface{}) error {
	path, ok := kv.idx[key]
	if !ok {
		return nil
	}

	leaf, err := kv.tree.Get(path)
	if err != nil {
		return fmt.Errorf("get by path: %w", err)
	}

	leaf.Value = value

	return nil
}

func (kv *KV) GetString(key string) (string, error) {
	path, ok := kv.idx[key]
	if !ok {
		return "", fmt.Errorf("value by key %q: %w", key, ErrorNotFoundInKV)
	}

	leaf, err := kv.tree.Get(path)
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

func (kv *KV) Exists(fullKey string) bool {
	if _, ok := kv.idx[fullKey]; ok {
		return true
	}

	_, err := kv.tree.GetByFullKey(fullKey)

	return err == nil
}

func (kv *KV) AddIfNotSet(m tree.Marshalable) error {
	if _, ok := kv.idx[m.GetFullKey()]; ok {
		return nil
	}

	lastSepIdx := strings.LastIndexAny(m.GetFullKey(), consulSep)
	if lastSepIdx < 0 {
		return ErrorParentNotFoundInKV
	}
	parentFullKey := m.GetFullKey()[:lastSepIdx]

	parent, err := kv.tree.GetByFullKey(parentFullKey)
	if err != nil {
		if errors.Is(err, tree.ErrorNotFound) {
			return ErrorParentNotFoundInKV
		}
		return fmt.Errorf("get parent: %w", err)
	}

	if item, err := parent.GetByFullKey(m.GetFullKey()); err == nil || item != nil {
		return nil
	} else if !errors.Is(err, tree.ErrorNotFound) {
		return fmt.Errorf("check value %q existence: %w", m.GetFullKey(), err)
	}

	switch parItem := parent.(type) {
	case *tree.Tree:
		parItem.AddOrReplaceDirectly(m.GetName(), m)
	case *tree.Branch:
		parItem.Add(m)
	default:
		return ErrorTypeIncorrect
	}

	return nil
}

func (kv *KV) DeleteIfExists(fullKey string) error {
	if err := kv.tree.Delete(fullKey); err != nil {
		if errors.Is(err, tree.ErrorNotFound) {
			return nil
		}

		return fmt.Errorf("delete by key %q: %w", fullKey, err)
	}

	delete(kv.idx, fullKey)

	return nil
}

func (kv *KV) AddPrefix(prefix string) {
	if !strings.HasSuffix(prefix, consulSep) {
		prefix = prefix + consulSep
	}
	kv.globalPrefix = prefix
}

func (kv *KV) SetTree(t *tree.Tree) {
	kv.tree = t
	kv.idx.clear()
	kv.idx.addKeys(t, nil)
}

func (kv *KV) Walk(walkFunc tree.WalkFunc) {
	kv.tree.Walk(walkFunc)
}

func (kv *KV) DeepClone() *KV {
	newTree := kv.tree.DeepClone()
	newKV := NewKV(newTree)
	newKV.globalPrefix = kv.globalPrefix

	return newKV
}

func (kv *KV) ConvertBranchesToString(format FileFormat, indent int) error {
	tc := treeConverter{
		Format: format,
		Indent: indent,
	}

	convertedTree, err := tc.convertBranchesToString(kv.tree)
	if err != nil {
		return fmt.Errorf("convert branches to string: %w", err)
	}
	kv.SetTree(convertedTree)

	return nil
}

func (kv *KV) ConvertTreeNamesToCamelCase() {
	kv.setNamesToSnakeCase(kv.tree)
	kv.idx.clear()
	kv.idx.addKeys(kv.tree, nil)
}

func (kv *KV) setNamesToSnakeCase(m tree.Marshalable) {
	switch item := m.(type) {
	case *tree.Leaf:
		item.Name = tree.ToSnakeCase(item.Name)
	case *tree.Branch:
		for i := range item.Content {
			kv.setNamesToSnakeCase(item.Content[i])
		}
		item.Name = tree.ToSnakeCase(item.Name)
	case *tree.Tree:
		item.Name = tree.ToSnakeCase(item.Name)
		for k, v := range item.Content {
			kv.setNamesToSnakeCase(v)
			delete(item.Content, k)
			k = tree.ToSnakeCase(k)
			item.Content[k] = v
		}
		for i, name := range item.Order {
			item.Order[i] = tree.ToSnakeCase(name)
		}
	}
}

// ConvertBranchesToString walks recursively through the map and marshals all slices to strings
func (c treeConverter) convertBranchesToString(mt *tree.Tree) (*tree.Tree, error) {
	newTree := mt.ShallowClone()

	for k, v := range mt.Content {
		switch item := v.(type) {
		case *tree.Leaf:
			continue
		case *tree.Branch:
			buf := bytes.Buffer{}
			switch c.Format {
			case JSONFormat:
				e := json.NewEncoder(&buf)
				e.SetIndent("", strings.Repeat(" ", c.Indent))
				if err := e.Encode(item); err != nil {
					return nil, fmt.Errorf("JSON-encode %q: %w", k, err)
				}
			case YAMLFormat:
				e := yaml.NewEncoder(&buf)
				e.SetIndent(c.Indent)
				if err := e.Encode(item); err != nil {
					return nil, fmt.Errorf("YAML-encode %q: %w", k, err)
				}
			}

			leaf := tree.NewLeaf(k, mt.FullKey)
			leafValueBuf := bytes.NewBufferString("\n")
			leafValueBuf.Write(buf.Bytes())
			endLineAndIndent := []byte("\n" + strings.Repeat(" ", int(leaf.GetNestingLevel())*c.Indent))
			leafValue := bytes.ReplaceAll(
				leafValueBuf.Bytes(),
				[]byte("\n"),
				endLineAndIndent,
			)
			leafValue = bytes.TrimSuffix(leafValue, endLineAndIndent)
			leaf.Value = string(leafValue)

			newTree.AddOrReplaceDirectly(k, leaf)
		case *tree.Tree:
			newItem, err := c.convertBranchesToString(item)
			if err != nil {
				return nil, fmt.Errorf("convert branches of tree %q: %w", k, err)
			}
			newTree.AddOrReplaceDirectly(k, newItem)
		}
	}

	return newTree, nil
}

func (idx index) clear() {
	for k := range idx {
		delete(idx, k)
	}
}

func (idx index) addKeys(m tree.Marshalable, curPath tree.Path) {
	path := make(tree.Path, len(curPath))
	copy(path, curPath)

	switch cur := m.(type) {
	case *tree.Leaf:
		idx[cur.FullKey] = path
	case *tree.Tree:
		for _, k := range cur.Order {
			idx.addKeys(cur.Content[k], append(path, k))
		}
	case *tree.Branch:
		for k, v := range cur.Content {
			idx.addKeys(v, append(path, strconv.Itoa(k)))
		}
	}
}
