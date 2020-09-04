package a

import "fmt"

func f1() { // want "No need to define these variables: a"
	var a = "Hello"
	fmt.Println(a)
}

func f2() { // want "No need to define these variables: a, b\nUsed same consts multiple times, replace with variable: \"Hello\""
	var a = "Hello"
	fmt.Println(a)
	var b = "Hello"
	fmt.Println(b)
}

func f3() { // want "Used same consts multiple times, replace with variable: \"Hello\""
	fmt.Println("Hello")
	fmt.Println("Hello")
}
