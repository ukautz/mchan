[![Build Status](https://travis-ci.org/ukautz/mchan.svg?branch=master)](https://travis-ci.org/ukautz/mchan)

Simple Go package to merge arbitrary channels to drain them together.

This package uses reflection to cope with any kind of channel, which impacts
performance (run benchmarks).

Documentation
-------------

GoDoc can be [found here](http://godoc.org/github.com/ukautz/mchan)

Example
-------

``` go
// Merge multiple channels
ints := NewChannels()
i1 := make(chan int)
i2 := make(chan string)
i3 := make(chan struct{x int})
if err := ints.Add(i1, i2, i3); err != nil {
    panic(err)
}

// Feed the channels
for idx, ch := range []chan int{i1, i2, i3} {
    go func(c chan int, i int) {
        defer close(c)
        c <- i
        c <- i + 10
    }(ch, idx)
}

// Drain until all are closed
for i := range ints.Drain() {
    fmt.Printf("Got %d\n", i)
}

// Output (in random order):
// Got 0
// Got 1
// Got 2
// Got 10
// Got 11
// Got 12
```

See also
--------

* Awesome channel toolset: https://godoc.org/github.com/eapache/channels
