package signertest

import (
	bls "github.com/ava-labs/avalanchego/utils/crypto/bls"
	blst "github.com/supranational/blst/bindings/go"

)

type Signer {
	key *blst.SecretKey
}

func new