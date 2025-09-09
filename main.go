package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		fmt.Println("Invalid Arguments:")
	}

	switch args[0] {
	case "soal1":
		Soal1()
	case "soal2":
		Soal2()
	case "soal4":
		Soal4()
	default:
		fmt.Println("function not found")
	}

}
