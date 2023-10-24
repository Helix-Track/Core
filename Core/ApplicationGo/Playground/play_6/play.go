package main

import (
	"flag"
	"fmt"
	"strconv"
)

var port = flag.Int("port", 8080, "Port for the server")
var threads = flag.Int("threads", 1, "Number of the threads")

func main() {

	fmt.Println("Hello there!")
	flag.Parse()

	portStr := strconv.Itoa(*port)
	fmt.Println("Port: " + portStr)

	threadsStr := strconv.Itoa(*threads)
	fmt.Println("Threads: " + threadsStr)
}
