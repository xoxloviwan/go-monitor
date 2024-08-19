// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/xoxloviwan/go-monitor/internal/api (interfaces: ReaderWriter)

// Package mock_api is a generated GoMock package.
package mock_api

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	metrictypes "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

// MockReaderWriter is a mock of ReaderWriter interface.
type MockReaderWriter struct {
	ctrl     *gomock.Controller
	recorder *MockReaderWriterMockRecorder
}

// MockReaderWriterMockRecorder is the mock recorder for MockReaderWriter.
type MockReaderWriterMockRecorder struct {
	mock *MockReaderWriter
}

// NewMockReaderWriter creates a new mock instance.
func NewMockReaderWriter(ctrl *gomock.Controller) *MockReaderWriter {
	mock := &MockReaderWriter{ctrl: ctrl}
	mock.recorder = &MockReaderWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReaderWriter) EXPECT() *MockReaderWriterMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockReaderWriter) Add(arg0, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockReaderWriterMockRecorder) Add(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockReaderWriter)(nil).Add), arg0, arg1, arg2)
}

// AddMetrics mocks base method.
func (m *MockReaderWriter) AddMetrics(arg0 context.Context, arg1 *metrictypes.MetricsList) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddMetrics", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddMetrics indicates an expected call of AddMetrics.
func (mr *MockReaderWriterMockRecorder) AddMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMetrics", reflect.TypeOf((*MockReaderWriter)(nil).AddMetrics), arg0, arg1)
}

// Get mocks base method.
func (m *MockReaderWriter) Get(arg0, arg1 string) (string, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockReaderWriterMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockReaderWriter)(nil).Get), arg0, arg1)
}

// GetMetrics mocks base method.
func (m *MockReaderWriter) GetMetrics(arg0 context.Context, arg1 metrictypes.MetricsList) (metrictypes.MetricsList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetrics", arg0, arg1)
	ret0, _ := ret[0].(metrictypes.MetricsList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetrics indicates an expected call of GetMetrics.
func (mr *MockReaderWriterMockRecorder) GetMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetrics", reflect.TypeOf((*MockReaderWriter)(nil).GetMetrics), arg0, arg1)
}

// String mocks base method.
func (m *MockReaderWriter) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockReaderWriterMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockReaderWriter)(nil).String))
}