package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	hum "github.com/dustin/go-humanize"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Must specify url to fetch")
		return
	}

	resp, err := http.Get(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	before := time.Now()
	n, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	took := time.Now().Sub(before)
	bw := float64(n) / took.Seconds()
	fmt.Printf("fetched %d bytes in %s\n", n, took)
	fmt.Printf("speed = %s\n", hum.IBytes(uint64(bw)))
}
