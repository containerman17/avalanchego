// Copyright (C) 2019, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chains

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/hashing"
)

var _ validators.State = (*IsolatedL1)(nil)

// IsolatedL1 runs one chain without the P-chain. It is also the static
// validators.State that stands in for the P-chain, frozen at Height.
type IsolatedL1 struct {
	Chain      ChainParameters
	Validators map[ids.NodeID]*validators.GetValidatorOutput
	Height     uint64
}

func (l *IsolatedL1) GetMinimumHeight(context.Context) (uint64, error) {
	return l.Height, nil
}

func (l *IsolatedL1) GetCurrentHeight(context.Context) (uint64, error) {
	return l.Height, nil
}

func (l *IsolatedL1) GetSubnetID(_ context.Context, chainID ids.ID) (ids.ID, error) {
	switch chainID {
	case l.Chain.ID:
		return l.Chain.SubnetID, nil
	case constants.PlatformChainID:
		return constants.PrimaryNetworkID, nil
	}
	return ids.Empty, fmt.Errorf("isolated L1: unknown chain %s", chainID)
}

func (l *IsolatedL1) GetWarpValidatorSets(context.Context, uint64) (map[ids.ID]validators.WarpSet, error) {
	ws, err := validators.FlattenValidatorSet(l.Validators)
	if err != nil {
		return nil, err
	}
	return map[ids.ID]validators.WarpSet{l.Chain.SubnetID: ws, constants.PrimaryNetworkID: ws}, nil
}

func (l *IsolatedL1) GetValidatorSet(_ context.Context, _ uint64, subnetID ids.ID) (map[ids.NodeID]*validators.GetValidatorOutput, error) {
	if subnetID != l.Chain.SubnetID && subnetID != constants.PrimaryNetworkID {
		return nil, fmt.Errorf("isolated L1: unknown subnet %s", subnetID)
	}
	return l.Validators, nil
}

func (l *IsolatedL1) GetCurrentValidatorSet(ctx context.Context, subnetID ids.ID) (map[ids.ID]*validators.GetCurrentValidatorOutput, uint64, error) {
	vdrs, err := l.GetValidatorSet(ctx, 0, subnetID)
	if err != nil {
		return nil, 0, err
	}
	out := make(map[ids.ID]*validators.GetCurrentValidatorOutput, len(vdrs))
	for nodeID, vdr := range vdrs {
		// Deterministic stand-in: there is no P-chain validation ID yet.
		validationID := ids.ID(hashing.ComputeHash256Array(nodeID[:]))
		out[validationID] = &validators.GetCurrentValidatorOutput{
			ValidationID: validationID,
			NodeID:       nodeID,
			PublicKey:    vdr.PublicKey,
			Weight:       vdr.Weight,
			IsActive:     true,
		}
	}
	return out, l.Height, nil
}
