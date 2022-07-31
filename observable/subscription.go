package observable

type Subscription[V any] struct {
	ch chan V
	C  <-chan V
	v  *Value[V]
}

func (s *Subscription[V]) Close() {
	if s.v == nil {
		return
	}
	s.v.mu.Lock()
	defer s.v.mu.Unlock()
	for i, sub := range s.v.subs {
		if sub == s {
			s.v.subs[i] = s.v.subs[len(s.v.subs)-1]
			s.v.subs = s.v.subs[:len(s.v.subs)-1]
			close(sub.ch)
			return
		}
	}
}
