package mchan

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"sync"
)

func TestChannels(t *testing.T) {
	Convey("Drainging empty channel set stops", t, func() {
		m := NewChannels()
		c := 0
		for _ = range m.Drain() {
			c++
		}
		So(c, ShouldEqual, 0)
	})
	Convey("Merging channels into one", t, func() {
		m := NewChannels()
		c1 := make(chan string)
		So(m.Add(c1), ShouldBeNil)
		c2 := make(chan string)
		c3 := make(chan string)
		So(m.Add(c2, c3), ShouldBeNil)

		Convey("Drainging them at once works", func() {
			go func() {
				defer close(c1)
				for i := 0; i < 5; i++ {
					c1 <- fmt.Sprintf("From 1: %d", i)
				}
			}()
			go func() {
				defer close(c2)
				for i := 0; i < 5; i++ {
					c2 <- fmt.Sprintf("From 2: %d", i)
				}
			}()
			go func() {
				defer close(c3)
				for i := 0; i < 5; i++ {
					c3 <- fmt.Sprintf("From 3: %d", i)
				}
			}()

			r := make([]string, 0)
			for s := range m.Drain() {
				r = append(r, s.(string))
			}
			So(len(r), ShouldEqual, 15)
		})
		Convey("But merging non channels fails", func() {
			m := NewChannels()
			So(m.Add("foo"), ShouldResemble, fmt.Errorf("Cannot add string as channel!"))
			So(m.Add(123), ShouldResemble, fmt.Errorf("Cannot add int as channel!"))
			So(m.Add(struct{ x int }{x: 1}), ShouldResemble, fmt.Errorf("Cannot add struct as channel!"))
		})
		Convey("But merging non receiving channels fails", func() {
			m := NewChannels()
			c := make(chan int)
			e := func (c chan<- int) error {
				return m.Add(c)
			}(c)
			So(e, ShouldResemble, fmt.Errorf("Cannot add non-receiving channel"))
		})
	})
	Convey("Merging different channel types works", t, func() {
		m := NewChannels()
		c1 := make(chan string)
		So(m.Add(c1), ShouldBeNil)

		c2 := make(chan int)
		So(m.Add(c2), ShouldBeNil)

		c3 := make(chan struct{ x int })
		So(m.Add(c3), ShouldBeNil)

		go func() {
			defer close(c1)
			c1 <- "Hello"
		}()

		go func() {
			defer close(c2)
			c2 <- 123
		}()

		go func() {
			defer close(c3)
			c3 <- struct{ x int }{x: 5}
		}()

		r := make([]interface{}, 0)
		for s := range m.Drain() {
			r = append(r, s)
		}

		So(r, ShouldResemble, []interface{}{
			"Hello",
			123,
			struct{ x int }{x: 5},
		})
	})
}

func ExampleChannels() {
	// Merge multiple channels
	ints := NewChannels()
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

func BenchmarkChannelsMerged(b *testing.B) {
	m := NewChannels()
	c1 := make(chan int)
	c2 := make(chan int)
	c3 := make(chan int)
	m.Add(c1)
	m.Add(c2)
	m.Add(c3)

	go func() {
		for i := 0; i < b.N; i++ {
			c1 <- 1
		}
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			c2 <- 1
		}
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			c3 <- 1
		}
	}()

	s := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range m.Drain() {
			s += v.(int)
			if s == b.N * 3 {
				break
			}
		}
	}()

	wg.Wait()
}

func BenchmarkChannelsSingle(b *testing.B) {
	c1 := make(chan int)

	go func() {
		for i := 0; i < b.N; i++ {
			c1 <- 1
		}
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			c1 <- 1
		}
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			c1 <- 1
		}
	}()

	s := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range c1 {
			s += v
			if s == b.N {
				break
			}
		}
	}()

	wg.Wait()
}

func BenchmarkChannelsConcur(b *testing.B) {
	c1 := make(chan int)
	c2 := make(chan int)
	c3 := make(chan int)

	go func() {
		for i := 0; i < b.N; i++ {
			c1 <- 1
		}
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			c2 <- 1
		}
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			c3 <- 1
		}
	}()

	s1 := 0
	s2 := 0
	s3 := 0
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for v := range c1 {
			s1 += v
			if s1 == b.N {
				break
			}
		}
	}()
	go func() {
		defer wg.Done()
		for v := range c2 {
			s2 += v
			if s2 == b.N {
				break
			}
		}
	}()
	go func() {
		defer wg.Done()
		for v := range c3 {
			s3 += v
			if s3 == b.N {
				break
			}
		}
	}()

	wg.Wait()
}
