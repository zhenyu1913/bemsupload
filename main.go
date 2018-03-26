package main

import (
	"log"
)

func main() {
	err := upload()
	if err != nil {
		log.Fatal(err)
	}
}
