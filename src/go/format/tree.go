package format

import (
	"fmt"
	"strings"
)

type TreeNode struct {
	isRoot   bool
	isLeaf   bool
	Value    string
	Children []*TreeNode
}

func NewTree(rootNodeValue string) *TreeNode {
	return &TreeNode{
		isRoot:   true,
		Value:    rootNodeValue,
		Children: []*TreeNode{},
	}
}

func (n *TreeNode) NewChild(value string) *TreeNode {
	child := &TreeNode{
		Value:    value,
		Children: []*TreeNode{},
	}
	n.AddChild(child)
	return child
}

func (n *TreeNode) AddChild(child *TreeNode) {
	n.isLeaf = false
	n.Children = append(n.Children, child)
}

func (n *TreeNode) AddLeaf(leafValue string) {
	n.Children = append(n.Children, &TreeNode{
		Value:    leafValue,
		Children: []*TreeNode{},
		isLeaf:   true,
	})
}

func (n *TreeNode) String() string {
	maxColonPos := n.maxColonPosition()
	return n.string("", true, maxColonPos)
}

func (n *TreeNode) string(prefix string, isTail bool, maxColonPos int) string {
	var sb strings.Builder

	// If it's a leaf, pad the value for alignment
	if n.isLeaf {
		colonIndex := strings.Index(n.Value, ":")
		if colonIndex == -1 {
			// If no colon is found, just print the value
			sb.WriteString(fmt.Sprintf("%s%s\n", prefix+leafConnector(isTail), n.Value))
		} else {
			// Align the colon position
			key := n.Value[:colonIndex+1]
			value := n.Value[colonIndex+1:]
			paddedKey := fmt.Sprintf("%-*s", maxColonPos, key) // Align the key part
			sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix+leafConnector(isTail), paddedKey, value))
		}
	} else {
		// Non-leaf node, just print the value
		sb.WriteString(fmt.Sprintf("%s%s\n", prefix+nonLeafConnector(isTail, n.isRoot), n.Value))
	}

	// Process the children
	for i, child := range n.Children {
		childPrefix := prefix
		if isTail {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
		// Check if the child is the last node
		isLast := i == len(n.Children)-1
		sb.WriteString(child.string(childPrefix, isLast, maxColonPos))
	}

	return sb.String()
}

func (n *TreeNode) maxColonPosition() int {
	maxColonPos := 0
	if n.isLeaf {
		colonIndex := strings.Index(n.Value, ":")
		if colonIndex != -1 && colonIndex+1 > maxColonPos {
			maxColonPos = colonIndex + 1
		}
	}
	for _, child := range n.Children {
		childColonPos := child.maxColonPosition()
		if childColonPos > maxColonPos {
			maxColonPos = childColonPos
		}
	}
	return maxColonPos
}

func leafConnector(isTail bool) string {
	if isTail {
		return "└── "
	}
	return "├── "
}

func nonLeafConnector(isTail bool, isRoot bool) string {
	if isRoot {
		return ""
	}

	if isTail {
		return "└── "
	}
	return "├── "
}
