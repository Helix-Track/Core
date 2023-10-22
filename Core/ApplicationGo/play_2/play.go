package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {

	counts := make(map[string]int)
	input := bufio.NewScanner(os.Stdin)

	for input.Scan() && len(counts) < 10 {

		counts[input.Text()]++
	}

	fmt.Printf("Lines: %d\n", len(counts))

	for line, n := range counts {

		if n > 1 {

			fmt.Printf("%s -> %d\n", line, n)
		}
	}
}
