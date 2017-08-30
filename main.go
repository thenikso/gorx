package main

import (
	"fmt"
	"net/http"
	"time"
)

func fetchThing (id int) Signal {
	return NewSignal(func (subscriber Subscriber) {
		go func () {
			time.Sleep(2 * time.Second)
			subscriber.OnNext(fmt.Sprintf("thing fetched %d", id))
			subscriber.OnCompleted()
		}()
	})
}

func WriteToResponse(s Signal, w http.ResponseWriter) {
	done := make(chan bool, 1)
	s.SubscribeAuto(func (c string) {
		fmt.Fprint(w, c)
	}, func () {
		done <- true
	})
	<- done
}

func handler(w http.ResponseWriter, r *http.Request) {
	WriteToResponse(fetchThing(2), w)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}