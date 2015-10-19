# ipfs network tests
This is a set of scripts to gather performance benchmarks about
various ipfs network operations (add, cat, lookups, etc) in a controlled
network environment. Intended to be used with [ctrlnet](github.com/whyrusleeping/testnet-utils)

## multinode_fetch
multinode fetch will spawn up a bunch of nodes (default 10), add a file on one
(default file size is 1MB) and cat it on all the other nodes, recording the
transfer speed for each node. The test takes two arguments:

#### size
Size is the size of the file in whole megabytes (2^20 bytes)

#### numnodes
numnodes is the number of nodes used in the test

Also configurable via environment variables are:

#### `NET_SETUP`
NET_SETUP should be set to a command or script to be run after the docker
network has been created. For example: `ctrlnet lat on 50ms` to set a 100ms
rtt between all nodes in the test. It defaults to running nothing.

#### `RESULTS_DIR`
RESULTS_DIR specifies a directory where the benchmark results should be written
to. It defaults to `results`

### how to run
First, ensure docker is running, and you have the 'ubuntu' image pulled, then:
```
$ make
$ ./tests/multinode_fetch
```

To run with a simulated network, try:
```
$ export NET_SETUP="ctrlnet lat 30ms 5ms; ctrlnet rate 10mbit"
$ ./tests/multinode_fetch
```

This will run tests on a network where nodes have a 60ms average RTT between
them (5ms jitter). And a bandwidth limit of 10mbit.

