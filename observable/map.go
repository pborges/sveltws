package observable

import (
	"sync"
)

type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

type Map[K comparable, V any] struct {
	db   map[K]*Value[V]
	last Value[Entry[K, V]]
	mu   sync.RWMutex
}

func (m *Map[K, V]) Get(key K) (v V, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var val *Value[V]
	if val, ok = m.db[key]; ok && val != nil {
		v = val.current
	}
	return
}

func (m *Map[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.db == nil {
		m.db = map[K]*Value[V]{}
	}
	if _, ok := m.db[key]; !ok {
		m.db[key] = &Value[V]{}
	}
	m.db[key].Set(value)
	m.last.Set(Entry[K, V]{
		Key:   key,
		Value: value,
	})
}

func (m *Map[K, V]) Delete(key K) {
	var v V
	m.Set(key, v)
}

func (m *Map[K, V]) Subscribe(key K) *Subscription[V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.db[key]; !ok {
		m.db[key] = &Value[V]{}
	}
	return m.db[key].Subscribe()
}

func (m *Map[K, V]) SubscribeAll() *Subscription[Entry[K, V]] {
	return m.last.Subscribe()
}

func (m *Map[K, V]) Snapshot() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := map[K]V{}
	for k, v := range m.db {
		if v != nil {
			res[k] = v.current
		}
	}
	return res
}
