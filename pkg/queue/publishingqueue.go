package queue

type Publisher interface {
	// Publish takes a destination and a value, and "publishes" to the given target destination
	Publish(target, val string) error
}
