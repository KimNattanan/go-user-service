package main

import (
	"encoding/json"
	"fmt"
)

type A struct {
	UserID string
}

func main() {
	a := &A{
		UserID: "u-123",
	}
	b, err := json.Marshal(a)
	if err != nil {
		fmt.Println("ERROR 1:", err)
	}
	c := &A{}
	err = json.Unmarshal(b, c)
	if err != nil {
		fmt.Println("ERROR 2:", err)
	}
	fmt.Println(c)
}
