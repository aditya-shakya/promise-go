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

func (p *Promise) init(runner func(resolve func(string), reject func(string))) *Promise {
	p.state = "pending"
	p.success = make(chan string)
	p.failure = make(chan string)
	runner(p.resolve, p.reject)
	return p
}

func execute_goroutine(channel chan string, onMessage func(string)) {
	message := <-channel
	onMessage(message)
}

func (p *Promise) catch(onRejected func(string)) *Promise {
	go execute_goroutine(p.failure, onRejected)
	return p
}

func (p *Promise) then(onFulfilled func(string), onRejected func(string)) *Promise {
	go func() {
		select {
		case error_message := <-p.failure:
			go onRejected(error_message)
		case result := <-p.success:
			go onFulfilled(result)
		}
	}()
	return p
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

// func add(a int, b int) (int, error){
// 	sum := a + b
// 	if sum == 4 {
// 		return sum, Errors.new("this is a error")
// 	}
// 	return a + b, nil
// }

func main() {
	p := new(Promise)
	// fmt.Println(p)
	p.init(func(resolve func(string), reject func(string)) {
		is_success := false
		if is_success {
			resolve("done")
		} else {
			reject("Rejected Message")
		}
	}).then(func(msg string) {
		fmt.Println("Then Resolve")
		fmt.Println(msg)
	}, func(msg string) {
		fmt.Println("Then Reject")
		fmt.Println(msg)
	}).fanally(func() {
		fmt.Println("finally")
		fmt.Println(p)
	})
	// .catch(func(msg string) {
	// 	fmt.Println("In catch")
	// 	fmt.Println(msg)
	// })
	time.Sleep(time.Second * 2)

}
