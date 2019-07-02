package capacityset

func WrapBlockingCall(f func() interface{}) <-chan interface{} {
	finish := make(chan interface{}, 1)

	go func() {
		finish <- f()

	}()

	return finish
}
