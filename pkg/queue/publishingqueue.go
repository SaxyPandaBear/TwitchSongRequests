package queue

type Publisher interface {
	Publish(val interface{}) error
}
