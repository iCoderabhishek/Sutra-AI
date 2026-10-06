package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {

	fmt.Println("Welcome to SutraAI by Abhishek")
	for {
		fmt.Printf("Enter your text >>> ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		text := scanner.Text()

		fmt.Println("You entered - ", text)
	}

}
