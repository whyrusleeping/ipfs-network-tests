package main

import (
	"github.com/dustin/randbo"

	"fmt"
	"io"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("need to specify size!")
		os.Exit(1)
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r := randbo.New()
	io.CopyN(os.Stdout, r, int64(n))
}
