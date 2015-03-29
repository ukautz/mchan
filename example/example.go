package main

import (
	"fmt"
	"github.com/ukautz/mchan"
)

func main() {
	// Merge multiple channels
	ints := mchan.NewChannels()
	ci := make(chan int)
	cs := make(chan string)
	cx := make(chan struct{x int})
	if err := ints.Add(ci, cs, cx); err != nil {
		panic(err)
	}

	// Feed the channels
	go func() {
		defer close(ci)
		for i := 0; i < 2; i++ {
			ci <- i
		}
	}()
	go func() {
		defer close(cs)
		for i := 0; i < 2; i++ {
			cs <- fmt.Sprintf("Num %d", i)
		}
	}()
	go func() {
		defer close(cx)
		for i := 0; i < 2; i++ {
			cx <- struct{x int}{x: i}
		}
	}()

	// Drain until all are closed
	for v := range ints.Drain() {
		fmt.Printf("Got %v\n", v)
	}

	// Output (in random order):
	// Got 0
	// Got Num 0
	// Got {0}
	// Got 1
	// Got Num 1
	// Got {1}
}
