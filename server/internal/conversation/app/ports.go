package app

type IDGenerator interface {
	New() string
}
