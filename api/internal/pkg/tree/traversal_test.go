// Package tree_test provides unit tests for tree traversal utilities.
package tree_test

import (
	"testing"

	"github.com/sed-evaluacion-desempeno/api/internal/pkg/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNode is a test implementation of tree.Node.
type mockNode struct {
	id       string
	parentID string
	children []tree.Node
}

func (m *mockNode) GetID() string       { return m.id }
func (m *mockNode) GetParentID() string { return m.parentID }
func (m *mockNode) GetChildren() []tree.Node {
	if m.children == nil {
		return []tree.Node{}
	}
	return m.children
}

func TestFlatten_SingleLevel(t *testing.T) {
	t.Parallel()

	root := &mockNode{
		id: "1",
		children: []tree.Node{
			&mockNode{id: "2", parentID: "1"},
			&mockNode{id: "3", parentID: "1"},
		},
	}

	flat := tree.Flatten(root, 0)

	require.Len(t, flat, 3)
	assert.Equal(t, "1", flat[0].ID)
	assert.Equal(t, "", flat[0].ParentID)
	assert.Equal(t, 0, flat[0].Depth)
	assert.Equal(t, "1", flat[0].Path)

	assert.Equal(t, "2", flat[1].ID)
	assert.Equal(t, "1", flat[1].ParentID)
	assert.Equal(t, 1, flat[1].Depth)
	assert.Equal(t, "1.2", flat[1].Path)

	assert.Equal(t, "3", flat[2].ID)
	assert.Equal(t, "1", flat[2].ParentID)
	assert.Equal(t, 1, flat[2].Depth)
	assert.Equal(t, "1.3", flat[2].Path)
}

func TestFlatten_DeepTree(t *testing.T) {
	t.Parallel()

	root := &mockNode{
		id: "1",
		children: []tree.Node{
			&mockNode{
				id:       "2",
				parentID: "1",
				children: []tree.Node{
					&mockNode{
						id:       "4",
						parentID: "2",
						children: []tree.Node{
							&mockNode{id: "5", parentID: "4"},
						},
					},
				},
			},
			&mockNode{id: "3", parentID: "1"},
		},
	}

	flat := tree.Flatten(root, 0)

	require.Len(t, flat, 5)
	assert.Equal(t, 0, flat[0].Depth) // 1
	assert.Equal(t, 1, flat[1].Depth) // 2
	assert.Equal(t, 2, flat[2].Depth) // 4
	assert.Equal(t, 3, flat[3].Depth) // 5
	assert.Equal(t, 1, flat[4].Depth) // 3
}

func TestToNested_FlatToTree(t *testing.T) {
	t.Parallel()

	flat := []tree.FlatNode{
		{ID: "1", ParentID: "", Depth: 0, Path: "1"},
		{ID: "2", ParentID: "1", Depth: 1, Path: "1.2"},
		{ID: "3", ParentID: "1", Depth: 1, Path: "1.3"},
		{ID: "4", ParentID: "2", Depth: 2, Path: "1.2.4"},
		{ID: "5", ParentID: "2", Depth: 2, Path: "1.2.5"},
	}

	root := tree.ToNested(flat)
	require.NotNil(t, root)
	assert.Equal(t, "1", root.ID)
	assert.Empty(t, root.ParentID)
	require.Len(t, root.Children, 2)

	// First child: 2, with its own children 4 and 5
	child2 := root.Children[0]
	assert.Equal(t, "2", child2.ID)
	require.Len(t, child2.Children, 2)
	assert.Equal(t, "4", child2.Children[0].ID)
	assert.Equal(t, "5", child2.Children[1].ID)

	// Second child: 3, leaf
	child3 := root.Children[1]
	assert.Equal(t, "3", child3.ID)
	assert.Empty(t, child3.Children)
}

func TestToNested_Empty(t *testing.T) {
	t.Parallel()
	assert.Nil(t, tree.ToNested([]tree.FlatNode{}))
}

func TestFilterDepth(t *testing.T) {
	t.Parallel()

	nodes := []tree.FlatNode{
		{ID: "1", Depth: 0},
		{ID: "2", Depth: 1},
		{ID: "3", Depth: 1},
		{ID: "4", Depth: 2},
		{ID: "5", Depth: 3},
	}

	t.Run("maxDepth zero returns all", func(t *testing.T) {
		filtered := tree.FilterDepth(nodes, 0)
		assert.Len(t, filtered, 5)
	})

	t.Run("maxDepth negative returns all", func(t *testing.T) {
		filtered := tree.FilterDepth(nodes, -1)
		assert.Len(t, filtered, 5)
	})

	t.Run("maxDepth 2 keeps depths 0,1,2", func(t *testing.T) {
		filtered := tree.FilterDepth(nodes, 2)
		require.Len(t, filtered, 4)
		for _, n := range filtered {
			assert.LessOrEqual(t, n.Depth, 2)
		}
	})

	t.Run("maxDepth 1 keeps depths 0,1", func(t *testing.T) {
		filtered := tree.FilterDepth(nodes, 1)
		require.Len(t, filtered, 3)
		for _, n := range filtered {
			assert.LessOrEqual(t, n.Depth, 1)
		}
	})
}

func TestPathString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		parentPath string
		nodeID     string
		want       string
	}{
		{"empty parent", "", "abc", "abc"},
		{"simple join", "1.2", "3", "1.2.3"},
		{"hyphen replacement", "", "a-b-c", "a_b_c"},
		{"single segment", "root", "child", "root.child"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tree.PathString(tt.parentPath, tt.nodeID))
		})
	}
}

func TestNlevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want int
	}{
		{"", 0},
		{"1", 1},
		{"1.2", 2},
		{"1.2.3.4.5", 5},
		{"a.b.c", 3},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, tree.Nlevel(tt.path))
		})
	}
}

func TestCycleDetection(t *testing.T) {
	t.Parallel()

	t.Run("no cycle in valid tree", func(t *testing.T) {
		nodes := []tree.FlatNode{
			{ID: "1", ParentID: ""},
			{ID: "2", ParentID: "1"},
			{ID: "3", ParentID: "2"},
		}
		assert.False(t, detectCycle(nodes), "valid tree should have no cycle")
	})

	t.Run("self-referencing cycle", func(t *testing.T) {
		nodes := []tree.FlatNode{
			{ID: "1", ParentID: ""},
			{ID: "2", ParentID: "2"}, // self-reference
		}
		assert.True(t, detectCycle(nodes), "self-reference is a cycle")
	})

	t.Run("mutual cycle", func(t *testing.T) {
		nodes := []tree.FlatNode{
			{ID: "1", ParentID: "2"},
			{ID: "2", ParentID: "1"},
		}
		assert.True(t, detectCycle(nodes), "mutual reference is a cycle")
	})

	t.Run("deep cycle", func(t *testing.T) {
		nodes := []tree.FlatNode{
			{ID: "1", ParentID: ""},
			{ID: "2", ParentID: "1"},
			{ID: "3", ParentID: "2"},
			{ID: "4", ParentID: "3"},
			{ID: "1", ParentID: "4"}, // 1 is now a child of 4, creating cycle 1->2->3->4->1
		}
		assert.True(t, detectCycle(nodes), "deep cycle should be detected")
	})
}

// detectCycle checks if a set of flat nodes forms a cycle via parent references.
func detectCycle(nodes []tree.FlatNode) bool {
	byID := make(map[string]string, len(nodes))
	for _, n := range nodes {
		byID[n.ID] = n.ParentID
	}

	for id := range byID {
		visited := make(map[string]bool)
		cur := id
		for cur != "" {
			if visited[cur] {
				return true
			}
			visited[cur] = true
			parent, ok := byID[cur]
			if !ok {
				break
			}
			cur = parent
		}
	}
	return false
}

func TestIsDescendantOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		descendant string
		ancestor   string
		want       bool
	}{
		{"1.2.3", "1.2", true},
		{"1.2.3", "1", true},
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "1.3", false},
		{"1.2.3", "", true},
		{"1.2", "1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.descendant+"_vs_"+tt.ancestor, func(t *testing.T) {
			assert.Equal(t, tt.want, tree.IsDescendantOf(tt.descendant, tt.ancestor))
		})
	}
}

func TestIsAncestorOf(t *testing.T) {
	t.Parallel()
	assert.True(t, tree.IsAncestorOf("1.2", "1.2.3"))
	assert.False(t, tree.IsAncestorOf("1.2.3", "1.2"))
}

func TestParsePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want []string
	}{
		{"1.2.3", []string{"1", "2", "3"}},
		{"a", []string{"a"}},
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, tree.ParsePath(tt.path))
		})
	}
}

func TestSubpath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path   string
		offset int
		want   string
		err    bool
	}{
		{"1.2.3.4", 0, "1.2.3.4", false},
		{"1.2.3.4", 1, "2.3.4", false},
		{"1.2.3.4", 2, "3.4", false},
		{"1.2.3.4", 3, "4", false},
		{"1.2.3.4", 4, "", true},
		{"1.2.3.4", -1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+string(rune('0'+tt.offset)), func(t *testing.T) {
			got, err := tree.Subpath(tt.path, tt.offset)
			if tt.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
