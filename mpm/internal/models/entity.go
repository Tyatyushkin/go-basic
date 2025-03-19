package models

type Entity interface {
	GetID() int
	GetType() string
}
