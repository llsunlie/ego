package app

// IDGenerator generates unique identifiers for new entities.
type IDGenerator interface {
	New() string
}
