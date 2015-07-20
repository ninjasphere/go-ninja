// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ninjasphere/go-ninja/simtime"
)

func main() {

	//simtime.SetCurrentTime(time.Now().Add(-time.Hour * 24 * 365 * 5))

	go func() {

		log.Printf("1. Before day sleep: %s", time.Now())
		simtime.Sleep(time.Hour * 24)
		log.Printf("1. After day sleep: %s", time.Now())
		simtime.Continue()

		x := simtime.Tick(time.Hour)

		for t := range x {
			log.Println("1. -------")
			log.Printf("1. Fake Time : %s", t)
			log.Printf("1. Fake Now(): %s", simtime.Now())
			log.Printf("1. Real Now(): %s", time.Now())

			if simtime.Now() != t {
				panic("DIFFERENT!")
			}

			simtime.Continue()
		}
	}()

	go func() {
		x := simtime.Tick(time.Hour * 5)

		for t := range x {
			log.Println("2. -------")
			log.Printf("2. Fake Time : %s", t)
			log.Printf("2. Fake Now(): %s", simtime.Now())
			log.Printf("2. Real Now(): %s", time.Now())

			if simtime.Now() != t {
				panic("DIFFERENT!")
			}

			simtime.Continue()
		}

	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

}
