// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import "github.com/ava-labs/avalanchego/vms/platformvm/fees"

type FeeConfig struct {
	DefaultUnitPrices fees.Dimensions

	// Pre E Fork, fee that is burned by every non-state creating transaction
	TxFee uint64

	// Pre E Fork, fee that must be burned by every state creating transaction before AP3
	CreateAssetTxFee uint64

	// Pre E Fork, fee that must be burned by every subnet creating transaction after AP3
	CreateSubnetTxFee uint64

	// Pre E Fork, fee that must be burned by every transform subnet transaction
	TransformSubnetTxFee uint64

	// Pre E Fork, fee that must be burned by every blockchain creating transaction after AP3
	CreateBlockchainTxFee uint64

	// Pre E Fork, transaction fee for adding a primary network validator
	AddPrimaryNetworkValidatorFee uint64

	// Pre E Fork, transaction fee for adding a primary network delegator
	AddPrimaryNetworkDelegatorFee uint64

	// Pre E Fork, transaction fee for adding a subnet validator
	AddSubnetValidatorFee uint64

	// Pre E Fork, transaction fee for adding a subnet delegator
	AddSubnetDelegatorFee uint64
}
