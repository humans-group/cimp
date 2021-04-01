package tree

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Marshalable interface {
	MarshalYAML() (interface{}, error)
	UnmarshalYAML(value *yaml.Node) error
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(raw []byte) error
	GetByFullKey(string) (Marshalable, error)
	GetFullKey() string
	GetName() string
	GetNestingLevel() int
	Delete(string) error
	IsEmpty() bool
}

type Tree struct {
	Content      map[string]Marshalable
	Name         string
	Order        []string
	FullKey      string
	nestingLevel int
	decoder      *json.Decoder
}

type Branch struct {
	Content      []Marshalable
	Name         string
	FullKey      string
	nestingLevel int
	decoder      *json.Decoder
}

type Leaf struct {
	Value            interface{}
	Name             string
	FullKey          string
	decoder          *json.Decoder
	nestingLevel     int
	yamlMarshalStyle yaml.Style
}

type Path []string

func New() *Tree {
	return &Tree{
		Content: make(map[string]Marshalable),
	}
}

func NewSubTree(name, parentFullKey string) *Tree {
	return &Tree{
		Content:      make(map[string]Marshalable),
		Name:         name,
		FullKey:      MakeFullKey(parentFullKey, name),
		nestingLevel: initNestingLevel(parentFullKey),
	}
}

func NewBranch(name, parentFullKey string) *Branch {
	return &Branch{
		Content:      []Marshalable{},
		Name:         name,
		FullKey:      MakeFullKey(parentFullKey, name),
		nestingLevel: initNestingLevel(parentFullKey),
	}
}

func NewLeaf(name, parentFullKey string) *Leaf {
	return &Leaf{
		Name:         name,
		FullKey:      MakeFullKey(parentFullKey, name),
		nestingLevel: initNestingLevel(parentFullKey),
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

func (mt *Tree) GetByFullKey(fullKey string) (Marshalable, error) {
	if mt.FullKey == fullKey {
		return mt, nil
	}

	relativeKey := strings.TrimPrefix(fullKey, mt.FullKey+sep)
	if len(relativeKey) > len(fullKey) {
		return nil, ErrorNotFound
	}

	sepIdx := strings.IndexAny(relativeKey, sep)
	// If separator is not found - it should be final item
	isFinal := sepIdx == -1
	curRelativeKey := relativeKey
	if !isFinal {
		curRelativeKey = relativeKey[:sepIdx]
	}
	searchKey := curRelativeKey
	if len(mt.FullKey) != 0 {
		searchKey = mt.FullKey + sep + searchKey
	}

	for _, v := range mt.Content {
		if v.GetFullKey() == searchKey {
			return v.GetByFullKey(fullKey)
		}
	}

	return nil, ErrorNotFound
}

func (mb *Branch) GetByFullKey(fullKey string) (Marshalable, error) {
	if mb.FullKey == fullKey {
		return mb, nil
	}

	relativeKey := strings.TrimPrefix(fullKey, mb.FullKey+sep)
	if len(relativeKey) > len(fullKey) {
		return nil, ErrorNotFound
	}

	sepIdx := strings.IndexAny(relativeKey, sep)
	// If separator is not found - it should be final item
	isFinal := sepIdx == -1
	curRelativeKey := relativeKey
	if !isFinal {
		curRelativeKey = relativeKey[:sepIdx]
	}
	searchIdx, err := strconv.Atoi(curRelativeKey)
	if err != nil {
		return nil, fmt.Errorf("wrong fullKey format, %q is branch: %w", mb.FullKey, err)
	}

	if len(mb.Content)-1 >= searchIdx {
		return mb.Content[searchIdx].GetByFullKey(fullKey)
	}

	return nil, ErrorNotFound
}

func (ml *Leaf) GetByFullKey(fullKey string) (Marshalable, error) {
	if ml.FullKey == fullKey {
		return ml, nil
	}

	return nil, ErrorNotFound
}

func (mt *Tree) Delete(fullKey string) error {
	if mt.FullKey == fullKey {
		return fmt.Errorf("you can't delete a tree from itself")
	}

	relativeKey := strings.TrimPrefix(fullKey, mt.FullKey+sep)
	if len(relativeKey) > len(fullKey) {
		return ErrorNotFound
	}

	sepIdx := strings.IndexAny(relativeKey, sep)
	isFinal := sepIdx == -1
	curRelativeKey := relativeKey
	if !isFinal {
		curRelativeKey = relativeKey[:sepIdx]
	}
	searchKey := curRelativeKey
	if len(mt.FullKey) != 0 {
		searchKey = mt.FullKey + sep + searchKey
	}

	var deletedName string
	for _, v := range mt.Content {
		if v.GetFullKey() != searchKey {
			continue
		}
		if !isFinal {
			if err := v.Delete(fullKey); err != nil {
				return err
			}
			if !v.IsEmpty() {
				return nil
			}
		}
		deletedName = v.GetName()
		delete(mt.Content, deletedName)
		break
	}

	if len(deletedName) == 0 {
		return ErrorNotFound
	}

	for i, orderedName := range mt.Order {
		if orderedName == deletedName {
			copy(mt.Order[i:], mt.Order[i+1:])    // Shift a[i+1:] left one index
			mt.Order[len(mt.Order)-1] = ""        // Erase last element (write zero value)
			mt.Order = mt.Order[:len(mt.Order)-1] // Truncate slice
			return nil
		}
	}

	return fmt.Errorf("item %q was deleted, but its name is not found in order slice", deletedName)
}

func (mb *Branch) Delete(fullKey string) error {
	if mb.FullKey == fullKey {
		return fmt.Errorf("you can't delete a branch from itself")
	}

	relativeKey := strings.TrimPrefix(fullKey, mb.FullKey+sep)

	sepIdx := strings.IndexAny(relativeKey, sep)
	isFinal := sepIdx == -1
	curRelativeKey := relativeKey
	if !isFinal {
		curRelativeKey = relativeKey[:sepIdx]
	}
	searchIdx, err := strconv.Atoi(curRelativeKey)
	if err != nil {
		return fmt.Errorf("wrong fullKey format, %q is branch: %w", mb.FullKey, err)
	}

	if len(mb.Content)-1 < searchIdx {
		return ErrorNotFound
	}

	if !isFinal {
		if err := mb.Content[searchIdx].Delete(fullKey); err != nil {
			return err
		}
		if !mb.Content[searchIdx].IsEmpty() {
			return nil
		}
	}

	copy(mb.Content[searchIdx:], mb.Content[searchIdx+1:]) // Shift a[i+1:] left one index
	mb.Content[len(mb.Content)-1] = nil                    // Erase last element (write zero value)
	mb.Content = mb.Content[:len(mb.Content)-1]            // Truncate slice

	return nil
}

func (ml *Leaf) Delete(fullKey string) error {
	if ml.FullKey != fullKey {
		return fmt.Errorf("incorrect key for leaf deletion: %q", fullKey)
	}
	ml.Value = nil

	return nil
}

func (mt *Tree) AddOrReplaceDirectly(name string, value Marshalable) {
	if _, ok := mt.Content[name]; !ok {
		mt.Order = append(mt.Order, name)
	}
	switch item := value.(type) {
	case *Tree:
		item.nestingLevel = mt.nestingLevel + 1
	case *Branch:
		item.nestingLevel = mt.nestingLevel + 1
	case *Leaf:
		item.nestingLevel = mt.nestingLevel + 1
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
		nestingLevel: mt.nestingLevel,
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
		nestingLevel: mt.nestingLevel,
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
		nestingLevel: mb.nestingLevel,
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
		nestingLevel:     ml.nestingLevel,
		decoder:          ml.decoder,
		yamlMarshalStyle: ml.yamlMarshalStyle,
	}
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

func (mb *Branch) GetName() string {
	return mb.Name
}

func (mb *Branch) GetFullKey() string {
	return mb.FullKey
}

func (mb *Branch) GetNestingLevel() int {
	return mb.nestingLevel
}

func (mb *Branch) IsEmpty() bool {
	return len(mb.Content) == 0
}

func (mt *Tree) GetName() string {
	return mt.Name
}

func (mt *Tree) GetFullKey() string {
	return mt.FullKey
}

func (mt *Tree) GetNestingLevel() int {
	return mt.nestingLevel
}

func (mt *Tree) IsEmpty() bool {
	return len(mt.Content) == 0
}

func (ml *Leaf) GetName() string {
	return ml.Name
}

func (ml *Leaf) GetFullKey() string {
	return ml.FullKey
}

func (ml *Leaf) GetNestingLevel() int {
	return ml.nestingLevel
}

func (ml *Leaf) IsEmpty() bool {
	return ml.Value == nil
}

func initNestingLevel(parentFullKey string) int {
	if len(parentFullKey) == 0 {
		return 1
	}

	return strings.Count(parentFullKey, sep) + 2
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
