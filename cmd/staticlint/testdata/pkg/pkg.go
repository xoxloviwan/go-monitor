package main

import "os"

func main() {
	os.Exit(0) // want `os.Exit\(\) from main\(\) function of main package not allowed`
}
