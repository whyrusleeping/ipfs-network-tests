
connect_nodes() {
	addr=$(docker exec $1 /bin/ipfs id -f "<addrs>" | grep 172.17)
	if [ -z $addr ]
	then
		echo "no addresses found on ipfsnode: $1"
		exit 1
	fi
	docker exec $2 /bin/ipfs swarm connect $addr
}

addfile() {
	ctnr=$1
	size=$2
	docker exec $ctnr /bin/addfile $size
}

catfile() {
	ctnr=$1
	file=$2
	docker exec $ctnr /bin/bwcurl http://localhost:8080/ipfs/$file
}

