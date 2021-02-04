package tree

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Marshalable interface {
	MarshalYAML() (interface{}, error)
	UnmarshalYAML(value *yaml.Node) error
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(raw []byte) error
}

type Tree struct {
	Content map[string]Marshalable
	Name    string
	Order   []string
	fullKey string
	decoder *json.Decoder
}

type Branch struct {
	Content []Marshalable
	Name    string
	fullKey string
	decoder *json.Decoder
}

type Leaf struct {
	Value   interface{}
	Name    string
	FullKey string
	decoder *json.Decoder
}

type Path []string

func New() *Tree {
	return &Tree{
		Content: make(map[string]Marshalable),
	}
}

func NewSubTree(name, prefix string) *Tree {
	return &Tree{
		Content: make(map[string]Marshalable),
		Name:    name,
		fullKey: MakeFullKey(prefix, name),
	}
}

func NewBranch(name, prefix string) *Branch {
	return &Branch{
		Content: []Marshalable{},
		Name:    name,
		fullKey: MakeFullKey(prefix, name),
	}
}

func NewLeaf(name, prefix string) *Leaf {
	return &Leaf{
		Name:    name,
		FullKey: MakeFullKey(prefix, name),
	}
}

func (mt *Tree) Get(path Path) (*Leaf, error) {
	if len(path) < 1 {
		return nil, fmt.Errorf("path is empty")
	}
	currentLevel := mt.Content
	for i, breadcrumb := range path {
		if _, ok := currentLevel[breadcrumb]; !ok {
			return nil, fmt.Errorf("path %v is incorrect", path)
		}
		switch nextLevel := currentLevel[breadcrumb].(type) {
		case *Tree:
			currentLevel = nextLevel.Content
			continue
		case *Leaf:
			if i != len(path)-1 {
				return nil, fmt.Errorf("path %v is too long", path)
			}
			return nextLevel, nil
		default:
			return nil, fmt.Errorf("tree value type %T is incorrect", nextLevel)
		}
	}

	return nil, ErrorNotFound
}

func (mt *Tree) AddDirectly(name string, value Marshalable) {
	mt.Order = append(mt.Order, name)
	mt.Content[name] = value
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
