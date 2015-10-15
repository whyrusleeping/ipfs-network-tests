docker:
	docker build --tag=ipfs-node .

ipfs: 
	mkdir -p bin
	go build -o bin/ipfs github.com/ipfs/go-ipfs/cmd/ipfs

deps: bwcurl

bwcurl: utils/bwcurl/main.go
	go build -o bin/bwcurl utils/bwcurl/main.go
