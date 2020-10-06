package context

import "sync"

type Batcher interface {
	Add(maker func() interface{})
	Join() <-chan interface{}
}

type batcher struct {
	values chan interface{}
	wg     *sync.WaitGroup
	join   chan interface{}
}

func NewBatcher() Batcher {
	b := batcher{
		values: make(chan interface{}),
		wg:     &sync.WaitGroup{},
		join:   make(chan interface{}),
	}
	go func() {
		defer close(b.join)

		values := []interface{}{}
		for {
			value, more := <-b.values
			if more {
				values = append(values, value)
				continue
			}

			break
		}

		// push to join channel.
		// use seperate join channel to avoid bloking write to of
		// valeus chan (thus batcher.Add)
		for _, value := range values {
			b.join <- value
		}
	}()
	return b
}

func (b batcher) Add(maker func() interface{}) {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.values <- maker()
	}()
}

func (b batcher) Join() <-chan interface{} {
	go func() {
		b.wg.Wait()
		close(b.values)
	}()

	return b.join
}
