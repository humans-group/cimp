package tree

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func (mt *Tree) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("yaml node should have Mapping kind for unmarshal, not %v", node.Kind)
	}
	mt.clearValues()

	// Every even item for name only, odd items - for values
	for i := 0; i < len(node.Content); i += 2 {
		curKey := node.Content[i].Value
		curNode := node.Content[i+1]

		switch curNode.Kind {
		case yaml.ScalarNode:
			leaf := NewLeaf(curKey, mt.FullKey)
			if err := leaf.UnmarshalYAML(curNode); err != nil {
				return fmt.Errorf("unmarshal leaf %q: %w", curKey, err)
			}
			mt.AddOrReplaceDirectly(curKey, leaf)
		case yaml.MappingNode:
			childTree := NewSubTree(curKey, mt.FullKey)
			if err := childTree.UnmarshalYAML(curNode); err != nil {
				return fmt.Errorf("unmarshal sub-tree %q: %w", curKey, err)
			}
			mt.AddOrReplaceDirectly(curKey, childTree)
		case yaml.SequenceNode:
			branch := NewBranch(curKey, mt.FullKey)
			if err := branch.UnmarshalYAML(curNode); err != nil {
				return fmt.Errorf("unmarshal branch %q: %w", curKey, err)
			}
			mt.AddOrReplaceDirectly(curKey, branch)
		default:
			return fmt.Errorf("unprocessable content type of %q: %v", curKey, curNode.Kind)
		}
	}

	return nil
}

func (mb *Branch) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("yaml node should have Sequence kind for unmarshal as branch, not %v", node.Kind)
	}
	mb.clearValues()

	for i, curNode := range node.Content {
		curKey := fmt.Sprint(i)

		switch curNode.Kind {
		case yaml.ScalarNode:
			leaf := NewLeaf(curKey, mb.FullKey)
			leaf.Value = curNode.Value
			mb.Add(leaf)
		case yaml.MappingNode:
			tree := NewSubTree(curKey, mb.FullKey)
			if err := tree.UnmarshalYAML(curNode); err != nil {
				return fmt.Errorf("unmarshal %q: %w", curKey, err)
			}
			mb.Add(tree)
		case yaml.SequenceNode:
			childBranch := NewBranch(curKey, mb.FullKey)
			if err := childBranch.UnmarshalYAML(curNode); err != nil {
				return fmt.Errorf("unamrshal child branch #%s: %w", curKey, err)
			}
			mb.Add(childBranch)
		default:
			return fmt.Errorf("unprocessable content type of %q: %v", curKey, curNode.Kind)
		}
	}

	return nil
}

func (mt *Tree) MarshalYAML() (interface{}, error) {
	var content []*yaml.Node
	for _, leafName := range mt.Order {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode}

		if err := keyNode.Encode(leafName); err != nil {
			return nil, fmt.Errorf("encode node %q: %w", leafName, err)
		}

		marshaled, err := mt.Content[leafName].MarshalYAML()
		if err != nil {
			return nil, fmt.Errorf("convert %q to YAML-node: %w", leafName, err)
		}
		leafNode, ok := marshaled.(*yaml.Node)
		if !ok {
			return nil, fmt.Errorf("MarshalYAML return %T insread of yaml.Node", leafNode)
		}
		content = append(content, keyNode, leafNode)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Value:   mt.Name,
		Content: content,
	}, nil
}

func (ml *Leaf) UnmarshalYAML(node *yaml.Node) error {
	ml.clearValues()
	switch node.Kind {
	case yaml.ScalarNode:
		ml.Value = node.Value
	case yaml.SequenceNode:
		leafValue, err := yaml.Marshal(node)
		if err != nil {
			return fmt.Errorf("marshal sequence %q: %w", ml.FullKey, err)
		}
		ml.Value = string(leafValue)
	default:
		return fmt.Errorf("unprocessable content type `%v` for leaf %q", node.Kind, ml.FullKey)
	}

	return nil
}

func (mb *Branch) MarshalYAML() (interface{}, error) {
	var nodeContent []*yaml.Node
	for i, item := range mb.Content {
		marshaled, err := item.MarshalYAML()
		if err != nil {
			return nil, fmt.Errorf("convert #%d item to YAML-node: %w", i, err)
		}
		childNode, ok := marshaled.(*yaml.Node)
		if !ok {
			return nil, fmt.Errorf("MarshalYAML return %T insread of yaml.Node", childNode)
		}
		nodeContent = append(nodeContent, childNode)
	}

	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: nodeContent,
	}, nil
}

func (ml *Leaf) MarshalYAML() (interface{}, error) {
	if ml.yamlMarshalStyle == 0 {
		ml.yamlMarshalStyle = yaml.TaggedStyle
	}

	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Style: ml.yamlMarshalStyle,
		Value: fmt.Sprint(ml.Value),
	}, nil
}
