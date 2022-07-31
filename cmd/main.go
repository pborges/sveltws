package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"test/entity"
	"test/entity/person"
	"test/rpc"
	"test/ws"
)

type ctxShim struct {
	*ws.Context
	*rpc.Mux
	subs map[string]entity.Closer
}

func (c ctxShim) Notify(method string, params any) {
	c.WriteNotification(c.Context, rpc.Notification[any]{
		Method: method,
		Params: params,
	})
}

func (c ctxShim) RegisterSubscription(id string, sub entity.Closer) {
	c.subs[id] = sub
}

func main() {
	subs := map[string]entity.Closer{}

	mux := rpc.Mux{}
	mux.Handle("person.get", person.HandleGet)
	mux.Handle("person.reset", person.HandleReset)

	mux.Handle("unsubscribe", func(ctx rpc.Context, req rpc.Request) (any, error) {
		var id string
		if err := json.Unmarshal(req.Params, &id); err != nil {
			return nil, err
		}
		sub, ok := subs[id]
		if ok {
			delete(subs, id)
			sub.Close()
		}
		return ok, nil
	})

	upgrader := ws.Upgrader{
		Log: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
		OnMessage: func(ctx *ws.Context, buf []byte) {
			shim := ctxShim{Context: ctx, Mux: &mux, subs: subs}
			go mux.Dispatch(shim, buf)
		},
	}

	http.HandleFunc("/ws", upgrader.Upgrade)
	fs := http.FileServer(http.Dir("ux/public"))
	http.Handle("/", fs)

	http.ListenAndServe(":8081", nil)
}
