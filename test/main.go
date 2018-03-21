package main

import (
	"fmt"
)

type people struct {
	Age int
}

func (p people) new() {
	fmt.Println(p.Age)
}

func main() {
	p := people{}
	p.Age = 10
	p.new()
}
