package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	files, _ := ioutil.ReadDir("/media/pi/Dragon")
	for _, file := range files {
		fmt.Println(file.Name())
	}
}
