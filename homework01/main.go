package main

import (
	"fmt"
	"os/exec"
)

func main() {

	fmt.Println("Hello, Otus")
	git_version := exec.Command("git", "version")
	go_version := exec.Command("go", "version")
	go_out, _ := go_version.Output()
	git_out, _ := git_version.Output()
	fmt.Println(string(git_out))
	fmt.Println(string(go_out))

}
