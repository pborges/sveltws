package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"test/observable"
	"test/rpc"
	"test/ws"
	"time"
)

type Person struct {
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Age       int      `json:"age"`
	Children  []Person `json:"children,omitempty"`
}

type Request[T any] struct {
	Method string `json:"method"`
	Params T      `json:"params"`
}

type Name struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

type ctxShim struct {
	*ws.Context
	*rpc.Mux
}

func (c ctxShim) Notify(method string, params any) {
	c.WriteNotification(c.Context, rpc.Notification[any]{
		Method: method,
		Params: params,
	})
}

func main() {
	db := observable.Map[Name, Person]{}

	go func() {
		for range time.NewTicker(1 * time.Second).C {
			for k, v := range db.Snapshot() {
				v.Age++
				db.Set(k, v)
			}
		}
	}()

	mux := rpc.Mux{}
	mux.Handle("person.get", func(ctx rpc.Context, req rpc.Request) (any, error) {
		var param Name
		if err := json.Unmarshal(req.Params, &param); err != nil {
			return nil, err
		}
		if _, ok := db.Get(param); !ok {
			db.Set(param, Person{
				FirstName: param.First,
				LastName:  param.Last,
				Age:       0,
			})
		}
		v, _ := db.Get(param)

		sub := db.Subscribe(param)
		ctx.OnDisconnect(sub.Close)
		go func() {
			for v := range sub.C {
				ctx.Notify("person.get", v)
			}
		}()
		return v, nil
	})
	mux.Handle("person.reset", func(ctx rpc.Context, req rpc.Request) (any, error) {
		var param Name
		if err := json.Unmarshal(req.Params, &param); err != nil {
			return nil, err
		}
		v, ok := db.Get(param)
		if !ok {
			return false, errors.New("not found")
		}
		v.Age = 0
		db.Set(param, v)
		return true, nil
	})

	upgrader := ws.Upgrader{
		Log: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
		OnMessage: func(ctx *ws.Context, buf []byte) {
			go mux.Dispatch(ctxShim{Context: ctx, Mux: &mux}, buf)
		},
	}

	http.HandleFunc("/ws", upgrader.Upgrade)
	fs := http.FileServer(http.Dir("ux/public"))
	http.Handle("/", fs)

	http.ListenAndServe(":8081", nil)
}
