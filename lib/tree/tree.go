package tree

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Marshalable interface {
	MarshalYAML() (interface{}, error)
	UnmarshalYAML(value *yaml.Node) error
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(raw []byte) error
}

type Tree struct {
	Content      map[string]Marshalable
	Name         string
	Order        []string
	FullKey      string
	NestingLevel uint
	decoder      *json.Decoder
}

type Branch struct {
	Content      []Marshalable
	Name         string
	FullKey      string
	NestingLevel uint
	decoder      *json.Decoder
}

type Leaf struct {
	Value            interface{}
	Name             string
	FullKey          string
	decoder          *json.Decoder
	NestingLevel     uint
	yamlMarshalStyle yaml.Style
}

type Path []string

func New() *Tree {
	return &Tree{
		Content: make(map[string]Marshalable),
	}
}

func NewSubTree(name, parentFullKey string, parentNestingLevel uint) *Tree {
	return &Tree{
		Content:      make(map[string]Marshalable),
		Name:         name,
		FullKey:      MakeFullKey(parentFullKey, name),
		NestingLevel: parentNestingLevel + 1,
	}
}

func NewBranch(name, parentFullKey string, parentNestingLevel uint) *Branch {
	return &Branch{
		Content:      []Marshalable{},
		Name:         name,
		FullKey:      MakeFullKey(parentFullKey, name),
		NestingLevel: parentNestingLevel + 1,
	}
}

func NewLeaf(name, parentFullKey string, parentNestingLevel uint) *Leaf {
	return &Leaf{
		Name:         name,
		FullKey:      MakeFullKey(parentFullKey, name),
		NestingLevel: parentNestingLevel + 1,
	}
}

func (mt *Tree) Get(path Path) (*Leaf, error) {
	if len(path) < 1 {
		return nil, ErrorNotFound
	}

	desiredName := path[0]
	if _, ok := mt.Content[desiredName]; !ok {
		return nil, fmt.Errorf("desired value %q is not found in tree %q", desiredName, mt.Name)
	}
	switch found := mt.Content[desiredName].(type) {
	case *Tree:
		return found.Get(path[1:])
	case *Branch:
		return found.Get(path[1:])
	case *Leaf:
		if len(path) != 1 {
			return nil, fmt.Errorf("path is too long, %q already a leaf", found.Name)
		}
		return found, nil
	default:
		return nil, fmt.Errorf("tree value type %T is incorrect", found)
	}
}

func (mb *Branch) Get(path Path) (*Leaf, error) {
	if len(path) < 1 {
		return nil, ErrorNotFound
	}

	desiredIdx, err := strconv.Atoi(path[0])
	if err != nil {
		return nil, fmt.Errorf("not-number index value %q for branch %q", path[0], mb.Name)
	}
	if len(mb.Content)-1 < desiredIdx {
		return nil, fmt.Errorf("desired value #%d is not found in branch %q", desiredIdx, mb.Name)
	}
	switch found := mb.Content[desiredIdx].(type) {
	case *Tree:
		return found.Get(path[1:])
	case *Branch:
		return found.Get(path[1:])
	case *Leaf:
		if len(path) != 1 {
			return nil, fmt.Errorf("path is too long, %q already a leaf", found.Name)
		}
		return found, nil
	default:
		return nil, fmt.Errorf("tree value type %T is incorrect", found)
	}
}

func (mt *Tree) AddOrReplaceDirectly(name string, value Marshalable) {
	if _, ok := mt.Content[name]; !ok {
		mt.Order = append(mt.Order, name)
	}
	switch item := value.(type) {
	case *Tree:
		item.NestingLevel = mt.NestingLevel + 1
	case *Branch:
		item.NestingLevel = mt.NestingLevel + 1
	case *Leaf:
		item.NestingLevel = mt.NestingLevel + 1
	}

	mt.Content[name] = value
}

func (mb *Branch) Add(value Marshalable) {
	mb.Content = append(mb.Content, value)
}

func (mt *Tree) ShallowClone() *Tree {
	newOrder := make([]string, len(mt.Order))
	copy(newOrder, mt.Order)

	newContent := make(map[string]Marshalable, len(mt.Content))
	for k, v := range mt.Content {
		newContent[k] = v
	}

	newTree := &Tree{
		Content:      newContent,
		Name:         mt.Name,
		Order:        newOrder,
		FullKey:      mt.FullKey,
		decoder:      mt.decoder,
		NestingLevel: mt.NestingLevel,
	}

	return newTree
}

func (mt *Tree) DeepClone() *Tree {
	newOrder := make([]string, len(mt.Order))
	copy(newOrder, mt.Order)

	newContent := make(map[string]Marshalable, len(mt.Content))
	for k, v := range mt.Content {
		switch item := v.(type) {
		case *Tree:
			newContent[k] = item.DeepClone()
		case *Branch:
			newContent[k] = item.DeepClone()
		case *Leaf:
			newContent[k] = item.DeepClone()
		}
	}

	newTree := &Tree{
		Content:      newContent,
		Name:         mt.Name,
		Order:        newOrder,
		FullKey:      mt.FullKey,
		NestingLevel: mt.NestingLevel,
		decoder:      mt.decoder,
	}

	return newTree
}

func (mb *Branch) DeepClone() *Branch {
	newContent := make([]Marshalable, len(mb.Content))
	for i := range mb.Content {
		switch item := mb.Content[i].(type) {
		case *Tree:
			newContent[i] = item.DeepClone()
		case *Branch:
			newContent[i] = item.DeepClone()
		case *Leaf:
			newContent[i] = item.DeepClone()
		}
	}

	newBranch := &Branch{
		Content:      newContent,
		Name:         mb.Name,
		FullKey:      mb.FullKey,
		NestingLevel: mb.NestingLevel,
		decoder:      mb.decoder,
	}

	return newBranch
}

func (ml *Leaf) DeepClone() *Leaf {
	var newValue string
	switch oldValue := ml.Value.(type) {
	case string:
		newValue = oldValue
	case rune:
		newValue = string(oldValue)
	case []byte:
		newValue = string(oldValue)
	case nil:
		newValue = ""
	default:
		newValue = fmt.Sprint(oldValue)
	}

	return &Leaf{
		Value:            newValue,
		Name:             ml.Name,
		FullKey:          ml.FullKey,
		NestingLevel:     ml.NestingLevel,
		decoder:          ml.decoder,
		yamlMarshalStyle: ml.yamlMarshalStyle,
	}
}

func (mt *Tree) IsEmpty() bool {
	return len(mt.Order) == 0
}

func (ml *Leaf) SetYamlMarshalStyle(s yaml.Style) {
	ml.yamlMarshalStyle = s
}

type WalkFunc func(*Leaf)

func (mt *Tree) Walk(wf WalkFunc) {
	for _, item := range mt.Content {
		switch curItem := item.(type) {
		case *Leaf:
			wf(curItem)
		case *Tree:
			curItem.Walk(wf)
		case *Branch:
			curItem.Walk(wf)
		}
	}
}

func (mb *Branch) Walk(wf WalkFunc) {
	for _, item := range mb.Content {
		switch curItem := item.(type) {
		case *Leaf:
			wf(curItem)
		case *Tree:
			curItem.Walk(wf)
		case *Branch:
			curItem.Walk(wf)
		}
	}
}

func (mt *Tree) clearValues() {
	if len(mt.Content) > 0 {
		mt.Content = make(map[string]Marshalable)
	}
	if len(mt.Order) > 0 {
		mt.Order = nil
	}
}

func (mb *Branch) clearValues() {
	if len(mb.Content) > 0 {
		mb.Content = nil
	}
}

func (ml *Leaf) clearValues() {
	if ml.Value != nil {
		ml.Value = nil
	}
}
