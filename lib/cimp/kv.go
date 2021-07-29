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

type TreeConverter interface {
	Convert(tree.Tree) (*tree.Tree, error)
}

type branchesToStringConverter struct {
	Format FileFormat
	Indent int
}

type branchesToTreeConverter struct {
	branchPathToBranchElementFieldName map[string]string
	onlyKeys                           bool // it'll convert only keys as for tree. Don't use converted tree after that!!! It'll be broken and can be only marshalled.
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
	tc := branchesToStringConverter{
		Format: format,
		Indent: indent,
	}

	convertedTree, err := tc.Convert(kv.tree)
	if err != nil {
		return fmt.Errorf("convert branches to string: %w", err)
	}
	kv.SetTree(convertedTree)

	return nil
}

func (kv *KV) ConvertBranchesToTree(branchPathToBranchElementFieldName map[string]string) error {
	tc := branchesToTreeConverter{
		branchPathToBranchElementFieldName: branchPathToBranchElementFieldName,
		onlyKeys:                           false,
	}

	convertedTree, err := tc.Convert(kv.tree)
	if err != nil {
		return fmt.Errorf("convert branches to trees: %w", err)
	}
	kv.SetTree(convertedTree)

	return nil
}

// ConvertBranchesKeysAsForTree converts only keys.
// It can be useful before marshaling, but if you want to change values or something after that - keys may to be returned to default values.
func (kv *KV) ConvertBranchesKeysAsForTree(branchPathToBranchElementFieldName map[string]string) error {
	tc := branchesToTreeConverter{
		branchPathToBranchElementFieldName: branchPathToBranchElementFieldName,
		onlyKeys:                           true,
	}

	convertedTree, err := tc.Convert(kv.tree)
	if err != nil {
		return fmt.Errorf("convert branches' keys as for tree: %w", err)
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

// Convert walks recursively through the map and marshals all slices to strings
func (c branchesToStringConverter) Convert(mt *tree.Tree) (*tree.Tree, error) {
	newTree := mt.ShallowClone()

	for k, v := range mt.Content {
		switch item := v.(type) {
		case *tree.Leaf:
			continue // already added by ShallowClone
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
			newItem, err := c.Convert(item)
			if err != nil {
				return nil, fmt.Errorf("convert branches of tree %q: %w", k, err)
			}
			newTree.AddOrReplaceDirectly(k, newItem)
		}
	}

	return newTree, nil
}

// Convert walks recursively through the map and marshals needed slices to trees
func (c branchesToTreeConverter) Convert(mt *tree.Tree) (*tree.Tree, error) {
	newTree := mt.ShallowClone()

	for k, v := range mt.Content {
		switch item := v.(type) {
		case *tree.Leaf:
			continue // already added by ShallowClone
		case *tree.Branch:
			fieldName, isBranchShouldBeConverted := c.branchPathToBranchElementFieldName[v.GetFullKey()]
			if !isBranchShouldBeConverted {
				continue
			}

			convertedTree := tree.NewSubTree(v.GetName(), mt.GetFullKey())
			branchWithConvertedKeys := v.(*tree.Branch).DeepClone()
			for elementIdx, branchElement := range v.(*tree.Branch).Content {
				branchElementAsTree, isTree := branchElement.(*tree.Tree)
				branchElementName := ""
				if !isTree {
					return nil, fmt.Errorf("branch %q should be converted to tree, but it's not branch of trees", v.GetFullKey())
				}

				isBranchElementHaveField := false
				for _, treeElement := range branchElementAsTree.Content {
					if treeElement.GetName() != fieldName {
						continue
					}
					isBranchElementHaveField = true
					treeElementAsLeaf, isTreeElementLeaf := treeElement.(*tree.Leaf)
					if !isTreeElementLeaf {
						return nil, fmt.Errorf("field %q of element #%d of branch %q is not a leaf", fieldName, elementIdx+1, v.GetFullKey())
					}
					treeElementValueAsString, isTreeElementValueString := treeElementAsLeaf.Value.(string)
					if !isTreeElementValueString {
						return nil, fmt.Errorf("field %q of element #%d of branch %q is not a string", fieldName, elementIdx+1, v.GetFullKey())
					}
					branchElementName = treeElementValueAsString
					break
				}
				if !isBranchElementHaveField {
					return nil, fmt.Errorf("branch element #%d of branch %q doesn't have field %q", elementIdx+1, v.GetFullKey(), fieldName)
				}

				if !c.onlyKeys {
					convertedTree.AddOrReplaceDirectly(branchElementName, branchElementAsTree)
				} else {
					branchWithConvertedKeys.Content[elementIdx].ChangeName(branchElementName, v.GetFullKey())
				}
			}
			if !c.onlyKeys {
				newTree.AddOrReplaceDirectly(k, convertedTree)
			} else {
				newTree.AddOrReplaceDirectly(k, branchWithConvertedKeys)
			}
		case *tree.Tree:
			newItem, err := c.Convert(item)
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
