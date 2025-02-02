package main

import "fmt"

func main() {
	var x int
	fmt.Print("Введите размер доски: ")
	_, err := fmt.Scan(&x)
	if err != nil {
		return
	}
	chessBoard(x)
}

func chessBoard(x int) {
	for i := 0; i < x; i++ {
		for j := 0; j < x; j++ {
			if (i+j)%2 == 0 {
				print("#")
			} else {
				print(" ")
			}
			if j == x-1 {
				print("\n")
			}

		}
	}
}
