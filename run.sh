#!/bin/bash

# ./scripts/build.sh

mkdir -p $HOME/.avalanchego/configs/chains/C
echo '{
  "state-sync-enabled": false,
  "pruning-enabled": false,
  "allow-missing-tries": false
}' > $HOME/.avalanchego/configs/chains/C/config.json


rm -rf $HOME/.avagotemp


mkdir -p $HOME/.avagotemp


go run ./main/ --staking-port 9951 --public-ip 52.192.174.252 --data-dir $HOME/.avagotemp
