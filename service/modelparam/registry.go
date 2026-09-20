package modelparam

import "sync/atomic"

type Registry struct{ current atomic.Pointer[Snapshot] }

var defaultRegistry atomic.Pointer[Registry]

func NewRegistry(snapshot *Snapshot) *Registry { r := &Registry{}; r.current.Store(snapshot); return r }
func (r *Registry) Snapshot() *Snapshot {
	if r == nil {
		return nil
	}
	return r.current.Load()
}
func (r *Registry) Replace(snapshot *Snapshot) {
	if r != nil && snapshot != nil {
		r.current.Store(snapshot)
	}
}
func SetDefaultRegistry(registry *Registry) { defaultRegistry.Store(registry) }
func DefaultRegistry() *Registry            { return defaultRegistry.Load() }
