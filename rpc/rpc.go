package rpc

import (
	"encoding/json"
	"errors"
	"io"
	"sync"
)

type HandlerFunc func(ctx Context, req Request) (any, error)

type Request struct {
	Id     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Response[T any] struct {
	Id     string
	Result T
	Error  error
}

func (r Response[T]) MarshalJSON() ([]byte, error) {
	res := map[string]any{}
	if r.Id != "" {
		res["id"] = r.Id
	}
	if r.Error == nil {
		res["result"] = r.Result
	} else {
		res["error"] = r.Error.Error()
	}
	return json.Marshal(res)
}

type Notification[T any] struct {
	Method string `json:"method"`
	Params T      `json:"params"`
}

type Context interface {
	Notify(method string, param any)
	OnDisconnect(fn func())
}

type WritableContext interface {
	Context
	Write([]byte) (int, error)
}

type Mux struct {
	methods map[string]HandlerFunc
	mu      sync.RWMutex
}

func (m *Mux) Handle(method string, fn HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.methods == nil {
		m.methods = map[string]HandlerFunc{}
	}
	m.methods[method] = fn
}

func (m *Mux) WriteNotification(w io.Writer, msg Notification[any]) error {
	return json.NewEncoder(w).Encode(msg)
}

func (m *Mux) Dispatch(ctx WritableContext, msg []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var req Request
	var res Response[any]

	defer func() {
		res.Id = req.Id
		json.NewEncoder(ctx).Encode(res)
	}()

	if err := json.Unmarshal(msg, &req); err != nil {
		res.Error = errors.New("unable to unmarshal request")
		return
	}
	if req.Id == "" {
		res.Error = errors.New("id cannot be empty")
		return
	}
	if req.Method == "" {
		res.Error = errors.New("method cannot be empty")
		return
	}

	if h, ok := m.methods[req.Method]; ok {
		res.Result, res.Error = h(ctx, req)
	} else {
		res.Error = errors.New("method not found")
	}
}
