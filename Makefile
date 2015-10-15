all: docker

docker: deps
	docker build --tag=ipfs-node .

deps: bwcurl ipfs

ipfs: 
	mkdir -p bin
	go build -o bin/ipfs github.com/ipfs/go-ipfs/cmd/ipfs

bwcurl: utils/bwcurl/main.go
	go build -o bin/bwcurl utils/bwcurl/main.go
