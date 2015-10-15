package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Output struct {
	Total    int64
	Duration time.Duration
	BW       float64
}

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

	out := Output{
		Total:    n,
		Duration: took,
		BW:       bw,
	}

	data, err := json.MarshalIndent(out, "", "\t")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(string(data))
}
