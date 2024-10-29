// Code generated by MockGen. DO NOT EDIT.
// Source: vms/platformvm/state/state.go
//
// Generated by this command:
//
//	mockgen -source=vms/platformvm/state/state.go -destination=vms/platformvm/state/mock_state.go -package=state -exclude_interfaces=Chain -mock_names=MockState=MockState
//

// Package state is a generated GoMock package.
package state

import (
	context "context"
	reflect "reflect"
	sync "sync"
	time "time"

	database "github.com/ava-labs/avalanchego/database"
	ids "github.com/ava-labs/avalanchego/ids"
	validators "github.com/ava-labs/avalanchego/snow/validators"
	iterator "github.com/ava-labs/avalanchego/utils/iterator"
	logging "github.com/ava-labs/avalanchego/utils/logging"
	avax "github.com/ava-labs/avalanchego/vms/components/avax"
	gas "github.com/ava-labs/avalanchego/vms/components/gas"
	block "github.com/ava-labs/avalanchego/vms/platformvm/block"
	fx "github.com/ava-labs/avalanchego/vms/platformvm/fx"
	status "github.com/ava-labs/avalanchego/vms/platformvm/status"
	txs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
	gomock "go.uber.org/mock/gomock"
)

// MockState is a mock of State interface.
type MockState struct {
	ctrl     *gomock.Controller
	recorder *MockStateMockRecorder
}

// MockStateMockRecorder is the mock recorder for MockState.
type MockStateMockRecorder struct {
	mock *MockState
}

// NewMockState creates a new mock instance.
func NewMockState(ctrl *gomock.Controller) *MockState {
	mock := &MockState{ctrl: ctrl}
	mock.recorder = &MockStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockState) EXPECT() *MockStateMockRecorder {
	return m.recorder
}

// Abort mocks base method.
func (m *MockState) Abort() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Abort")
}

// Abort indicates an expected call of Abort.
func (mr *MockStateMockRecorder) Abort() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Abort", reflect.TypeOf((*MockState)(nil).Abort))
}

// AddChain mocks base method.
func (m *MockState) AddChain(createChainTx *txs.Tx) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddChain", createChainTx)
}

// AddChain indicates an expected call of AddChain.
func (mr *MockStateMockRecorder) AddChain(createChainTx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddChain", reflect.TypeOf((*MockState)(nil).AddChain), createChainTx)
}

// AddRewardUTXO mocks base method.
func (m *MockState) AddRewardUTXO(txID ids.ID, utxo *avax.UTXO) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddRewardUTXO", txID, utxo)
}

// AddRewardUTXO indicates an expected call of AddRewardUTXO.
func (mr *MockStateMockRecorder) AddRewardUTXO(txID, utxo any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddRewardUTXO", reflect.TypeOf((*MockState)(nil).AddRewardUTXO), txID, utxo)
}

// AddStatelessBlock mocks base method.
func (m *MockState) AddStatelessBlock(block block.Block) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddStatelessBlock", block)
}

// AddStatelessBlock indicates an expected call of AddStatelessBlock.
func (mr *MockStateMockRecorder) AddStatelessBlock(block any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddStatelessBlock", reflect.TypeOf((*MockState)(nil).AddStatelessBlock), block)
}

// AddSubnet mocks base method.
func (m *MockState) AddSubnet(subnetID ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddSubnet", subnetID)
}

// AddSubnet indicates an expected call of AddSubnet.
func (mr *MockStateMockRecorder) AddSubnet(subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSubnet", reflect.TypeOf((*MockState)(nil).AddSubnet), subnetID)
}

// AddSubnetTransformation mocks base method.
func (m *MockState) AddSubnetTransformation(transformSubnetTx *txs.Tx) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddSubnetTransformation", transformSubnetTx)
}

// AddSubnetTransformation indicates an expected call of AddSubnetTransformation.
func (mr *MockStateMockRecorder) AddSubnetTransformation(transformSubnetTx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSubnetTransformation", reflect.TypeOf((*MockState)(nil).AddSubnetTransformation), transformSubnetTx)
}

// AddTx mocks base method.
func (m *MockState) AddTx(tx *txs.Tx, status status.Status) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddTx", tx, status)
}

// AddTx indicates an expected call of AddTx.
func (mr *MockStateMockRecorder) AddTx(tx, status any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTx", reflect.TypeOf((*MockState)(nil).AddTx), tx, status)
}

// AddUTXO mocks base method.
func (m *MockState) AddUTXO(utxo *avax.UTXO) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddUTXO", utxo)
}

// AddUTXO indicates an expected call of AddUTXO.
func (mr *MockStateMockRecorder) AddUTXO(utxo any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddUTXO", reflect.TypeOf((*MockState)(nil).AddUTXO), utxo)
}

// ApplyValidatorPublicKeyDiffs mocks base method.
func (m *MockState) ApplyValidatorPublicKeyDiffs(ctx context.Context, validators map[ids.NodeID]*validators.GetValidatorOutput, startHeight, endHeight uint64, subnetID ids.ID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyValidatorPublicKeyDiffs", ctx, validators, startHeight, endHeight, subnetID)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyValidatorPublicKeyDiffs indicates an expected call of ApplyValidatorPublicKeyDiffs.
func (mr *MockStateMockRecorder) ApplyValidatorPublicKeyDiffs(ctx, validators, startHeight, endHeight, subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyValidatorPublicKeyDiffs", reflect.TypeOf((*MockState)(nil).ApplyValidatorPublicKeyDiffs), ctx, validators, startHeight, endHeight, subnetID)
}

// ApplyValidatorWeightDiffs mocks base method.
func (m *MockState) ApplyValidatorWeightDiffs(ctx context.Context, validators map[ids.NodeID]*validators.GetValidatorOutput, startHeight, endHeight uint64, subnetID ids.ID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyValidatorWeightDiffs", ctx, validators, startHeight, endHeight, subnetID)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyValidatorWeightDiffs indicates an expected call of ApplyValidatorWeightDiffs.
func (mr *MockStateMockRecorder) ApplyValidatorWeightDiffs(ctx, validators, startHeight, endHeight, subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyValidatorWeightDiffs", reflect.TypeOf((*MockState)(nil).ApplyValidatorWeightDiffs), ctx, validators, startHeight, endHeight, subnetID)
}

// Checksum mocks base method.
func (m *MockState) Checksum() ids.ID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Checksum")
	ret0, _ := ret[0].(ids.ID)
	return ret0
}

// Checksum indicates an expected call of Checksum.
func (mr *MockStateMockRecorder) Checksum() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Checksum", reflect.TypeOf((*MockState)(nil).Checksum))
}

// Close mocks base method.
func (m *MockState) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStateMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockState)(nil).Close))
}

// Commit mocks base method.
func (m *MockState) Commit() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit.
func (mr *MockStateMockRecorder) Commit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockState)(nil).Commit))
}

// CommitBatch mocks base method.
func (m *MockState) CommitBatch() (database.Batch, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitBatch")
	ret0, _ := ret[0].(database.Batch)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CommitBatch indicates an expected call of CommitBatch.
func (mr *MockStateMockRecorder) CommitBatch() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitBatch", reflect.TypeOf((*MockState)(nil).CommitBatch))
}

// DeleteCurrentDelegator mocks base method.
func (m *MockState) DeleteCurrentDelegator(staker *Staker) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteCurrentDelegator", staker)
}

// DeleteCurrentDelegator indicates an expected call of DeleteCurrentDelegator.
func (mr *MockStateMockRecorder) DeleteCurrentDelegator(staker any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCurrentDelegator", reflect.TypeOf((*MockState)(nil).DeleteCurrentDelegator), staker)
}

// DeleteCurrentValidator mocks base method.
func (m *MockState) DeleteCurrentValidator(staker *Staker) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteCurrentValidator", staker)
}

// DeleteCurrentValidator indicates an expected call of DeleteCurrentValidator.
func (mr *MockStateMockRecorder) DeleteCurrentValidator(staker any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCurrentValidator", reflect.TypeOf((*MockState)(nil).DeleteCurrentValidator), staker)
}

// DeleteExpiry mocks base method.
func (m *MockState) DeleteExpiry(arg0 ExpiryEntry) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteExpiry", arg0)
}

// DeleteExpiry indicates an expected call of DeleteExpiry.
func (mr *MockStateMockRecorder) DeleteExpiry(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteExpiry", reflect.TypeOf((*MockState)(nil).DeleteExpiry), arg0)
}

// DeletePendingDelegator mocks base method.
func (m *MockState) DeletePendingDelegator(staker *Staker) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeletePendingDelegator", staker)
}

// DeletePendingDelegator indicates an expected call of DeletePendingDelegator.
func (mr *MockStateMockRecorder) DeletePendingDelegator(staker any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePendingDelegator", reflect.TypeOf((*MockState)(nil).DeletePendingDelegator), staker)
}

// DeletePendingValidator mocks base method.
func (m *MockState) DeletePendingValidator(staker *Staker) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeletePendingValidator", staker)
}

// DeletePendingValidator indicates an expected call of DeletePendingValidator.
func (mr *MockStateMockRecorder) DeletePendingValidator(staker any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePendingValidator", reflect.TypeOf((*MockState)(nil).DeletePendingValidator), staker)
}

// DeleteUTXO mocks base method.
func (m *MockState) DeleteUTXO(utxoID ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteUTXO", utxoID)
}

// DeleteUTXO indicates an expected call of DeleteUTXO.
func (mr *MockStateMockRecorder) DeleteUTXO(utxoID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUTXO", reflect.TypeOf((*MockState)(nil).DeleteUTXO), utxoID)
}

// GetAccruedFees mocks base method.
func (m *MockState) GetAccruedFees() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccruedFees")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetAccruedFees indicates an expected call of GetAccruedFees.
func (mr *MockStateMockRecorder) GetAccruedFees() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccruedFees", reflect.TypeOf((*MockState)(nil).GetAccruedFees))
}

// GetActiveSubnetOnlyValidatorsIterator mocks base method.
func (m *MockState) GetActiveSubnetOnlyValidatorsIterator() (iterator.Iterator[SubnetOnlyValidator], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetActiveSubnetOnlyValidatorsIterator")
	ret0, _ := ret[0].(iterator.Iterator[SubnetOnlyValidator])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetActiveSubnetOnlyValidatorsIterator indicates an expected call of GetActiveSubnetOnlyValidatorsIterator.
func (mr *MockStateMockRecorder) GetActiveSubnetOnlyValidatorsIterator() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActiveSubnetOnlyValidatorsIterator", reflect.TypeOf((*MockState)(nil).GetActiveSubnetOnlyValidatorsIterator))
}

// GetBlockIDAtHeight mocks base method.
func (m *MockState) GetBlockIDAtHeight(height uint64) (ids.ID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockIDAtHeight", height)
	ret0, _ := ret[0].(ids.ID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockIDAtHeight indicates an expected call of GetBlockIDAtHeight.
func (mr *MockStateMockRecorder) GetBlockIDAtHeight(height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockIDAtHeight", reflect.TypeOf((*MockState)(nil).GetBlockIDAtHeight), height)
}

// GetChains mocks base method.
func (m *MockState) GetChains(subnetID ids.ID) ([]*txs.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChains", subnetID)
	ret0, _ := ret[0].([]*txs.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetChains indicates an expected call of GetChains.
func (mr *MockStateMockRecorder) GetChains(subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChains", reflect.TypeOf((*MockState)(nil).GetChains), subnetID)
}

// GetCurrentDelegatorIterator mocks base method.
func (m *MockState) GetCurrentDelegatorIterator(subnetID ids.ID, nodeID ids.NodeID) (iterator.Iterator[*Staker], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentDelegatorIterator", subnetID, nodeID)
	ret0, _ := ret[0].(iterator.Iterator[*Staker])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCurrentDelegatorIterator indicates an expected call of GetCurrentDelegatorIterator.
func (mr *MockStateMockRecorder) GetCurrentDelegatorIterator(subnetID, nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentDelegatorIterator", reflect.TypeOf((*MockState)(nil).GetCurrentDelegatorIterator), subnetID, nodeID)
}

// GetCurrentStakerIterator mocks base method.
func (m *MockState) GetCurrentStakerIterator() (iterator.Iterator[*Staker], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentStakerIterator")
	ret0, _ := ret[0].(iterator.Iterator[*Staker])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCurrentStakerIterator indicates an expected call of GetCurrentStakerIterator.
func (mr *MockStateMockRecorder) GetCurrentStakerIterator() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentStakerIterator", reflect.TypeOf((*MockState)(nil).GetCurrentStakerIterator))
}

// GetCurrentSupply mocks base method.
func (m *MockState) GetCurrentSupply(subnetID ids.ID) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentSupply", subnetID)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCurrentSupply indicates an expected call of GetCurrentSupply.
func (mr *MockStateMockRecorder) GetCurrentSupply(subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentSupply", reflect.TypeOf((*MockState)(nil).GetCurrentSupply), subnetID)
}

// GetCurrentValidator mocks base method.
func (m *MockState) GetCurrentValidator(subnetID ids.ID, nodeID ids.NodeID) (*Staker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentValidator", subnetID, nodeID)
	ret0, _ := ret[0].(*Staker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCurrentValidator indicates an expected call of GetCurrentValidator.
func (mr *MockStateMockRecorder) GetCurrentValidator(subnetID, nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentValidator", reflect.TypeOf((*MockState)(nil).GetCurrentValidator), subnetID, nodeID)
}

// GetCurrentValidatorSet mocks base method.
func (m *MockState) GetCurrentValidatorSet(ctx context.Context, subnetID ids.ID) (map[ids.ID]*validators.GetCurrentValidatorOutput, uint64, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentValidatorSet", ctx, subnetID)
	ret0, _ := ret[0].(map[ids.ID]*validators.GetCurrentValidatorOutput)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(bool)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// GetCurrentValidatorSet indicates an expected call of GetCurrentValidatorSet.
func (mr *MockStateMockRecorder) GetCurrentValidatorSet(ctx, subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentValidatorSet", reflect.TypeOf((*MockState)(nil).GetCurrentValidatorSet), ctx, subnetID)
}

// GetDelegateeReward mocks base method.
func (m *MockState) GetDelegateeReward(subnetID ids.ID, nodeID ids.NodeID) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDelegateeReward", subnetID, nodeID)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDelegateeReward indicates an expected call of GetDelegateeReward.
func (mr *MockStateMockRecorder) GetDelegateeReward(subnetID, nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDelegateeReward", reflect.TypeOf((*MockState)(nil).GetDelegateeReward), subnetID, nodeID)
}

// GetExpiryIterator mocks base method.
func (m *MockState) GetExpiryIterator() (iterator.Iterator[ExpiryEntry], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExpiryIterator")
	ret0, _ := ret[0].(iterator.Iterator[ExpiryEntry])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExpiryIterator indicates an expected call of GetExpiryIterator.
func (mr *MockStateMockRecorder) GetExpiryIterator() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExpiryIterator", reflect.TypeOf((*MockState)(nil).GetExpiryIterator))
}

// GetFeeState mocks base method.
func (m *MockState) GetFeeState() gas.State {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFeeState")
	ret0, _ := ret[0].(gas.State)
	return ret0
}

// GetFeeState indicates an expected call of GetFeeState.
func (mr *MockStateMockRecorder) GetFeeState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFeeState", reflect.TypeOf((*MockState)(nil).GetFeeState))
}

// GetLastAccepted mocks base method.
func (m *MockState) GetLastAccepted() ids.ID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLastAccepted")
	ret0, _ := ret[0].(ids.ID)
	return ret0
}

// GetLastAccepted indicates an expected call of GetLastAccepted.
func (mr *MockStateMockRecorder) GetLastAccepted() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLastAccepted", reflect.TypeOf((*MockState)(nil).GetLastAccepted))
}

// GetPendingDelegatorIterator mocks base method.
func (m *MockState) GetPendingDelegatorIterator(subnetID ids.ID, nodeID ids.NodeID) (iterator.Iterator[*Staker], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPendingDelegatorIterator", subnetID, nodeID)
	ret0, _ := ret[0].(iterator.Iterator[*Staker])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPendingDelegatorIterator indicates an expected call of GetPendingDelegatorIterator.
func (mr *MockStateMockRecorder) GetPendingDelegatorIterator(subnetID, nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPendingDelegatorIterator", reflect.TypeOf((*MockState)(nil).GetPendingDelegatorIterator), subnetID, nodeID)
}

// GetPendingStakerIterator mocks base method.
func (m *MockState) GetPendingStakerIterator() (iterator.Iterator[*Staker], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPendingStakerIterator")
	ret0, _ := ret[0].(iterator.Iterator[*Staker])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPendingStakerIterator indicates an expected call of GetPendingStakerIterator.
func (mr *MockStateMockRecorder) GetPendingStakerIterator() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPendingStakerIterator", reflect.TypeOf((*MockState)(nil).GetPendingStakerIterator))
}

// GetPendingValidator mocks base method.
func (m *MockState) GetPendingValidator(subnetID ids.ID, nodeID ids.NodeID) (*Staker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPendingValidator", subnetID, nodeID)
	ret0, _ := ret[0].(*Staker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPendingValidator indicates an expected call of GetPendingValidator.
func (mr *MockStateMockRecorder) GetPendingValidator(subnetID, nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPendingValidator", reflect.TypeOf((*MockState)(nil).GetPendingValidator), subnetID, nodeID)
}

// GetRewardUTXOs mocks base method.
func (m *MockState) GetRewardUTXOs(txID ids.ID) ([]*avax.UTXO, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRewardUTXOs", txID)
	ret0, _ := ret[0].([]*avax.UTXO)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRewardUTXOs indicates an expected call of GetRewardUTXOs.
func (mr *MockStateMockRecorder) GetRewardUTXOs(txID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRewardUTXOs", reflect.TypeOf((*MockState)(nil).GetRewardUTXOs), txID)
}

// GetSoVExcess mocks base method.
func (m *MockState) GetSoVExcess() gas.Gas {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSoVExcess")
	ret0, _ := ret[0].(gas.Gas)
	return ret0
}

// GetSoVExcess indicates an expected call of GetSoVExcess.
func (mr *MockStateMockRecorder) GetSoVExcess() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSoVExcess", reflect.TypeOf((*MockState)(nil).GetSoVExcess))
}

// GetStartTime mocks base method.
func (m *MockState) GetStartTime(nodeID ids.NodeID) (time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStartTime", nodeID)
	ret0, _ := ret[0].(time.Time)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStartTime indicates an expected call of GetStartTime.
func (mr *MockStateMockRecorder) GetStartTime(nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStartTime", reflect.TypeOf((*MockState)(nil).GetStartTime), nodeID)
}

// GetStatelessBlock mocks base method.
func (m *MockState) GetStatelessBlock(blockID ids.ID) (block.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatelessBlock", blockID)
	ret0, _ := ret[0].(block.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatelessBlock indicates an expected call of GetStatelessBlock.
func (mr *MockStateMockRecorder) GetStatelessBlock(blockID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatelessBlock", reflect.TypeOf((*MockState)(nil).GetStatelessBlock), blockID)
}

// GetSubnetConversion mocks base method.
func (m *MockState) GetSubnetConversion(subnetID ids.ID) (SubnetConversion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSubnetConversion", subnetID)
	ret0, _ := ret[0].(SubnetConversion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubnetConversion indicates an expected call of GetSubnetConversion.
func (mr *MockStateMockRecorder) GetSubnetConversion(subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubnetConversion", reflect.TypeOf((*MockState)(nil).GetSubnetConversion), subnetID)
}

// GetSubnetIDs mocks base method.
func (m *MockState) GetSubnetIDs() ([]ids.ID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSubnetIDs")
	ret0, _ := ret[0].([]ids.ID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubnetIDs indicates an expected call of GetSubnetIDs.
func (mr *MockStateMockRecorder) GetSubnetIDs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubnetIDs", reflect.TypeOf((*MockState)(nil).GetSubnetIDs))
}

// GetSubnetOwner mocks base method.
func (m *MockState) GetSubnetOwner(subnetID ids.ID) (fx.Owner, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSubnetOwner", subnetID)
	ret0, _ := ret[0].(fx.Owner)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubnetOwner indicates an expected call of GetSubnetOwner.
func (mr *MockStateMockRecorder) GetSubnetOwner(subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubnetOwner", reflect.TypeOf((*MockState)(nil).GetSubnetOwner), subnetID)
}

// GetSubnetTransformation mocks base method.
func (m *MockState) GetSubnetTransformation(subnetID ids.ID) (*txs.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSubnetTransformation", subnetID)
	ret0, _ := ret[0].(*txs.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubnetTransformation indicates an expected call of GetSubnetTransformation.
func (mr *MockStateMockRecorder) GetSubnetTransformation(subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubnetTransformation", reflect.TypeOf((*MockState)(nil).GetSubnetTransformation), subnetID)
}

// GetTimestamp mocks base method.
func (m *MockState) GetTimestamp() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTimestamp")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetTimestamp indicates an expected call of GetTimestamp.
func (mr *MockStateMockRecorder) GetTimestamp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTimestamp", reflect.TypeOf((*MockState)(nil).GetTimestamp))
}

// GetTx mocks base method.
func (m *MockState) GetTx(txID ids.ID) (*txs.Tx, status.Status, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTx", txID)
	ret0, _ := ret[0].(*txs.Tx)
	ret1, _ := ret[1].(status.Status)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetTx indicates an expected call of GetTx.
func (mr *MockStateMockRecorder) GetTx(txID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTx", reflect.TypeOf((*MockState)(nil).GetTx), txID)
}

// GetUTXO mocks base method.
func (m *MockState) GetUTXO(utxoID ids.ID) (*avax.UTXO, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUTXO", utxoID)
	ret0, _ := ret[0].(*avax.UTXO)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUTXO indicates an expected call of GetUTXO.
func (mr *MockStateMockRecorder) GetUTXO(utxoID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUTXO", reflect.TypeOf((*MockState)(nil).GetUTXO), utxoID)
}

// GetUptime mocks base method.
func (m *MockState) GetUptime(nodeID ids.NodeID) (time.Duration, time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUptime", nodeID)
	ret0, _ := ret[0].(time.Duration)
	ret1, _ := ret[1].(time.Time)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetUptime indicates an expected call of GetUptime.
func (mr *MockStateMockRecorder) GetUptime(nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUptime", reflect.TypeOf((*MockState)(nil).GetUptime), nodeID)
}

// HasExpiry mocks base method.
func (m *MockState) HasExpiry(arg0 ExpiryEntry) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasExpiry", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasExpiry indicates an expected call of HasExpiry.
func (mr *MockStateMockRecorder) HasExpiry(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasExpiry", reflect.TypeOf((*MockState)(nil).HasExpiry), arg0)
}

// HasSubnetOnlyValidator mocks base method.
func (m *MockState) HasSubnetOnlyValidator(subnetID ids.ID, nodeID ids.NodeID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasSubnetOnlyValidator", subnetID, nodeID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasSubnetOnlyValidator indicates an expected call of HasSubnetOnlyValidator.
func (mr *MockStateMockRecorder) HasSubnetOnlyValidator(subnetID, nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasSubnetOnlyValidator", reflect.TypeOf((*MockState)(nil).HasSubnetOnlyValidator), subnetID, nodeID)
}

// NumActiveSubnetOnlyValidators mocks base method.
func (m *MockState) NumActiveSubnetOnlyValidators() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NumActiveSubnetOnlyValidators")
	ret0, _ := ret[0].(int)
	return ret0
}

// NumActiveSubnetOnlyValidators indicates an expected call of NumActiveSubnetOnlyValidators.
func (mr *MockStateMockRecorder) NumActiveSubnetOnlyValidators() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NumActiveSubnetOnlyValidators", reflect.TypeOf((*MockState)(nil).NumActiveSubnetOnlyValidators))
}

// PutCurrentDelegator mocks base method.
func (m *MockState) PutCurrentDelegator(staker *Staker) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PutCurrentDelegator", staker)
}

// PutCurrentDelegator indicates an expected call of PutCurrentDelegator.
func (mr *MockStateMockRecorder) PutCurrentDelegator(staker any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutCurrentDelegator", reflect.TypeOf((*MockState)(nil).PutCurrentDelegator), staker)
}

// PutCurrentValidator mocks base method.
func (m *MockState) PutCurrentValidator(staker *Staker) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutCurrentValidator", staker)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutCurrentValidator indicates an expected call of PutCurrentValidator.
func (mr *MockStateMockRecorder) PutCurrentValidator(staker any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutCurrentValidator", reflect.TypeOf((*MockState)(nil).PutCurrentValidator), staker)
}

// PutExpiry mocks base method.
func (m *MockState) PutExpiry(arg0 ExpiryEntry) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PutExpiry", arg0)
}

// PutExpiry indicates an expected call of PutExpiry.
func (mr *MockStateMockRecorder) PutExpiry(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutExpiry", reflect.TypeOf((*MockState)(nil).PutExpiry), arg0)
}

// PutPendingDelegator mocks base method.
func (m *MockState) PutPendingDelegator(staker *Staker) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PutPendingDelegator", staker)
}

// PutPendingDelegator indicates an expected call of PutPendingDelegator.
func (mr *MockStateMockRecorder) PutPendingDelegator(staker any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutPendingDelegator", reflect.TypeOf((*MockState)(nil).PutPendingDelegator), staker)
}

// PutPendingValidator mocks base method.
func (m *MockState) PutPendingValidator(staker *Staker) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutPendingValidator", staker)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutPendingValidator indicates an expected call of PutPendingValidator.
func (mr *MockStateMockRecorder) PutPendingValidator(staker any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutPendingValidator", reflect.TypeOf((*MockState)(nil).PutPendingValidator), staker)
}

// PutSubnetOnlyValidator mocks base method.
func (m *MockState) PutSubnetOnlyValidator(sov SubnetOnlyValidator) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutSubnetOnlyValidator", sov)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutSubnetOnlyValidator indicates an expected call of PutSubnetOnlyValidator.
func (mr *MockStateMockRecorder) PutSubnetOnlyValidator(sov any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutSubnetOnlyValidator", reflect.TypeOf((*MockState)(nil).PutSubnetOnlyValidator), sov)
}

// ReindexBlocks mocks base method.
func (m *MockState) ReindexBlocks(lock sync.Locker, log logging.Logger) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReindexBlocks", lock, log)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReindexBlocks indicates an expected call of ReindexBlocks.
func (mr *MockStateMockRecorder) ReindexBlocks(lock, log any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReindexBlocks", reflect.TypeOf((*MockState)(nil).ReindexBlocks), lock, log)
}

// SetAccruedFees mocks base method.
func (m *MockState) SetAccruedFees(f uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetAccruedFees", f)
}

// SetAccruedFees indicates an expected call of SetAccruedFees.
func (mr *MockStateMockRecorder) SetAccruedFees(f any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAccruedFees", reflect.TypeOf((*MockState)(nil).SetAccruedFees), f)
}

// SetCurrentSupply mocks base method.
func (m *MockState) SetCurrentSupply(subnetID ids.ID, cs uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetCurrentSupply", subnetID, cs)
}

// SetCurrentSupply indicates an expected call of SetCurrentSupply.
func (mr *MockStateMockRecorder) SetCurrentSupply(subnetID, cs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCurrentSupply", reflect.TypeOf((*MockState)(nil).SetCurrentSupply), subnetID, cs)
}

// SetDelegateeReward mocks base method.
func (m *MockState) SetDelegateeReward(subnetID ids.ID, nodeID ids.NodeID, amount uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDelegateeReward", subnetID, nodeID, amount)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDelegateeReward indicates an expected call of SetDelegateeReward.
func (mr *MockStateMockRecorder) SetDelegateeReward(subnetID, nodeID, amount any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDelegateeReward", reflect.TypeOf((*MockState)(nil).SetDelegateeReward), subnetID, nodeID, amount)
}

// SetFeeState mocks base method.
func (m *MockState) SetFeeState(f gas.State) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetFeeState", f)
}

// SetFeeState indicates an expected call of SetFeeState.
func (mr *MockStateMockRecorder) SetFeeState(f any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFeeState", reflect.TypeOf((*MockState)(nil).SetFeeState), f)
}

// SetHeight mocks base method.
func (m *MockState) SetHeight(height uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetHeight", height)
}

// SetHeight indicates an expected call of SetHeight.
func (mr *MockStateMockRecorder) SetHeight(height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeight", reflect.TypeOf((*MockState)(nil).SetHeight), height)
}

// SetLastAccepted mocks base method.
func (m *MockState) SetLastAccepted(blkID ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetLastAccepted", blkID)
}

// SetLastAccepted indicates an expected call of SetLastAccepted.
func (mr *MockStateMockRecorder) SetLastAccepted(blkID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLastAccepted", reflect.TypeOf((*MockState)(nil).SetLastAccepted), blkID)
}

// SetSoVExcess mocks base method.
func (m *MockState) SetSoVExcess(e gas.Gas) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetSoVExcess", e)
}

// SetSoVExcess indicates an expected call of SetSoVExcess.
func (mr *MockStateMockRecorder) SetSoVExcess(e any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSoVExcess", reflect.TypeOf((*MockState)(nil).SetSoVExcess), e)
}

// SetSubnetConversion mocks base method.
func (m *MockState) SetSubnetConversion(subnetID ids.ID, c SubnetConversion) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetSubnetConversion", subnetID, c)
}

// SetSubnetConversion indicates an expected call of SetSubnetConversion.
func (mr *MockStateMockRecorder) SetSubnetConversion(subnetID, c any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSubnetConversion", reflect.TypeOf((*MockState)(nil).SetSubnetConversion), subnetID, c)
}

// SetSubnetOwner mocks base method.
func (m *MockState) SetSubnetOwner(subnetID ids.ID, owner fx.Owner) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetSubnetOwner", subnetID, owner)
}

// SetSubnetOwner indicates an expected call of SetSubnetOwner.
func (mr *MockStateMockRecorder) SetSubnetOwner(subnetID, owner any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSubnetOwner", reflect.TypeOf((*MockState)(nil).SetSubnetOwner), subnetID, owner)
}

// SetTimestamp mocks base method.
func (m *MockState) SetTimestamp(tm time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTimestamp", tm)
}

// SetTimestamp indicates an expected call of SetTimestamp.
func (mr *MockStateMockRecorder) SetTimestamp(tm any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTimestamp", reflect.TypeOf((*MockState)(nil).SetTimestamp), tm)
}

// SetUptime mocks base method.
func (m *MockState) SetUptime(nodeID ids.NodeID, upDuration time.Duration, lastUpdated time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetUptime", nodeID, upDuration, lastUpdated)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetUptime indicates an expected call of SetUptime.
func (mr *MockStateMockRecorder) SetUptime(nodeID, upDuration, lastUpdated any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUptime", reflect.TypeOf((*MockState)(nil).SetUptime), nodeID, upDuration, lastUpdated)
}

// UTXOIDs mocks base method.
func (m *MockState) UTXOIDs(addr []byte, previous ids.ID, limit int) ([]ids.ID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UTXOIDs", addr, previous, limit)
	ret0, _ := ret[0].([]ids.ID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UTXOIDs indicates an expected call of UTXOIDs.
func (mr *MockStateMockRecorder) UTXOIDs(addr, previous, limit any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UTXOIDs", reflect.TypeOf((*MockState)(nil).UTXOIDs), addr, previous, limit)
}

// WeightOfSubnetOnlyValidators mocks base method.
func (m *MockState) WeightOfSubnetOnlyValidators(subnetID ids.ID) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WeightOfSubnetOnlyValidators", subnetID)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WeightOfSubnetOnlyValidators indicates an expected call of WeightOfSubnetOnlyValidators.
func (mr *MockStateMockRecorder) WeightOfSubnetOnlyValidators(subnetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WeightOfSubnetOnlyValidators", reflect.TypeOf((*MockState)(nil).WeightOfSubnetOnlyValidators), subnetID)
}
