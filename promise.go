package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var once sync.Once

type Promise struct {
	success                 chan interface{}
	failure                 chan error
	state                   string
	result                  interface{}
	error_msg               error
	successBroadcastListner []chan interface{}
	failureBroadcastListner []chan error

	wg  sync.WaitGroup
	mux sync.RWMutex
}

func (p *Promise) resolve(result interface{}) {
	go func(p *Promise, result interface{}) {
		p.mux.RLock()
		p.success <- result
		p.closeChannels()
		p.state = "success"
		p.result = result
		p.mux.RUnlock()
	}(p, result)
}

func (p *Promise) reject(error_message error) {
	go func(p *Promise, error_message error) {
		p.mux.RLock()
		p.failure <- error_message
		p.closeChannels()
		p.state = "rejected"
		p.error_msg = error_message
		p.mux.RUnlock()
	}(p, error_message)
}
func (p *Promise) closeChannels() {
	close(p.success)
	close(p.failure)
}
func (p *Promise) addSuccessListner() chan interface{} {
	ch := make(chan interface{})
	p.successBroadcastListner = append(p.successBroadcastListner, ch)
	return ch
}

func (p *Promise) addFailureListner() chan error {
	ch := make(chan error)
	p.failureBroadcastListner = append(p.failureBroadcastListner, ch)
	return ch
}

func (p *Promise) broadcastSuccess(msg interface{}) {
	for _, ch := range p.successBroadcastListner {
		go func(ch chan interface{}, msg interface{}) {
			ch <- msg
			close(ch)
		}(ch, msg)
	}
}

func (p *Promise) broadcastFailure(err error) {
	for _, ch := range p.failureBroadcastListner {
		go func(ch chan error, err error) {
			ch <- err
			close(ch)
		}(ch, err)
	}
}

func (p *Promise) init(runner func(resolve func(interface{}), reject func(error))) {
	p.state = "pending"
	p.success = make(chan interface{})
	p.failure = make(chan error)
	go runner(p.resolve, p.reject)
	go func() {
		select {
		case result := <-p.success:
			p.result = result
			p.state = "success"
			p.broadcastSuccess(result)
		case error_message := <-p.failure:
			p.error_msg = error_message
			p.state = "rejected"
			p.broadcastFailure(error_message)
		}
	}()
}

func (p *Promise) newPromise() *Promise {
	pr := new(Promise)
	pr.state = "pending"
	pr.success = make(chan interface{})
	pr.failure = make(chan error)

	go func() {
		select {
		case result := <-pr.success:
			pr.mux.RLock()
			pr.closeChannels()
			pr.result = result
			pr.state = "success"
			pr.broadcastSuccess(result)
			pr.mux.RUnlock()
		case error_message := <-pr.failure:
			pr.mux.RLock()
			pr.closeChannels()
			pr.error_msg = error_message
			pr.state = "rejected"
			pr.broadcastFailure(error_message)
			pr.mux.RUnlock()
		}
	}()

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
			failure := p.addFailureListner()
			select {
			case error_message := <-failure:
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
			failure := p.addFailureListner()
			success := p.addSuccessListner()
			select {
			case error_message := <-failure:
				pr.execute_and_pass_rejected(onRejected, error_message)
			case result := <-success:
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
			failure := p.addFailureListner()
			success := p.addSuccessListner()
			select {
			case <-success:
				pr.execute_and_pass_final(onFinally)
			case <-failure:
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
		return 2
	}).catch(func(err error) interface{} {
		fmt.Println(err)
		return 8
	}).finally(func() interface{} {
		fmt.Println("Finally")
		return 4
	})

	// fmt.Println(x)

	//===============================================================
	time.Sleep(time.Second * 2)

}
