package obs

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func setup() {
	Enable()
}

func TestWithAndFromContext(t *testing.T) {
	setup()

	ctx := With(context.Background(), "root")
	tracker := FromContext(ctx)

	if tracker == nil {
		t.Fatal("expected tracker, got nil")
	}

	if tracker.Name != "root" {
		t.Fatalf("expected name root, got %s", tracker.Name)
	}
}

func TestParentChildRelationship(t *testing.T) {
	setup()

	rootCtx := With(context.Background(), "root")
	root := FromContext(rootCtx)

	childCtx := With(rootCtx, "child")
	child := FromContext(childCtx)

	if child.ParentId != root.Id {
		t.Fatalf("expected parent ID %d, got %d", root.Id, child.ParentId)
	}
}

func TestStateTransitions(t *testing.T) {
	setup()

	ctx := With(context.Background(), "worker")
	tk := FromContext(ctx)

	tk.MarkIdle()
	if State(tk.state.Load()) != StateIdle {
		t.Fatal("expected idle state")
	}

	tk.MarkActive()
	if State(tk.state.Load()) != StateRunning {
		t.Fatal("expected running state")
	}

	tk.Done()
	if State(tk.state.Load()) != StateCompleted {
		t.Fatal("expected completed state")
	}
}

func TestSnapshotReturnsData(t *testing.T) {
	setup()

	ctx := With(context.Background(), "snapshot-test")
	tk := FromContext(ctx)
	defer tk.Done()

	snapshots := Snapshot()

	if len(snapshots) == 0 {
		t.Fatal("expected at least one snapshot")
	}
}

func TestDurationAfterCompletion(t *testing.T) {
	setup()

	ctx := With(context.Background(), "duration-test")
	tk := FromContext(ctx)

	time.Sleep(10 * time.Millisecond)
	tk.Done()

	snap := Snapshot()[0]

	if snap.Duration <= 0 {
		t.Fatal("expected positive duration")
	}
}

func TestStateJSONRoundTrip(t *testing.T) {
	state := StateRunning

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}

	var decoded State
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded != StateRunning {
		t.Fatal("expected running after roundtrip")
	}
}
