package mychannel

func Drain[T any](ch chan T) {
	for range ch {
	}
}

func CloseAndDrain[T any](ch chan T) {
	close(ch)
	Drain(ch)
}
