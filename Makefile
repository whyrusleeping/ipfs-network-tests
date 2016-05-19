all: docker nettest

docker: deps
	docker build --tag=ipfs-node .

deps: bwcurl rand ipfs

ipfs: 
	go get github.com/ipfs/go-ipfs
	mkdir -p bin
	go build -o bin/ipfs github.com/ipfs/go-ipfs/cmd/ipfs

bwcurl: utils/bwcurl/main.go
	go build -o bin/bwcurl utils/bwcurl/main.go

rand: utils/rand/main.go
	go get github.com/jbenet/go-random-files/random-files
	go build -o bin/randfiles github.com/jbenet/go-random-files/random-files

nettest: nettest.go
	go build
