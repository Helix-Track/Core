package main

import "fmt"

/*
Return pointer to integer
*/
func f() *int {

	v := 1
	return &v
}

func main() {

	x := 1
	p := &x

	// p, of type *int, points to x
	fmt.Println(*p) // "1"
	*p = 2

	// equivalent to x = 2
	fmt.Println(x) // "2"

	var y int = 2
	//            true     false     false     true
	fmt.Println(&x == &x, &x == &y, &x == nil, x == y)

	var pointer = f()
	fmt.Println(*pointer) // 1
	*pointer = 2
	fmt.Println(*pointer) // 2
}
