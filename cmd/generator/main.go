package main

import (
	"fmt"
	"os"
)

func main() {
	i := 0
	for {
		f, err := os.Create("./tmp/file" + fmt.Sprint(i) + ".txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		f.WriteString("Hello, World!")
		i++
	}
}
