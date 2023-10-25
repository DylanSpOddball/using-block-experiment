package main

import (
	"fmt"
	"sync"
)

// TODO - can we define an interface for Using/UsingVoid? CAn this interface "extend" Locker (i.e. require that an implementing type also implement Locker?)

func Using[L sync.Locker, T any](locker L, callback func() T) T {
	locker.Lock()
	defer locker.Unlock()
	return callback()
}

func UsingVoid[L sync.Locker](locker L, callback func()) {
	locker.Lock()
	defer locker.Unlock()
	callback()
}

// implements Locker
// realistically, rather than implementing Locker, you'd probably want to expose public methods that manage the mutex
type SharedCounter struct {
	mu    sync.Mutex
	Value int // intentionally publicly accessible
}

func (sc *SharedCounter) Lock() {
	fmt.Println("locking sc")
	sc.mu.Lock()
}

func (sc *SharedCounter) Unlock() {
	fmt.Println("unlocking sc")
	sc.mu.Unlock()
}

func main() {
	var mutex1 sync.Mutex

	n := Using(&mutex1, func() int {
		return 3
	})
	fmt.Println(n)

	// doesn't compile:
	//  type func() of func() {…} does not match func() T
	// Using(&mutex1, func() {
	// fmt.Println("Within using block!")
	// })

	UsingVoid(&mutex1, func() {
		fmt.Println("within using block!")
	})

	sc := &SharedCounter{
		Value: 0,
	}

	UsingVoid(sc, func() {
		sc.Value += 1
	})
}
