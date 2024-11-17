// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/xoxloviwan/go-monitor/internal/api (interfaces: Router)

// Package api is a generated GoMock package.
package api

import (
	slog "log/slog"
	reflect "reflect"

	gin "github.com/gin-gonic/gin"
	gomock "github.com/golang/mock/gomock"
)

// MockRouter is a mock of Router interface.
type MockRouter struct {
	ctrl     *gomock.Controller
	recorder *MockRouterMockRecorder
}

// MockRouterMockRecorder is the mock recorder for MockRouter.
type MockRouterMockRecorder struct {
	mock *MockRouter
}

// NewMockRouter creates a new mock instance.
func NewMockRouter(ctrl *gomock.Controller) *MockRouter {
	mock := &MockRouter{ctrl: ctrl}
	mock.recorder = &MockRouterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRouter) EXPECT() *MockRouterMockRecorder {
	return m.recorder
}

// Run mocks base method.
func (m *MockRouter) Run(arg0 ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Run", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockRouterMockRecorder) Run(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockRouter)(nil).Run), arg0...)
}

// SetupRouter mocks base method.
func (m *MockRouter) SetupRouter(arg0 gin.HandlerFunc, arg1 ReaderWriter, arg2 slog.Level, arg3 []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetupRouter", arg0, arg1, arg2, arg3)
}

// SetupRouter indicates an expected call of SetupRouter.
func (mr *MockRouterMockRecorder) SetupRouter(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetupRouter", reflect.TypeOf((*MockRouter)(nil).SetupRouter), arg0, arg1, arg2, arg3)
}
