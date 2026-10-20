package caller

import (
	"context"
	"sync"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/types"
)

type Caller interface {
	Get(ctx context.Context, id types.UUID) (coremodel.Model, error)
	GetJoined(ctx context.Context, id types.UUID) (coremodel.Model, error)
	FindFirst(ctx context.Context, options *model.Options) (coremodel.Model, error)
	FindFirstJoined(ctx context.Context, options *model.Options) (coremodel.Model, error)
	FindAll(ctx context.Context, options *model.Options) ([]coremodel.Model, error)
	FindAllJoined(ctx context.Context, options *model.Options) ([]coremodel.Model, error)
	New() any
	NewSlice() any
	NewSlicePtr() any
}

type callerRegistry struct {
	mut     sync.RWMutex
	Callers map[string]Caller
}

var (
	registry *callerRegistry
	once     sync.Once
)

func Registry() *callerRegistry {
	once.Do(func() {
		registry = &callerRegistry{
			Callers: make(map[string]Caller),
		}
	})
	return registry
}

func (r *callerRegistry) Register(name string, caller Caller) {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.Callers[name] = caller
}

func (r *callerRegistry) Get(name string) Caller {
	r.mut.RLock()
	defer r.mut.RUnlock()
	return r.Callers[name]
}

func (r *callerRegistry) GetAll() map[string]Caller {
	r.mut.RLock()
	defer r.mut.RUnlock()

	copyMap := make(map[string]Caller, len(r.Callers))
	for k, v := range r.Callers {
		copyMap[k] = v
	}
	return copyMap
}
