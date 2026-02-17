package obs

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// TODO: TRACKER INTERFACE WITH ALL METHODS TO IMPLEMENT
type ITracker interface {
	Done()
	MarkActive()
	MarkIdle()
}

type trackerKey struct{}

var registry sync.Map
var idCounter atomic.Uint64
var enabled atomic.Bool

func Enable() {
	enabled.Store(true)
}

func Disable() {
	enabled.Store(false)
}

type State int32

const (
	StateRunning State = iota
	StateIdle
	StateCompleted
)

func (s State) String() string {
	return [...]string{"running", "idle", "completed"}[s]
}

func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *State) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	switch str {
	case "running":
		*s = StateRunning
	case "idle":
		*s = StateIdle
	case "completed":
		*s = StateCompleted
	default:
		*s = StateRunning
	}

	return nil
}

type Tracker struct {
	Id       uint64
	Name     string
	ParentId uint64

	startedAt  time.Time
	endedAt    atomic.Int64
	state      atomic.Int32
	lastActive atomic.Int64
}

func Listen(addr string) error {
	if !enabled.Load() {
		return nil
	}

	http.HandleFunc("/debug/obs", func(w http.ResponseWriter, r *http.Request) {
		snapshots := Snapshot()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(snapshots)
	})

	go func() {
		log.Printf("obs listening on port %s\n", addr)
		_ = http.ListenAndServe(addr, nil)
	}()

	return nil
}

// With creates a new tracker and attaches it to context
func With(ctx context.Context, name string) context.Context {
	if !enabled.Load() {
		return ctx
	}

	parent := FromContext(ctx)

	id := idCounter.Add(1)

	t := &Tracker{
		Id:        id,
		Name:      name,
		ParentId:  0,
		startedAt: time.Now(),
	}

	if parent != nil {
		t.ParentId = parent.Id
	}

	t.state.Store(int32(StateRunning))
	t.lastActive.Store(time.Now().Unix())

	registry.Store(id, t)

	return context.WithValue(ctx, trackerKey{}, t)
}

func FromContext(ctx context.Context) *Tracker {
	if ctx == nil {
		return nil
	}

	val := ctx.Value(trackerKey{})

	if val == nil {
		return nil
	}

	tracker, ok := val.(*Tracker)
	if !ok {
		return nil
	}

	return tracker
}

func Snapshot() []TrackerSnapshot {
	if !enabled.Load() {
		return nil
	}

	var snapshots []TrackerSnapshot

	registry.Range(func(key, value any) bool {
		t := value.(*Tracker)
		snapshots = append(snapshots, TrackerSnapshot{
			Id:          t.Id,
			Name:        t.Name,
			ParentId:    t.ParentId,
			StartTimeAt: t.startedAt,
			State:       State(t.state.Load()),
			Duration:    t.duration(),
		})

		return true
	})

	return snapshots
}

type TrackerSnapshot struct {
	Id          uint64        `json:"id"`
	Name        string        `json:"name"`
	ParentId    uint64        `json:"parent_id"`
	StartTimeAt time.Time     `json:"start_time_at"`
	State       State         `json:"state"`
	Duration    time.Duration `json:"duration"`
}

func (t *Tracker) Done() {
	if t == nil {
		return
	}

	t.state.Store(int32(StateCompleted))
	t.endedAt.Store(time.Now().UnixNano())
}

func (t *Tracker) MarkActive() {
	if t == nil {
		return
	}

	t.state.Store(int32(StateRunning))
	t.lastActive.Store(time.Now().UnixNano())
}

func (t *Tracker) MarkIdle() {
	if t == nil {
		return
	}

	t.state.Store(int32(StateIdle))
}

func (t *Tracker) duration() time.Duration {
	end := t.endedAt.Load()

	if end == 0 {
		return time.Since(t.startedAt)
	}

	return time.Duration(end - t.startedAt.UnixNano())
}

func (t *Tracker) Emit(event string) {}
