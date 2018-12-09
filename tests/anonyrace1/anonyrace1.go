package main

import "fmt"

func test() {
	sum := 0

	for j := 0; j < 10; j ++ {
		go func() {
			sum += 1
		}()
	}

	fmt.Println(sum)
}



func main() {

	done := make(chan bool)

	for i := 0; i < 10; i ++ {
		go func() {
			j := i
			fmt.Println(j)
			done <- true
		}()
	}

	for i := 0; i < 10; i ++ {
		<- done
	}
}