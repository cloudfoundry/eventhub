package eventhub

type Event interface {
	EventType() string
}
