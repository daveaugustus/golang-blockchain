package main

import "fmt"

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

func main() {
	fmt.Println("Starting")
}
