package main

import "fmt"

func main() {

	done := make(chan bool)

	for i := 0; i < 10; i ++ {
		go func() {
			fmt.Println(i)
			done <- true
		}()
	}

	for i := 0; i < 10; i ++ {
		<- done
	}
}