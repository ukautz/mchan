/*
Simple package to merge arbitrary channels to drain them together.

This package uses reflection to cope with any kind of channel, which impacts
performance (run benchmarks).
*/
package mchan

import (
	"fmt"
	"reflect"
	"sync"
)

// Channels is a container for the merged channels
type Channels struct {
	channels []reflect.Value
}

// Constructor to create a new merged channel container
func NewChannels() *Channels {
	return &Channels{
		channels: make([]reflect.Value, 0),
	}
}

// Add a new channel to the list
func (this *Channels) Add(chs ...interface{}) error {
	for _, ch := range chs {
		v := reflect.ValueOf(ch)
		if k := v.Kind(); k != reflect.Chan {
			return fmt.Errorf("Cannot add %s as channel!", k)
		} else if v.Type().ChanDir() & reflect.RecvDir != reflect.RecvDir {
			return fmt.Errorf("Cannot add non-receiving channel")
		}
		this.channels = append(this.channels, v)
	}
	return nil
}

// Drain all channels until all are closed
func (this *Channels) Drain() chan interface{} {
	res := make(chan interface{})
	var wg sync.WaitGroup
	for _, channel := range this.channels {
		wg.Add(1)
		go func(ch reflect.Value) {
			defer wg.Done()
			for {
				if x, ok := ch.Recv(); !ok {
					return
				} else {
					res <- x.Interface()
				}
			}
		}(channel)
	}
	go func() {
		defer close(res)
		wg.Wait()
	}()
	return res
}
