package observable

import (
	"sync"
	"time"
)

type Value[V any] struct {
	current   V
	subs      []*Subscription[V]
	mu        sync.RWMutex
	createdAt time.Time
}

func (v *Value[V]) CreatedAt() time.Time {
	return v.createdAt
}

func (v *Value[V]) Get() V {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.current
}

func (v *Value[V]) Set(val V) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.current = val
	v.createdAt = time.Now()
	v.fanout()
}

func (v *Value[V]) Clear() {
	var val V
	v.Set(val)
}

func (v *Value[V]) fanout() {
	goodSubs := make([]*Subscription[V], 0, len(v.subs))
	wg := sync.WaitGroup{}
	wg.Add(len(v.subs))
	for _, sub := range v.subs {
		go func(sub *Subscription[V]) {
			defer wg.Done()
			select {
			case sub.ch <- v.current:
				goodSubs = append(goodSubs, sub)
			case <-time.After(200 * time.Millisecond):
				sub.Close()
			}
		}(sub)
	}
	wg.Wait()
	v.subs = goodSubs
}

func (v *Value[V]) Subscribe() *Subscription[V] {
	v.mu.Lock()
	defer v.mu.Unlock()
	sub := &Subscription[V]{
		ch: make(chan V),
	}
	sub.C = sub.ch
	v.subs = append(v.subs, sub)
	return sub
}
