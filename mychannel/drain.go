package mychannel

func Drain[T any](ch chan T) {
	close(ch)
	for range ch {
	}
}
