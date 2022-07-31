package entity

type Closer interface {
	Close()
}

type Context interface {
	Notify(method string, param any)
	OnDisconnect(fn func())
}

type CanRegisterSubscription interface {
	RegisterSubscription(id string, sub Closer)
}
