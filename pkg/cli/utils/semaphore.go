package utils

type Semaphore struct {
	channel chan struct{}
}

func NewSemaphore(size int) *Semaphore {
	return &Semaphore{channel: make(chan struct{}, size)}
}

func (r *Semaphore) Locked(v func()) {
	r.channel <- struct{}{}

	defer func() {
		<-r.channel
	}()

	v()
}
