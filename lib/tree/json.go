package tree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

func (mt *Tree) MarshalJSON() ([]byte, error) {
	var raw []byte
	buf := bytes.NewBuffer(raw)
	buf.WriteRune('{')
	end := len(mt.Order) - 1
	for i, name := range mt.Order {
		encodedName, err := json.Marshal(name)
		if err != nil {
			return nil, fmt.Errorf("marshal name %q: %w", name, err)
		}
		buf.Write(encodedName)
		buf.WriteRune(':')
		encodedLeaf, err := json.Marshal(mt.Content[name])
		if err != nil {
			return nil, fmt.Errorf("marshal leaf %q: %w", name, err)
		}
		buf.Write(encodedLeaf)
		if i != end {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune('}')

	return buf.Bytes(), nil
}

func (mb *Branch) MarshalJSON() ([]byte, error) {
	var raw []byte
	buf := bytes.NewBuffer(raw)
	buf.WriteRune('[')
	end := len(mb.Content) - 1
	for i, item := range mb.Content {
		encodedValue, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("marshal item #%d %q: %w", i, item, err)
		}
		buf.Write(encodedValue)
		if i != end {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune(']')

	return buf.Bytes(), nil
}

func (ml *Leaf) MarshalJSON() ([]byte, error) {
	encodedValue, err := json.Marshal(ml.Value)
	if err != nil {
		return nil, fmt.Errorf("marshal leaf value %q: %w", ml.Value, err)
	}

	return encodedValue, nil
}

func (ml *Leaf) UnmarshalJSON(raw []byte) error {
	if ml.decoder != nil {
		// leaf already filled in parent
		ml.decoder = nil
		return nil
	}
	ml.clearValues()

	raw = bytes.TrimSpace(raw)
	ml.decoder = json.NewDecoder(bytes.NewReader(raw))
	defer func() { ml.decoder = nil }()

	token, err := ml.decoder.Token()
	if err != nil {
		return fmt.Errorf("get next token from decoder: %w", err)
	}
	if delimToken, ok := token.(json.Delim); !ok || delimToken != '{' {
		return fmt.Errorf("JSON object must start from '{', not %v", token)
	}

	keyToken, err := ml.decoder.Token()
	if err != nil {
		return fmt.Errorf("get next token: %w", err)
	}
	name, ok := keyToken.(string)
	if !ok {
		return fmt.Errorf("name must be a string, got: %T", keyToken)
	}

	valueToken, err := ml.decoder.Token()
	if err != nil {
		return fmt.Errorf("get next token: %w", err)
	}
	value, ok := valueToken.(string)
	if !ok {
		return fmt.Errorf("value must be a string, got: %T", keyToken)
	}

	closeToken, err := ml.decoder.Token()
	if err != nil {
		return err
	}
	if delim, ok := closeToken.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expect JSON object close with '}'")
	}

	ml.Name = name
	ml.Value = value
	ml.FullKey = MakeFullKey("", name)

	return nil
}

func (mb *Branch) UnmarshalJSON(raw []byte) error {
	mb.clearValues()
	if mb.decoder == nil {
		raw = bytes.TrimSpace(raw)
		dec := json.NewDecoder(bytes.NewReader(raw))

		token, err := dec.Token()
		if err != nil {
			return fmt.Errorf("get next token from decoder: %w", err)
		}

		if delimToken, ok := token.(json.Delim); !ok || delimToken != '[' {
			return fmt.Errorf("JSON array must start from '[', not %v", token)
		}
		mb.decoder = dec
	}
	defer func() { mb.decoder = nil }()

	var i = -1
	var name string
	for mb.decoder.More() {
		i++
		name = strconv.Itoa(i)
		var child Marshalable

		token, err := mb.decoder.Token()
		if err != nil {
			return fmt.Errorf("get next token: %w", err)
		}

		delim, ok := token.(json.Delim)
		if !ok {
			leaf := NewLeaf(name, mb.FullKey)
			leaf.decoder = mb.decoder
			leaf.Value = token
			child = leaf
		} else {
			switch delim {
			case '{':
				childTree := NewSubTree(name, mb.FullKey)
				childTree.decoder = mb.decoder
				child = childTree
			case '[':
				branch := NewBranch(name, mb.FullKey)
				branch.decoder = mb.decoder
				child = branch
			default:
				return fmt.Errorf("got unpredictable token '%v'", token)
			}
		}

		// raw not needed to child item, it have decoder already
		if err := child.UnmarshalJSON(nil); err != nil {
			return fmt.Errorf("unmarshal %q: %w", name, err)
		}
		mb.Content = append(mb.Content, child)
	}

	token, err := mb.decoder.Token() // '}'
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); !ok || delim != ']' {
		return fmt.Errorf("JSON array must be closed with ']'")
	}

	return nil
}

func (mt *Tree) UnmarshalJSON(raw []byte) error {
	mt.clearValues()
	if mt.decoder == nil {
		raw = bytes.TrimSpace(raw)
		dec := json.NewDecoder(bytes.NewReader(raw))

		token, err := dec.Token()
		if err != nil {
			return fmt.Errorf("get next token from decoder: %w", err)
		}

		// must open with a delimToken token '{'
		if delimToken, ok := token.(json.Delim); !ok || delimToken != '{' {
			return fmt.Errorf("expect JSON object open with '{'")
		}
		mt.decoder = dec
	}
	defer func() { mt.decoder = nil }()

	for mt.decoder.More() {
		keyToken, err := mt.decoder.Token()
		if err != nil {
			return fmt.Errorf("get next token: %w", err)
		}

		name, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("name must be a string, got: %T", keyToken)
		}

		var child Marshalable

		token, err := mt.decoder.Token()
		if err != nil {
			return fmt.Errorf("get next token: %w", err)
		}

		delim, ok := token.(json.Delim)
		if !ok {
			leaf := NewLeaf(name, mt.FullKey)
			leaf.decoder = mt.decoder
			leaf.Value = token
			child = leaf
		} else {
			switch delim {
			case '{':
				childTree := NewSubTree(name, mt.FullKey)
				childTree.decoder = mt.decoder
				child = childTree
			case '[':
				branch := NewBranch(name, mt.FullKey)
				branch.decoder = mt.decoder
				child = branch
			default:
				return fmt.Errorf("got unpredictable token '%v'", token)
			}

			// raw not needed to child item, it have decoder already
			if err := child.UnmarshalJSON(nil); err != nil {
				return fmt.Errorf("unmarshal %q: %w", name, err)
			}
		}

		mt.AddOrReplaceDirectly(name, child)
	}

	token, err := mt.decoder.Token() // '}'
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expect JSON object close with '}'")
	}

	return nil
}
