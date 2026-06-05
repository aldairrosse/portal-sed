// Package tree provides generic tree-traversal utilities for the org hierarchy.
// Supports flattening, nesting, depth filtering, and ltree path operations.
package tree

import (
	"fmt"
	"strings"
)

// Node is the minimal interface required for tree operations.
// Implementations must provide ID, ParentID, and Children accessors.
type Node interface {
	GetID() string
	GetParentID() string
	GetChildren() []Node
}

// FlatNode is a flat representation with an ID and ParentID.
type FlatNode struct {
	ID       string `json:"id"`
	ParentID string `json:"parentId,omitempty"`
	Depth    int    `json:"depth"`
	Path     string `json:"path"`
}

// Flatten converts a nested tree (root node with children) into a flat slice
// ordered by depth-first traversal. Each node gets its depth computed.
func Flatten(root Node, depth int) []FlatNode {
	result := make([]FlatNode, 0)
	flattenRecursive(root, "", depth, &result)
	return result
}

func flattenRecursive(node Node, parentPath string, depth int, result *[]FlatNode) {
	path := PathString(parentPath, node.GetID())
	*result = append(*result, FlatNode{
		ID:       node.GetID(),
		ParentID: node.GetParentID(),
		Depth:    depth,
		Path:     path,
	})
	children := node.GetChildren()
	for _, child := range children {
		flattenRecursive(child, path, depth+1, result)
	}
}

// NestedNode represents a node in a nested tree structure.
type NestedNode struct {
	ID       string        `json:"id"`
	ParentID string        `json:"parentId,omitempty"`
	Children []*NestedNode `json:"children,omitempty"`
}

// ToNested builds a nested tree from a flat slice of nodes ordered by path.
// Returns the root node with children populated recursively.
// The input must be ordered by path (e.g., "1", "1.1", "1.2", "1.2.1", "2", ...).
func ToNested(nodes []FlatNode) *NestedNode {
	if len(nodes) == 0 {
		return nil
	}

	byID := make(map[string]*NestedNode, len(nodes))
	for _, n := range nodes {
		byID[n.ID] = &NestedNode{
			ID:       n.ID,
			ParentID: n.ParentID,
			Children: make([]*NestedNode, 0),
		}
	}

	var root *NestedNode
	for _, n := range nodes {
		nn := byID[n.ID]
		if n.ParentID == "" {
			root = nn
		} else if parent, ok := byID[n.ParentID]; ok {
			parent.Children = append(parent.Children, nn)
		}
	}

	return root
}

// FilterDepth returns only nodes where depth <= maxDepth.
// If maxDepth <= 0, returns all nodes unchanged (depth 0 = root only with no filtering).
// With maxDepth > 0, only nodes at depth <= maxDepth are kept.
func FilterDepth(nodes []FlatNode, maxDepth int) []FlatNode {
	if maxDepth <= 0 {
		return nodes
	}
	filtered := make([]FlatNode, 0, len(nodes))
	for _, n := range nodes {
		if n.Depth <= maxDepth {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

// PathString concatenates a parent path and a node ID to produce an ltree-style path.
// Parent path "1.2" + node ID "3" → "1.2.3"
// Empty parent path + node ID "1" → "1"
func PathString(parentPath, nodeID string) string {
	// Replace hyphens in UUIDs since ltree segments cannot contain hyphens
	cleanID := strings.ReplaceAll(nodeID, "-", "_")
	if parentPath == "" {
		return cleanID
	}
	return parentPath + "." + cleanID
}

// ParsePath splits an ltree path string into its component segments.
// "1.2.3" → ["1", "2", "3"]
func ParsePath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}

// IsDescendantOf returns true if descendantPath is a descendant of (or equal to) ancestorPath.
// In ltree terms: descendantPath <@ ancestorPath
// "1.2.3" is a descendant of "1.2" → true
// "1.3" is a descendant of "1.2" → false
func IsDescendantOf(descendantPath, ancestorPath string) bool {
	if ancestorPath == "" {
		return true
	}
	if descendantPath == ancestorPath {
		return true
	}
	return strings.HasPrefix(descendantPath, ancestorPath+".")
}

// IsAncestorOf returns true if ancestorPath is an ancestor of (or equal to) descendantPath.
// In ltree terms: ancestorPath @> descendantPath
func IsAncestorOf(ancestorPath, descendantPath string) bool {
	return IsDescendantOf(descendantPath, ancestorPath)
}

// Nlevel returns the number of segments in an ltree path.
// "1.2.3" → 3, "" → 0
func Nlevel(path string) int {
	if path == "" {
		return 0
	}
	segments := strings.Split(path, ".")
	// Filter empty strings (shouldn't happen with valid paths, but guard)
	count := 0
	for _, s := range segments {
		if s != "" {
			count++
		}
	}
	return count
}

// Subpath returns the portion of path starting at offset.
// Subpath("1.2.3.4", 2) → "3.4"
// Subpath("1.2.3.4", 1) → "2.3.4"
func Subpath(path string, offset int) (string, error) {
	segments := ParsePath(path)
	if offset < 0 || offset >= len(segments) {
		return "", fmt.Errorf("tree: subpath offset %d out of range for path %q (len=%d)", offset, path, len(segments))
	}
	return strings.Join(segments[offset:], "."), nil
}
