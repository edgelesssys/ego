package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Getting https://www.edgeless.systems/")
	resp, err := http.Get("https://www.edgeless.systems/")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Status)
	err = resp.Body.Close()
	if err != nil {
		panic(err)
	}
}
