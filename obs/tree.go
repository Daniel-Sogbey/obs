package obs

type TreeNode struct {
	Snapshot TrackerSnapshot
	Children []*TreeNode
}

func BuildTree(snapshots []TrackerSnapshot) []*TreeNode {
	nodeMap := make(map[uint64]*TreeNode)

	for _, snap := range snapshots {
		nodeMap[snap.Id] = &TreeNode{
			Snapshot: snap,
		}
	}

	var roots []*TreeNode

	for _, node := range nodeMap {
		parentId := node.Snapshot.ParentId

		if parentId == 0 {
			roots = append(roots, node)
			continue
		}

		parent, ok := nodeMap[parentId]
		if !ok {
			roots = append(roots, node)
			continue
		}

		parent.Children = append(parent.Children, node)
	}

	return roots
}
