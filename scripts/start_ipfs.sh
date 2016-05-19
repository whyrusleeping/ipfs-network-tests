#!/bin/bash
ipfs init
ipfs config --json Bootstrap []
ipfs config --json Discovery.MDNS.Enabled false
ipfs config --json Datastore.NoSync false
ipfs daemon
