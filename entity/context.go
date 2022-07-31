package entity

type Closer interface {
	Close()
}

type CanRegisterSubscription interface {
	RegisterSubscription(id string, sub Closer)
}
