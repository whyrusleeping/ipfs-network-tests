#!/bin/bash
ipfs init
ipfs config --json Bootstrap []
ipfs daemon
