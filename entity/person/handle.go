package person

import (
	"encoding/json"
	"errors"
	"test/entity"
	"test/observable"
	"test/rpc"
	"time"
)

var db = observable.Map[Name, Person]{}

func init() {
	go func() {
		for range time.NewTicker(1 * time.Second).C {
			for k, v := range db.Snapshot() {
				v.Age++
				db.Set(k, v)
			}
		}
	}()
}

var HandleGet = func(ctx rpc.Context, req rpc.Request) (any, error) {
	var param Request
	if err := json.Unmarshal(req.Params, &param); err != nil {
		return nil, err
	}
	if _, ok := db.Get(param.Name); !ok {
		db.Set(param.Name, Person{
			FirstName: param.First,
			LastName:  param.Last,
			Age:       0,
		})
	}
	v, _ := db.Get(param.Name)

	sub := db.Subscribe(param.Name)
	if c, ok := ctx.(entity.CanRegisterSubscription); ok {
		c.RegisterSubscription(param.Subscribe, sub)
	}
	ctx.OnDisconnect(sub.Close)
	go func() {
		for v := range sub.C {
			ctx.Notify("person.get", v)
		}
	}()
	return v, nil
}

var HandleReset = func(ctx rpc.Context, req rpc.Request) (any, error) {
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
}
