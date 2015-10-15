#!/bin/bash

A=$(docker run -d ipfs-node)
B=$(docker run -d ipfs-node)

./tests/transfer_file $A $B 100

docker kill $A
docker kill $B
