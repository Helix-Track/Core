package main

import "fmt"

func newInt() *int {

	return new(int)
}

func newInt2() *int {

	var dummy int
	return &dummy
}

func main() {

	p := new(int)
	// p, of type *int, points to an unnamed int variable
	fmt.Println(*p) // "0"
	*p = 2
	// sets the unnamed int to 2
	fmt.Println(*p) // "2"
	*p = 3
	// sets the unnamed int to 3
	fmt.Println(*p) // "3"

	a := newInt()
	b := newInt2()

	fmt.Println(*a) // "0"
	fmt.Println(*b) // "0"

	*a = 100
	*b = *a

	fmt.Println(*a) // "100"
	fmt.Println(*b) // "100"

	fmt.Println(a == b) // "false"

	a = b

	fmt.Println(*a) // "100"
	fmt.Println(*b) // "100"

	fmt.Println(a == b) // "true"
}
