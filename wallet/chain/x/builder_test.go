// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package x

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto/secp256k1"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/avalanchego/vms/avm/config"
	"github.com/ava-labs/avalanchego/vms/avm/txs/fees"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/nftfx"
	"github.com/ava-labs/avalanchego/vms/propertyfx"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/avalanchego/wallet/chain/x/backends"
	"github.com/ava-labs/avalanchego/wallet/subnet/primary/common"

	stdcontext "context"
	commonfees "github.com/ava-labs/avalanchego/vms/components/fees"
)

var (
	testKeys = secp256k1.TestKeys()

	// We hard-code [avaxAssetID] and [subnetAssetID] to make
	// ordering of UTXOs generated by [testUTXOsList] is reproducible
	avaxAssetID     = ids.Empty.Prefix(1789)
	xChainID        = ids.Empty.Prefix(2021)
	nftAssetID      = ids.Empty.Prefix(2022)
	propertyAssetID = ids.Empty.Prefix(2023)

	testCtx = backends.NewContext(
		constants.UnitTestID,
		xChainID,
		avaxAssetID,
		units.MicroAvax,    // BaseTxFee
		99*units.MilliAvax, // CreateAssetTxFee
	)

	testUnitFees = commonfees.Dimensions{
		1 * units.MicroAvax,
		2 * units.MicroAvax,
		3 * units.MicroAvax,
		4 * units.MicroAvax,
	}
	testBlockMaxConsumedUnits = commonfees.Max
)

// These tests create and sign a tx, then verify that utxos included
// in the tx are exactly necessary to pay fees for it

func TestBaseTx(t *testing.T) {
	var (
		require = require.New(t)

		// backend
		utxosKey       = testKeys[1]
		utxos          = makeTestUTXOs(utxosKey)
		genericBackend = common.NewDeterministicChainUTXOs(
			require,
			map[ids.ID][]*avax.UTXO{
				xChainID: utxos,
			},
		)
		backend = NewBackend(testCtx, genericBackend)

		// builder and signer
		utxoAddr = utxosKey.Address()
		builder  = backends.NewBuilder(set.Of(utxoAddr), backend)
		kc       = secp256k1fx.NewKeychain(utxosKey)
		s        = backends.NewSigner(kc, genericBackend)

		// data to build the transaction
		outputsToMove = []*avax.TransferableOutput{{
			Asset: avax.Asset{ID: avaxAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: 7 * units.Avax,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{utxoAddr},
				},
			},
		}}
	)

	{ // Post E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
		}
		utx, err := builder.NewBaseTx(
			outputsToMove,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Credentials:      tx.Creds,
			Codec:            backends.Parser.Codec(),
		}
		require.NoError(utx.Visit(fc))
		require.Equal(9930*units.MicroAvax, fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 2)
		require.Len(outs, 2)

		expectedConsumed := fc.Fee + outputsToMove[0].Out.Amount()
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
		require.Equal(outputsToMove[0], outs[1])
	}

	{ // Pre E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
		}
		utx, err := builder.NewBaseTx(
			outputsToMove,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(testCtx.BaseTxFee(), fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 2)
		require.Len(outs, 2)

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount() - outs[1].Out.Amount()
		require.Equal(expectedConsumed, consumed)
		require.Equal(outputsToMove[0], outs[1])
	}
}

func TestCreateAssetTx(t *testing.T) {
	require := require.New(t)

	var (
		// backend
		utxosKey       = testKeys[1]
		utxos          = makeTestUTXOs(utxosKey)
		genericBackend = common.NewDeterministicChainUTXOs(
			require,
			map[ids.ID][]*avax.UTXO{
				xChainID: utxos,
			},
		)
		backend = NewBackend(testCtx, genericBackend)

		// builder and signer
		utxoAddr = utxosKey.Address()
		builder  = backends.NewBuilder(set.Of(utxoAddr), backend)
		kc       = secp256k1fx.NewKeychain(utxosKey)
		s        = backends.NewSigner(kc, genericBackend)

		// data to build the transaction
		assetName          = "Team Rocket"
		symbol             = "TR"
		denomination uint8 = 0
		initialState       = map[uint32][]verify.State{
			0: {
				&secp256k1fx.MintOutput{
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{testKeys[0].PublicKey().Address()},
					},
				}, &secp256k1fx.MintOutput{
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{testKeys[0].PublicKey().Address()},
					},
				},
			},
			1: {
				&nftfx.MintOutput{
					GroupID: 1,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{testKeys[1].PublicKey().Address()},
					},
				},
				&nftfx.MintOutput{
					GroupID: 2,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{testKeys[1].PublicKey().Address()},
					},
				},
			},
			2: {
				&propertyfx.MintOutput{
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{testKeys[2].PublicKey().Address()},
					},
				},
				&propertyfx.MintOutput{
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{testKeys[2].PublicKey().Address()},
					},
				},
			},
		}
	)

	{
		// Post E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewCreateAssetTx(
			assetName,
			symbol,
			denomination,
			initialState,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(9898*units.MicroAvax, fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 2)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}

	{
		// Pre E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				CreateAssetTxFee: testCtx.CreateAssetTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewCreateAssetTx(
			assetName,
			symbol,
			denomination,
			initialState,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				CreateAssetTxFee: testCtx.CreateAssetTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(99*units.MilliAvax, fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}
}

func TestMintNFTOperation(t *testing.T) {
	require := require.New(t)

	var (
		// backend
		utxosKey       = testKeys[1]
		utxos          = makeTestUTXOs(utxosKey)
		genericBackend = common.NewDeterministicChainUTXOs(
			require,
			map[ids.ID][]*avax.UTXO{
				xChainID: utxos,
			},
		)
		backend = NewBackend(testCtx, genericBackend)

		// builder and signer
		utxoAddr = utxosKey.Address()
		builder  = backends.NewBuilder(set.Of(utxoAddr), backend)
		kc       = secp256k1fx.NewKeychain(utxosKey)
		s        = backends.NewSigner(kc, genericBackend)

		// data to build the transaction
		payload  = []byte{'h', 'e', 'l', 'l', 'o'}
		NFTOwner = &secp256k1fx.OutputOwners{
			Threshold: 1,
			Addrs:     []ids.ShortID{utxoAddr},
		}
	)

	{
		// Post E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewOperationTxMintNFT(
			nftAssetID,
			payload,
			[]*secp256k1fx.OutputOwners{NFTOwner},
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(9818*units.MicroAvax, fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 2)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}

	{
		// Pre E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewOperationTxMintNFT(
			nftAssetID,
			payload,
			[]*secp256k1fx.OutputOwners{NFTOwner},
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(testCtx.BaseTxFee(), fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 1)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}
}

func TestMintFTOperation(t *testing.T) {
	require := require.New(t)

	var (
		// backend
		utxosKey       = testKeys[1]
		utxos          = makeTestUTXOs(utxosKey)
		genericBackend = common.NewDeterministicChainUTXOs(
			require,
			map[ids.ID][]*avax.UTXO{
				xChainID: utxos,
			},
		)
		backend = NewBackend(testCtx, genericBackend)

		// builder and signer
		utxoAddr = utxosKey.Address()
		builder  = backends.NewBuilder(set.Of(utxoAddr), backend)
		kc       = secp256k1fx.NewKeychain(utxosKey)
		s        = backends.NewSigner(kc, genericBackend)

		// data to build the transaction
		outputs = map[ids.ID]*secp256k1fx.TransferOutput{
			nftAssetID: {
				Amt: 1,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{utxoAddr},
				},
			},
		}
	)

	{
		// Post E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewOperationTxMintFT(
			outputs,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(9845*units.MicroAvax, fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 2)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}

	{
		// Pre E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewOperationTxMintFT(
			outputs,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(testCtx.BaseTxFee(), fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 1)
		require.Len(outs, 1)

		expectedConsumed := testCtx.BaseTxFee()
		consumed := ins[0].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}
}

func TestMintPropertyOperation(t *testing.T) {
	require := require.New(t)

	var (
		// backend
		utxosKey       = testKeys[1]
		utxos          = makeTestUTXOs(utxosKey)
		genericBackend = common.NewDeterministicChainUTXOs(
			require,
			map[ids.ID][]*avax.UTXO{
				xChainID: utxos,
			},
		)
		backend = NewBackend(testCtx, genericBackend)

		// builder and signer
		utxoAddr = utxosKey.Address()
		builder  = backends.NewBuilder(set.Of(utxoAddr), backend)
		kc       = secp256k1fx.NewKeychain(utxosKey)
		s        = backends.NewSigner(kc, genericBackend)

		// data to build the transaction
		propertyOwner = &secp256k1fx.OutputOwners{
			Threshold: 1,
			Addrs:     []ids.ShortID{utxoAddr},
		}
	)

	{
		// Post E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewOperationTxMintProperty(
			propertyAssetID,
			propertyOwner,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(9837*units.MicroAvax, fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 2)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}

	{
		// Pre E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewOperationTxMintProperty(
			propertyAssetID,
			propertyOwner,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(testCtx.BaseTxFee(), fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 1)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}
}

func TestBurnPropertyOperation(t *testing.T) {
	require := require.New(t)

	var (
		// backend
		utxosKey       = testKeys[1]
		utxos          = makeTestUTXOs(utxosKey)
		genericBackend = common.NewDeterministicChainUTXOs(
			require,
			map[ids.ID][]*avax.UTXO{
				xChainID: utxos,
			},
		)
		backend = NewBackend(testCtx, genericBackend)

		// builder and signer
		utxoAddr = utxosKey.Address()
		builder  = backends.NewBuilder(set.Of(utxoAddr), backend)
		kc       = secp256k1fx.NewKeychain(utxosKey)
		s        = backends.NewSigner(kc, genericBackend)
	)

	{
		// Post E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewOperationTxBurnProperty(
			propertyAssetID,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(9765*units.MicroAvax, fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 2)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}

	{
		// Pre E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
		}

		utx, err := builder.NewOperationTxBurnProperty(
			propertyAssetID,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(testCtx.BaseTxFee(), fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 1)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := ins[0].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}
}

func TestImportTx(t *testing.T) {
	var (
		require = require.New(t)

		// backend
		utxosKey       = testKeys[1]
		utxos          = makeTestUTXOs(utxosKey)
		sourceChainID  = ids.GenerateTestID()
		importedUTXOs  = utxos[:1]
		genericBackend = common.NewDeterministicChainUTXOs(
			require,
			map[ids.ID][]*avax.UTXO{
				xChainID:      utxos,
				sourceChainID: importedUTXOs,
			},
		)

		backend = NewBackend(testCtx, genericBackend)

		// builder and signer
		utxoAddr = utxosKey.Address()
		builder  = backends.NewBuilder(set.Of(utxoAddr), backend)
		kc       = secp256k1fx.NewKeychain(utxosKey)
		s        = backends.NewSigner(kc, genericBackend)

		// data to build the transaction
		importKey = testKeys[0]
		importTo  = &secp256k1fx.OutputOwners{
			Threshold: 1,
			Addrs: []ids.ShortID{
				importKey.Address(),
			},
		}
	)

	{ // Post E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
		}
		utx, err := builder.NewImportTx(
			sourceChainID,
			importTo,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Credentials:      tx.Creds,
			Codec:            backends.Parser.Codec(),
		}
		require.NoError(utx.Visit(fc))
		require.Equal(14251*units.MicroAvax, fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		importedIns := utx.ImportedIns
		require.Len(ins, 2)
		require.Len(importedIns, 1)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := importedIns[0].In.Amount() + ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}

	{ // Pre E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
		}
		utx, err := builder.NewImportTx(
			sourceChainID,
			importTo,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(testCtx.BaseTxFee(), fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		importedIns := utx.ImportedIns
		require.Empty(ins)
		require.Len(importedIns, 1)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee
		consumed := importedIns[0].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
	}
}

func TestExportTx(t *testing.T) {
	var (
		require = require.New(t)

		// backend
		utxosKey       = testKeys[1]
		utxos          = makeTestUTXOs(utxosKey)
		genericBackend = common.NewDeterministicChainUTXOs(
			require,
			map[ids.ID][]*avax.UTXO{
				xChainID: utxos,
			},
		)
		backend = NewBackend(testCtx, genericBackend)

		// builder and signer
		utxoAddr = utxosKey.Address()
		builder  = backends.NewBuilder(set.Of(utxoAddr), backend)
		kc       = secp256k1fx.NewKeychain(utxosKey)
		s        = backends.NewSigner(kc, genericBackend)

		// data to build the transaction
		subnetID        = ids.GenerateTestID()
		exportedOutputs = []*avax.TransferableOutput{{
			Asset: avax.Asset{ID: avaxAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: 7 * units.Avax,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{utxoAddr},
				},
			},
		}}
	)

	{ // Post E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Codec:            backends.Parser.Codec(),
		}
		utx, err := builder.NewExportTx(
			subnetID,
			exportedOutputs,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: true,
			FeeManager:       commonfees.NewManager(testUnitFees, commonfees.EmptyWindows),
			ConsumedUnitsCap: testBlockMaxConsumedUnits,
			Credentials:      tx.Creds,
			Codec:            backends.Parser.Codec(),
		}
		require.NoError(utx.Visit(fc))
		require.Equal(9966*units.MicroAvax, fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 2)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee + exportedOutputs[0].Out.Amount()
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
		require.Equal(utx.ExportedOuts, exportedOutputs)
	}

	{ // Pre E-Upgrade
		feeCalc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
		}
		utx, err := builder.NewExportTx(
			subnetID,
			exportedOutputs,
			feeCalc,
		)
		require.NoError(err)

		tx, err := backends.SignUnsigned(stdcontext.Background(), s, utx)
		require.NoError(err)

		fc := &fees.Calculator{
			IsEUpgradeActive: false,
			Config: &config.Config{
				TxFee: testCtx.BaseTxFee(),
			},
			FeeManager:       commonfees.NewManager(commonfees.Empty, commonfees.EmptyWindows),
			ConsumedUnitsCap: commonfees.Max,
			Codec:            backends.Parser.Codec(),
			Credentials:      tx.Creds,
		}
		require.NoError(utx.Visit(fc))
		require.Equal(testCtx.BaseTxFee(), fc.Fee)

		// check UTXOs selection and fee financing
		ins := utx.Ins
		outs := utx.Outs
		require.Len(ins, 2)
		require.Len(outs, 1)

		expectedConsumed := fc.Fee + exportedOutputs[0].Out.Amount()
		consumed := ins[0].In.Amount() + ins[1].In.Amount() - outs[0].Out.Amount()
		require.Equal(expectedConsumed, consumed)
		require.Equal(utx.ExportedOuts, exportedOutputs)
	}
}

func makeTestUTXOs(utxosKey *secp256k1.PrivateKey) []*avax.UTXO {
	// Note: we avoid ids.GenerateTestNodeID here to make sure that UTXO IDs won't change
	// run by run. This simplifies checking what utxos are included in the built txs.
	const utxosOffset uint64 = 2024

	return []*avax.UTXO{ // currently, the wallet scans UTXOs in the order provided here
		{ // a small UTXO first, which  should not be enough to pay fees
			UTXOID: avax.UTXOID{
				TxID:        ids.Empty.Prefix(utxosOffset),
				OutputIndex: uint32(utxosOffset),
			},
			Asset: avax.Asset{ID: avaxAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: 2 * units.MilliAvax,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Addrs:     []ids.ShortID{utxosKey.PublicKey().Address()},
					Threshold: 1,
				},
			},
		},
		{
			UTXOID: avax.UTXOID{
				TxID:        ids.Empty.Prefix(utxosOffset + 2),
				OutputIndex: uint32(utxosOffset + 2),
			},
			Asset: avax.Asset{ID: nftAssetID},
			Out: &nftfx.MintOutput{
				GroupID: 1,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{utxosKey.PublicKey().Address()},
				},
			},
		},
		{
			UTXOID: avax.UTXOID{
				TxID:        ids.Empty.Prefix(utxosOffset + 3),
				OutputIndex: uint32(utxosOffset + 3),
			},
			Asset: avax.Asset{ID: nftAssetID},
			Out: &secp256k1fx.MintOutput{
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{utxosKey.PublicKey().Address()},
				},
			},
		},
		{
			UTXOID: avax.UTXOID{
				TxID:        ids.Empty.Prefix(utxosOffset + 4),
				OutputIndex: uint32(utxosOffset + 4),
			},
			Asset: avax.Asset{ID: propertyAssetID},
			Out: &propertyfx.MintOutput{
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Addrs:     []ids.ShortID{utxosKey.PublicKey().Address()},
					Threshold: 1,
				},
			},
		},
		{
			UTXOID: avax.UTXOID{
				TxID:        ids.Empty.Prefix(utxosOffset + 5),
				OutputIndex: uint32(utxosOffset + 5),
			},
			Asset: avax.Asset{ID: propertyAssetID},
			Out: &propertyfx.OwnedOutput{
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Addrs:     []ids.ShortID{utxosKey.PublicKey().Address()},
					Threshold: 1,
				},
			},
		},
		{ // a large UTXO last, which should be enough to pay any fee by itself
			UTXOID: avax.UTXOID{
				TxID:        ids.Empty.Prefix(utxosOffset + 6),
				OutputIndex: uint32(utxosOffset + 6),
			},
			Asset: avax.Asset{ID: avaxAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: 9 * units.Avax,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Addrs:     []ids.ShortID{utxosKey.PublicKey().Address()},
					Threshold: 1,
				},
			},
		},
	}
}
