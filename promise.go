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
	success   chan interface{}
	failure   chan error
	final     chan struct{}
	state     string
	result    interface{}
	error_msg error
}

func (p *Promise) resolve(result interface{}) {
	go func(p *Promise, result interface{}) {
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

func (p *Promise) init(runner func(resolve func(interface{}), reject func(error))) {
	p.state = "pending"
	p.success = make(chan interface{})
	p.failure = make(chan error)
	p.final = make(chan struct{})
	go runner(p.resolve, p.reject)
}

func (p *Promise) newPromise() *Promise {
	pr := new(Promise)
	pr.state = "pending"
	pr.success = make(chan interface{})
	pr.failure = make(chan error)
	pr.final = make(chan struct{})

	return pr
}

//====================================================================
func (pr *Promise) execute_and_pass_rejected(onRejected func(error) interface{}, err error) {
	resp := onRejected(err)
	switch data := resp.(type) {
	case error:
		pr.failure <- data
	default:
		pr.success <- data
	}
}

func (pr *Promise) execute_and_pass_result(onFulfilled func(interface{}) interface{}, msg interface{}) {
	resp := onFulfilled(msg)
	switch data := resp.(type) {
	case error:
		pr.failure <- data
	default:
		pr.success <- data
	}
}

func (pr *Promise) execute_and_pass_final(onFinally func() interface{}) {
	resp := onFinally()
	switch data := resp.(type) {
	case error:
		pr.failure <- data
	default:
		pr.success <- data
	}
}

//+++++++++++++++++++++++++ catch, then and finally +++++++++++++++++++++
func (p *Promise) catch(onRejected func(error) interface{}) *Promise {
	pr := p.newPromise()
	go func() {
		if p.state == "pending" {
			select {
			case error_message := <-p.failure:
				pr.execute_and_pass_rejected(onRejected, error_message)
			}
		} else if p.state == "rejected" {
			pr.execute_and_pass_rejected(onRejected, p.error_msg)
		}
	}()
	return pr
}

func (p *Promise) then(onFulfilled func(interface{}) interface{}, onRejected func(error) interface{}) *Promise {
	pr := p.newPromise()
	go func() {
		if p.state == "pending" {
			select {
			case error_message := <-p.failure:
				pr.execute_and_pass_rejected(onRejected, error_message)
			case result := <-p.success:
				pr.execute_and_pass_result(onFulfilled, result)
			}
		} else if p.state == "success" {
			pr.execute_and_pass_result(onFulfilled, p.result)
		} else if p.state == "rejected" {
			pr.execute_and_pass_rejected(onRejected, p.error_msg)
		}
	}()
	return pr
}

func (p *Promise) finally(onFinally func() interface{}) *Promise {
	pr := p.newPromise()
	go func() {
		if p.state == "pending" {
			select {
			case <-p.success:
				pr.execute_and_pass_final(onFinally)
			case <-p.failure:
				pr.execute_and_pass_final(onFinally)
			}
		} else {
			pr.execute_and_pass_final(onFinally)
		}
	}()
	return pr
}

// =========================== Demo ========================================
func main() {
	p := new(Promise)

	// ++++++++++++++++++++++ Demo - Chained catch ++++++++++++++++++++

	// p.init(func(resolve func(interface{}), reject func(error)) {
	// 	is_success := false
	// 	if is_success {
	// 		resolve(1)
	// 	} else {
	// 		er := errors.New("Rejected Message ")
	// 		reject(er)
	// 	}
	// })
	// p.catch(func(err error) interface{} {
	// 	fmt.Println(err)
	// 	str := err.Error()
	// 	er := errors.New(str + "catch_1 ")
	// 	return er
	// }).catch(func(err error) interface{} {
	// 	fmt.Println(err)
	// 	str := err.Error()
	// 	er := errors.New(str + "catch_2 ")
	// 	return er
	// }).catch(func(err error) interface{} {
	// 	fmt.Println(err)
	// 	str := err.Error()
	// 	er := errors.New(str + "catch_3 ")
	// 	return er
	// })

	//===============================================================

	//+++++++++++++++++++ chained then demo +++++++++++++++++++++++++++

	// p.init(func(resolve func(interface{}), reject func(error)) {
	// 	is_success := true
	// 	if is_success {
	// 		resolve(1)
	// 	} else {
	// 		er := errors.New("Rejected Message ")
	// 		reject(er)
	// 	}
	// })

	// p.then(func(msg interface{}) interface{} {
	// 	fmt.Println(msg)
	// 	val := msg.(int)
	// 	return val * 3
	// }, func(msg error) interface{} {
	// 	fmt.Println("Then Reject 3")
	// 	return 1
	// }).then(func(msg interface{}) interface{} {
	// 	fmt.Println(msg)
	// 	val := msg.(int)

	// 	return val * 3
	// }, func(msg error) interface{} {
	// 	fmt.Println("Then Reject 4")
	// 	return 1
	// }).then(func(msg interface{}) interface{} {
	// 	fmt.Println(msg)
	// 	val := msg.(int)

	// 	return val * 3
	// }, func(msg error) interface{} {
	// 	fmt.Println("Then Reject 4")
	// 	return 1
	// })

	//===============================================================

	//++++++++++++++++++++++++ Demo - complex chain +++++++++++++++++

	p.init(func(resolve func(interface{}), reject func(error)) {
		is_success := true
		if is_success {
			resolve(1)
		} else {
			er := errors.New("Rejected Message ")
			reject(er)
		}
	})

	p.then(func(msg interface{}) interface{} {
		val := msg.(int)
		fmt.Println("Then", val*3)
		err := errors.New("Then Error")
		return err
	}, func(err error) interface{} {
		return 1
	}).catch(func(err error) interface{} {
		fmt.Println(err)
		return 1
	}).finally(func() interface{} {
		fmt.Println("Finally")
		return 1
	})

	//===============================================================
	time.Sleep(time.Second * 2)

}
