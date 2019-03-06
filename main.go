package main

import (
	"errors"
	"fmt"
	Promise "promise-go/promise"
	"time"
)

// =========================== Demo ========================================
func main() {

	// ++++++++++++++++++++++ Demo - Chained catch ++++++++++++++++++++

	// p := Promise.NewPromise(func(resolve func(interface{}), reject func(error)) {
	// 	is_success := false
	// 	if is_success {
	// 		resolve(1)
	// 	} else {
	// 		er := errors.New("Rejected Message ")
	// 		reject(er)
	// 	}
	// })
	// p.Catch(func(err error) interface{} {
	// 	fmt.Println(err)
	// 	str := err.Error()
	// 	er := errors.New(str + "catch_1 ")
	// 	return er
	// }).Catch(func(err error) interface{} {
	// 	fmt.Println(err)
	// 	str := err.Error()
	// 	er := errors.New(str + "catch_2 ")
	// 	return er
	// }).Catch(func(err error) interface{} {
	// 	fmt.Println(err)
	// 	str := err.Error()
	// 	er := errors.New(str + "catch_3 ")
	// 	return er
	// })

	//===============================================================

	//+++++++++++++++++++ chained then demo +++++++++++++++++++++++++++

	// p := Promise.NewPromise(func(resolve func(interface{}), reject func(error)) {
	// 	is_success := true
	// 	if is_success {
	// 		resolve(1)
	// 	} else {
	// 		er := errors.New("Rejected Message ")
	// 		reject(er)
	// 	}
	// })

	// p.Then(func(msg interface{}) interface{} {
	// 	fmt.Println(msg)
	// 	val := msg.(int)
	// 	return val * 3
	// }, func(msg error) interface{} {
	// 	fmt.Println("Then Reject 3")
	// 	return 1
	// }).Then(func(msg interface{}) interface{} {
	// 	fmt.Println(msg)
	// 	val := msg.(int)

	// 	return val * 3
	// }, func(msg error) interface{} {
	// 	fmt.Println("Then Reject 4")
	// 	return 1
	// }).Then(func(msg interface{}) interface{} {
	// 	fmt.Println(msg)
	// 	val := msg.(int)

	// 	return val * 3
	// }, func(msg error) interface{} {
	// 	fmt.Println("Then Reject 4")
	// 	return 1
	// })

	//===============================================================

	//++++++++++++++++++++++++ Demo - complex chain +++++++++++++++++

	p := Promise.NewPromise(func(resolve func(interface{}), reject func(error)) {
		is_success := true
		if is_success {
			resolve(1)
		} else {
			er := errors.New("Rejected Message ")
			reject(er)
		}
	})

	p.Then(func(msg interface{}) interface{} {
		val := msg.(int)
		fmt.Println("Then", val*3)
		err := errors.New("Then Error")
		return err
	}, func(err error) interface{} {
		return 2
	}).Catch(func(err error) interface{} {
		fmt.Println(err)
		return 8
	}).Finally(func() interface{} {
		fmt.Println("Finally")
		return 4
	})

	// fmt.Println(x)

	//===============================================================
	time.Sleep(time.Second * 2)

}
