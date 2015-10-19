package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type NetworkParams struct {
	Latency   int
	Bandwidth string
	Loss      string
}

type MultinodeParams struct {
	NumNodes int
	FileSize int

	Net *NetworkParams
}

type FetchStat struct {
	Duration time.Duration
	Total    int
	BW       float64
}

type MultinodeOutput struct {
	FetchStats []*FetchStat
}

func RunMultinode(p *MultinodeParams) (*MultinodeOutput, error) {
	var nodes []string
	for i := 0; i < p.NumNodes; i++ {
		nd, err := startDockerNode("ipfs-node")
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, nd)
	}

	defer func() {
		for _, n := range nodes {
			err := killNode(n)
			if err != nil {
				fmt.Printf("error killing node: %s", err)
			}
		}
	}()

	if err := setNetworkParams(p.Net); err != nil {
		return nil, err
	}

	zeroaddr, err := getNodeAddress(nodes[0])
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(nodes); i++ {
		_, err := runCmdOnNode(nodes[i], "/bin/ipfs", "swarm", "connect", zeroaddr)
		if err != nil {
			return nil, err
		}
	}

	hash, err := runCmdOnNode(nodes[1], "/bin/addfile", fmt.Sprint(p.FileSize))
	if err != nil {
		return nil, err
	}

	results := make(chan *FetchStat, len(nodes))
	errs := make(chan error)
	for _, node := range nodes {
		go func(n string) {
			out, err := catFile(n, hash)
			if err != nil {
				errs <- err
				return
			}

			results <- out
		}(node)
	}

	out := new(MultinodeOutput)
	for range nodes {
		select {
		case res := <-results:
			out.FetchStats = append(out.FetchStats, res)
		case err := <-errs:
			return nil, err
		}
	}
	return out, nil
}

func startDockerNode(img string) (string, error) {
	out, err := exec.Command("docker", "run", "-d", img).CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func setNetworkParams(np *NetworkParams) error {
	// TODO: this
	return nil
}

func runCmdOnNode(id string, cmd ...string) (string, error) {
	args := append([]string{"exec", id}, cmd...)
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func killNode(id string) error {
	_, err := exec.Command("docker", "kill", id).CombinedOutput()
	return err
}

func getNodeAddress(id string) (string, error) {
	out, err := runCmdOnNode(id, "/bin/ipfs", "id", "-f", "<addrs>")
	if err != nil {
		return "", err
	}

	parts := strings.Split(out, "\n")
	for _, a := range parts {
		if strings.HasPrefix(a, "/ip4/172.17") {
			return a, nil
		}
	}

	return "", errors.New("no valid addresses in output")
}

func catFile(id, file string) (*FetchStat, error) {
	out, err := runCmdOnNode(id, "/bin/bwcurl", "http://localhost:8080/ipfs/"+file)
	if err != nil {
		return nil, err
	}

	fs := new(FetchStat)
	err = json.Unmarshal([]byte(out), fs)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

func main() {
	params := &MultinodeParams{
		NumNodes: 10,
		FileSize: 10000000,
	}
	res, err := RunMultinode(params)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
