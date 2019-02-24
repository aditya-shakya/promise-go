package main

import (
	"fmt"
	"time"
)

// import "errors"

type Promise struct {
	success chan string
	failure chan string
	state   string
}

func (p *Promise) resolve(result string) {
	go func(p *Promise, result string) {
		p.success <- result
		p.state = "success"
	}(p, result)

}

func (p *Promise) reject(error_message string) {
	go func(p *Promise, error_message string) {
		p.failure <- error_message
		p.state = "rejected"
	}(p, error_message)
}

func (p *Promise) init(runner func(resolve func(string), reject func(string))) {
	p.state = "pending"
	p.success = make(chan string)
	p.failure = make(chan string)
	runner(p.resolve, p.reject)

}

func execute_goroutine(channel chan string, onMessage func(string)) {
	message := <-channel
	onMessage(message)
}

func (p *Promise) catch(onRejected func(string)) *Promise {
	go execute_goroutine(p.failure, onRejected)
	return p
}

func (p *Promise) then(onFulfilled func(string), onRejected func(string)) {
	go func() {
		select {
		case error_message := <-p.failure:
			go onRejected(error_message)
		case result := <-p.success:
			go onFulfilled(result)
		}
	}()
}

func (p *Promise) fanally(onFinally func()) {
	final := make(chan string)
	go func() {
		for {
			select {
			case v1 := <-p.failure:
				final <- v1
			case v2 := <-p.success:
				final <- v2
			}
		}
	}()
	x := <-final
	fmt.Println(x)
	go onFinally()
}

func main() {
	p := new(Promise)
	// fmt.Println(p)
	p.init(func(resolve func(string), reject func(string)) {
		is_success := true
		if is_success {
			resolve("done")
		} else {
			reject("Rejected Message")
		}
	})
	p.then(func(msg string) {
		fmt.Println("Then Resolve")
		fmt.Println(msg)
	}, func(msg string) {
		fmt.Println("Then Reject")
		fmt.Println(msg)
	})

	p.fanally(func() {
		fmt.Println("finally")
		fmt.Println(p)
	})

	time.Sleep(time.Second * 2)

}
