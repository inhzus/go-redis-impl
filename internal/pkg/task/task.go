package task

// Task is required to connect between handler & consumer
type Task interface {
	Task() Task
}

type Msg interface {
	Msg() Msg
}

