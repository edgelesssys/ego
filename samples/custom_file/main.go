package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	fmt.Println("Getting https://www.edgeless.systems/")
	resp, err := http.Get("https://www.edgeless.systems/")
	if err != nil {
		fmt.Println(err.Error())
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(body))
}
