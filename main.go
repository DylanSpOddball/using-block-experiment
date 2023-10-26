package main

import (
	"fmt"
	"sync"
)

// Having an interface for UseWithLock that embeds sync.Locker and some sort of Use() function isn't really worthwhile;
// having these standalone functions are probably the most useful thing

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

// implements sync.Locker
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

	NUM_GOROUTINES := 1000

	singleGoroutineCounter := 0
	for i := 0; i < NUM_GOROUTINES; i++ {
		singleGoroutineCounter += i
	}
	fmt.Printf("Expected sum result: %v\n", singleGoroutineCounter)

	unprotectedDone := make(chan struct{}, NUM_GOROUTINES)
	unprotectedCounter := 0 // not protected by mutex; race condition expected
	for i := 0; i < NUM_GOROUTINES; i++ {
		i := i // avoid loop variable capture
		go func() {
			unprotectedCounter += i
			unprotectedDone <- struct{}{}
		}()
	}
	for i := 0; i < NUM_GOROUTINES; i++ {
		<-unprotectedDone
	}
	fmt.Printf("Unprotected counter value: %v\n", unprotectedCounter)

	protectedDone := make(chan struct{}, NUM_GOROUTINES)
	protectedCounter := 0
	for i := 0; i < NUM_GOROUTINES; i++ {
		i := i // avoid loop variable capture
		go func() {
			UsingVoid(&mutex1, func() {
				protectedCounter += i
			})
			protectedDone <- struct{}{}
		}()
	}
	for i := 0; i < NUM_GOROUTINES; i++ {
		<-protectedDone
	}
	fmt.Printf("Protected counter value: %v\n", protectedCounter)
}
