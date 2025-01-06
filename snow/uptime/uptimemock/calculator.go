// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ava-labs/avalanchego/snow/uptime (interfaces: Calculator)
//
// Generated by this command:
//
//	mockgen -package=uptimemock -destination=uptimemock/calculator.go -mock_names=Calculator=Calculator . Calculator
//

// Package uptimemock is a generated GoMock package.
package uptimemock

import (
	reflect "reflect"
	time "time"

	ids "github.com/ava-labs/avalanchego/ids"
	gomock "go.uber.org/mock/gomock"
)

// Calculator is a mock of Calculator interface.
type Calculator struct {
	ctrl     *gomock.Controller
	recorder *CalculatorMockRecorder
	isgomock struct{}
}

// CalculatorMockRecorder is the mock recorder for Calculator.
type CalculatorMockRecorder struct {
	mock *Calculator
}

// NewCalculator creates a new mock instance.
func NewCalculator(ctrl *gomock.Controller) *Calculator {
	mock := &Calculator{ctrl: ctrl}
	mock.recorder = &CalculatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Calculator) EXPECT() *CalculatorMockRecorder {
	return m.recorder
}

// CalculateUptime mocks base method.
func (m *Calculator) CalculateUptime(nodeID ids.NodeID) (time.Duration, time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CalculateUptime", nodeID)
	ret0, _ := ret[0].(time.Duration)
	ret1, _ := ret[1].(time.Time)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CalculateUptime indicates an expected call of CalculateUptime.
func (mr *CalculatorMockRecorder) CalculateUptime(nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CalculateUptime", reflect.TypeOf((*Calculator)(nil).CalculateUptime), nodeID)
}

// CalculateUptimePercent mocks base method.
func (m *Calculator) CalculateUptimePercent(nodeID ids.NodeID) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CalculateUptimePercent", nodeID)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CalculateUptimePercent indicates an expected call of CalculateUptimePercent.
func (mr *CalculatorMockRecorder) CalculateUptimePercent(nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CalculateUptimePercent", reflect.TypeOf((*Calculator)(nil).CalculateUptimePercent), nodeID)
}

// CalculateUptimePercentFrom mocks base method.
func (m *Calculator) CalculateUptimePercentFrom(nodeID ids.NodeID, startTime time.Time) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CalculateUptimePercentFrom", nodeID, startTime)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CalculateUptimePercentFrom indicates an expected call of CalculateUptimePercentFrom.
func (mr *CalculatorMockRecorder) CalculateUptimePercentFrom(nodeID, startTime any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CalculateUptimePercentFrom", reflect.TypeOf((*Calculator)(nil).CalculateUptimePercentFrom), nodeID, startTime)
}
