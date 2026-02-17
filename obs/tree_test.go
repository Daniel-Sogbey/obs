package obs

import "testing"

func TestBuildTreeSimple(t *testing.T) {
	snapshots := []TrackerSnapshot{
		{Id: 1, Name: "root", ParentId: 0},
		{Id: 2, Name: "child1", ParentId: 1},
		{Id: 3, Name: "child2", ParentId: 1},
	}

	roots := BuildTree(snapshots)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}

	root := roots[0]

	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children))
	}
}

func TestBuildTreeOrphan(t *testing.T) {
	snapshots := []TrackerSnapshot{
		{Id: 1, Name: "orphan", ParentId: 99},
	}

	roots := BuildTree(snapshots)

	if len(roots) != 1 {
		t.Fatal("expected orphan to become root")
	}
}

func TestBuildTreeDeepHierarchy(t *testing.T) {
	snapshots := []TrackerSnapshot{
		{Id: 1, Name: "root", ParentId: 0},
		{Id: 2, Name: "level1", ParentId: 1},
		{Id: 3, Name: "level2", ParentId: 2},
	}

	roots := BuildTree(snapshots)

	if len(roots) != 1 {
		t.Fatal("expected one root")
	}

	if len(roots[0].Children) != 1 {
		t.Fatal("expected one child at level1")
	}

	if len(roots[0].Children[0].Children) != 1 {
		t.Fatal("expected one child at level2")
	}
}
