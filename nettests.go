package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	cli "github.com/codegangsta/cli"
	hum "github.com/dustin/go-humanize"
	color "github.com/fatih/color"
	cn "github.com/whyrusleeping/go-ctrlnet"
)

var verbose bool
var noclean bool

func log(fstr string, args ...interface{}) {
	if verbose {
		fstr = strings.TrimRight(fstr, "\n")

		numfmt := strings.Count(fstr, "%")
		fmt.Printf(fstr, args[:numfmt]...)
		fmt.Println(args[numfmt:]...)
	}
}

func perr(fstr string, args ...interface{}) {
	if !strings.HasSuffix(fstr, "\n") {
		fstr += "\n"
	}
	args = append([]interface{}{color.RedString("ERROR:")}, args...)
	fmt.Printf("%s "+fstr, args...)
}

type MultinodeParams struct {
	NumNodes int
	FileSize int

	Net *cn.LinkSettings
}

type FetchStat struct {
	Duration time.Duration
	Total    int
	BW       float64

	DupBlocks int
}

type MultinodeOutput struct {
	FetchStats []*FetchStat
}

func (mo *MultinodeOutput) AverageBandwidth() float64 {
	var sum float64
	for _, f := range mo.FetchStats {
		sum += f.BW
	}

	return sum / float64(len(mo.FetchStats))
}

func getDupBlocksFromNode(n string) (int, error) {
	bstat, err := runCmdOnNode(n, "/bin/ipfs", "bitswap", "stat")
	if err != nil {
		return -1, err
	}

	lines := strings.Split(bstat, "\n")
	for _, l := range lines {
		if strings.Contains(l, "dup blocks") {
			fs := strings.Fields(l)
			n, err := strconv.Atoi(fs[len(fs)-1])
			if err != nil {
				return -1, err
			}

			return int(n), nil
		}
	}

	return -1, fmt.Errorf("no dup blocks field in output")
}

func RunMultinode(p *MultinodeParams) (*MultinodeOutput, error) {
	var nodes []string
	for i := 0; i < p.NumNodes; i++ {
		nd, err := startDockerNode("ipfs-node")
		if err != nil {
			return nil, err
		}
		log("started node: ", nd)
		nodes = append(nodes, nd)
	}

	defer func() {
		if noclean {
			return
		}
		for _, n := range nodes {
			err := killNode(n)
			if err != nil {
				perr("error killing node: %s", err)
			}
		}
	}()

	if err := setNetworkParams(p.Net); err != nil {
		return nil, err
	}

	// wait for nodes to finish starting up
	time.Sleep(time.Second * 2)

	zeroaddr, err := getNodeAddress(nodes[0])
	if err != nil {
		return nil, err
	}

	log("connecting nodes to: ", zeroaddr)

	for i := 1; i < len(nodes); i++ {
		_, err := runCmdOnNode(nodes[i], "/bin/ipfs", "swarm", "connect", zeroaddr)
		if err != nil {
			return nil, err
		}
	}

	log("creating a file")

	hash, err := runCmdOnNode(nodes[1], "/bin/addfile", fmt.Sprint(p.FileSize))
	if err != nil {
		return nil, err
	}
	hash = strings.TrimSpace(hash)
	log("fetching ref: %s", hash)

	results := make(chan *FetchStat, len(nodes))
	errs := make(chan error, len(nodes))
	for i, node := range nodes {
		if i == 1 {
			// dont run test for the adder
			continue
		}
		func(n string) {
			out, err := getFile(n, hash)
			if err != nil {
				errs <- err
				return
			}

			dupn, err := getDupBlocksFromNode(n)
			if err != nil {
				errs <- err
				return
			}

			out.DupBlocks = dupn

			results <- out
		}(node)
	}

	out := new(MultinodeOutput)
	for range nodes[1:] {
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
	out, err := exec.Command("docker", "run", "-dt", img).CombinedOutput()
	if err != nil {
		perr("Node start failed: ", string(out))
		return "", err
	}
	return strings.Trim(string(out), "\n \t"), nil
}

func setNetworkParams(np *cn.LinkSettings) error {
	if np == nil {
		return nil
	}

	ifs, err := cn.GetInterfaces("veth")
	if err != nil {
		return err
	}

	for _, iface := range ifs {
		log("setting link info on %s", iface)
		err := cn.SetLink(iface, np)
		if err != nil {
			return err
		}
	}
	return nil
}

func runCmdOnNode(id string, cmd ...string) (string, error) {
	args := append([]string{"exec", "-t", id}, cmd...)
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		perr("cmd '%q' failed: %q\n", cmd, string(out))
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
			return strings.TrimSpace(a), nil
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

func getFile(id, hash string) (*FetchStat, error) {
	url := "http://localhost:8080/api/v0/get/" + hash
	log("getting url %q", url)
	out, err := runCmdOnNode(id, "/bin/bwcurl", url)
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
	app := cli.NewApp()
	app.Email = "why@ipfs.io"
	app.Name = "ipfs-bench"
	app.Action = func(context *cli.Context) error {
		verbose = context.Bool("verbose")
		noclean = context.Bool("noclean")
		params := &MultinodeParams{
			NumNodes: context.Int("numnodes"),
			FileSize: context.Int("filesize"),
			Net: &cn.LinkSettings{
				Latency: context.Int("latency"),
			},
		}
		res, err := RunMultinode(params)
		if err != nil {
			perr("error running tests: %s", err)
			return nil
		}

		for _, f := range res.FetchStats {
			hb := hum.IBytes(uint64(f.BW))
			fmt.Printf("Took: %s, Average BW: %s, TX Size: %d, DupBlocks: %d\n", f.Duration, hb, f.Total, f.DupBlocks)
		}
		fmt.Println(hum.IBytes(uint64(res.AverageBandwidth())))
		return nil
	}

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "filesize",
			Usage: "specify size of file to transfer",
			Value: 1000000,
		},
		cli.IntFlag{
			Name:  "numnodes",
			Usage: "number of nodes to run test with",
			Value: 5,
		},
		cli.IntFlag{
			Name:  "latency",
			Usage: "set per-link latency",
			Value: 0,
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "print verbose logging info",
		},
		cli.BoolFlag{
			Name:  "noclean",
			Usage: "do not clean up docker nodes after test",
		},
	}
	app.Run(os.Args)
}
