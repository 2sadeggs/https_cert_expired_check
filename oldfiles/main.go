package main

import (
	"fmt"
	"time"
)

//https://www.freecodecamp.org/news/how-to-validate-ssl-certificates-in-go/

func main() {
	timeout := make(chan bool, 3)
	go func() {
		time.Sleep(1e9) // sleep one second
		timeout <- true
		timeout <- false
		timeout <- true
	}()
	ch := make(chan int)

	for {
		select {
		case <-ch:
			fmt.Println("done!")
			return
		case <-timeout:
			fmt.Println("timeout!")
		}
	}
	//select {
	//case <-ch:
	//	//fmt.Println("done!")
	//	return
	//case <-timeout:
	//	fmt.Println("timeout!")
	//}
	//ch1 := make(chan int, 1)
	//ch2 := make(chan int, 1)
	//
	//select {
	//case <-ch1:
	//	fmt.Println("ch1 pop one element")
	//case <-ch2:
	//	fmt.Println("ch2 pop one element")
	//default:
	//	fmt.Println("default")
	//}
}
