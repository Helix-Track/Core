package main

import "fmt"

func printMe(word, secondWord string) {

	fmt.Println("Word: " + word + ", " + secondWord)
}

func getWords() (string, string) {

	return "Why", "Me"
}

func main() {

	var word string = "Hello"
	secondWord := "World"

	printMe(word, secondWord)

	word = "World"
	secondWord = "Hello"

	printMe(word, secondWord)

	word, secondWord = getWords()

	printMe(word, secondWord)
}
