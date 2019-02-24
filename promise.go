package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// import "errors"
var once sync.Once

type Promise struct {
	success chan string
	failure chan error
	final   chan struct{}
	state   string
	wg      sync.WaitGroup

	onFinallyCallbacks []func()

	onRejectedCallbacks  []func(error)
	onFulfilledCallbacks []func(string)
}

func (p *Promise) resolve(result string) {
	go func(p *Promise, result string) {
		p.success <- result
		p.state = "success"
	}(p, result)

}

func (p *Promise) reject(error_message error) {
	go func(p *Promise, error_message error) {
		p.failure <- error_message
		p.state = "rejected"
	}(p, error_message)
}

func (p *Promise) init(runner func(resolve func(string), reject func(error))) {
	p.state = "pending"
	p.success = make(chan string)
	p.failure = make(chan error)
	p.final = make(chan struct{})
	runner(p.resolve, p.reject)
}

//+++++++++++ executer methods for rejected +++++++++++++++++++++++++++
func (p *Promise) executeRejected(error_message error) {
	for _, callback := range p.onRejectedCallbacks {
		p.wg.Add(1)
		go func(error_message error, callback func(error)) {
			callback(error_message)
			p.wg.Done()
		}(error_message, callback)
	}
	p.wg.Wait()
	p.final <- struct{}{}
}

//+++++++++++ executer methods for fullfilled +++++++++++++++++++++++++++
func (p *Promise) executeFulfilled(msg string) {
	for _, callback := range p.onFulfilledCallbacks {
		p.wg.Add(1)
		go func(msg string, callback func(string)) {
			callback(msg)
			p.wg.Done()
		}(msg, callback)
	}
	p.wg.Wait()
	p.final <- struct{}{}
}

//+++++++++++ executer methods for finally +++++++++++++++++++++++++++
func (p *Promise) executeFinally() {
	for _, callback := range p.onFinallyCallbacks {
		go callback()
	}
}

//====================================================================

//+++++++++++++++++++++++++ catch, then and finally +++++++++++++++++++++
func (p *Promise) catch(onRejected func(error)) {
	p.onRejectedCallbacks = append(p.onRejectedCallbacks, onRejected)
	catchListner := func() {
		go func() {
			select {
			case error_message := <-p.failure:
				p.executeRejected(error_message)
			}
		}()
	}
	once.Do(catchListner)
}

func (p *Promise) then(onFulfilled func(string), onRejected func(error)) {
	p.onFulfilledCallbacks = append(p.onFulfilledCallbacks, onFulfilled)
	p.onRejectedCallbacks = append(p.onRejectedCallbacks, onRejected)

	thenListner := func() {
		go func() {
			select {
			case error_message := <-p.failure:
				p.executeRejected(error_message)
			case result := <-p.success:
				p.executeFulfilled(result)
			}
		}()
	}
	once.Do(thenListner)
}

func (p *Promise) fanally(onFinally func()) {
	p.onFinallyCallbacks = append(p.onFinallyCallbacks, onFinally)

	finallyListner := func() {
		go func() {
			select {
			case <-p.final:
				p.executeFinally()
			}
		}()
	}

	finallyListner()
}

// =========================== Demo ========================================
func main() {
	p := new(Promise)

	p.init(func(resolve func(string), reject func(error)) {
		is_success := false
		if is_success {
			resolve("done")
		} else {
			er := errors.New("Rejected Message")
			reject(er)
		}
	})

	p.then(func(msg string) {
		fmt.Println("Then Resolve 1")
	}, func(msg error) {
		fmt.Println("Then Reject 1")
	})

	p.then(func(msg string) {
		fmt.Println("Then Resolve 2")
	}, func(msg error) {
		fmt.Println("Then Reject 2")
	})

	p.then(func(msg string) {
		fmt.Println("Then Resolve 3")
	}, func(msg error) {
		fmt.Println("Then Reject 3")
	})

	p.then(func(msg string) {
		fmt.Println("Then Resolve 4")
	}, func(msg error) {
		fmt.Println("Then Reject 4")
	})
	p.catch(func(msg error) {
		fmt.Println("Then Reject 1")
	})

	p.catch(func(msg error) {
		fmt.Println("Then Reject 2")
	})
	p.catch(func(msg error) {
		fmt.Println("Then Reject 3")
	})
	p.catch(func(msg error) {
		fmt.Println("Then Reject 4")
	})

	p.fanally(func() {
		fmt.Println("finally 1")
	})

	p.fanally(func() {
		fmt.Println("finally 2")
	})

	p.fanally(func() {
		fmt.Println("finally 3")
	})

	p.fanally(func() {
		fmt.Println("finally 4")
	})
	time.Sleep(time.Second * 2)

}
