// Automatically generated by MockGen. DO NOT EDIT!
// Source: logging.go

package logging

import (
	gomock "github.com/sadbox/sadbot/Godeps/_workspace/src/code.google.com/p/gomock/gomock"
)

// Mock of Logger interface
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *_MockLoggerRecorder
}

// Recorder for MockLogger (not exported)
type _MockLoggerRecorder struct {
	mock *MockLogger
}

func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &_MockLoggerRecorder{mock}
	return mock
}

func (_m *MockLogger) EXPECT() *_MockLoggerRecorder {
	return _m.recorder
}

func (_m *MockLogger) Log(_param0 LogLevel, _param1 string, _param2 ...interface{}) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	_m.ctrl.Call(_m, "Log", _s...)
}

func (_mr *_MockLoggerRecorder) Log(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Log", _s...)
}

func (_m *MockLogger) Debug(_param0 string, _param1 ...interface{}) {
	_s := []interface{}{_param0}
	for _, _x := range _param1 {
		_s = append(_s, _x)
	}
	_m.ctrl.Call(_m, "Debug", _s...)
}

func (_mr *_MockLoggerRecorder) Debug(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Debug", _s...)
}

func (_m *MockLogger) Info(_param0 string, _param1 ...interface{}) {
	_s := []interface{}{_param0}
	for _, _x := range _param1 {
		_s = append(_s, _x)
	}
	_m.ctrl.Call(_m, "Info", _s...)
}

func (_mr *_MockLoggerRecorder) Info(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Info", _s...)
}

func (_m *MockLogger) Warn(_param0 string, _param1 ...interface{}) {
	_s := []interface{}{_param0}
	for _, _x := range _param1 {
		_s = append(_s, _x)
	}
	_m.ctrl.Call(_m, "Warn", _s...)
}

func (_mr *_MockLoggerRecorder) Warn(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Warn", _s...)
}

func (_m *MockLogger) Error(_param0 string, _param1 ...interface{}) {
	_s := []interface{}{_param0}
	for _, _x := range _param1 {
		_s = append(_s, _x)
	}
	_m.ctrl.Call(_m, "Error", _s...)
}

func (_mr *_MockLoggerRecorder) Error(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Error", _s...)
}

func (_m *MockLogger) Fatal(_param0 string, _param1 ...interface{}) {
	_s := []interface{}{_param0}
	for _, _x := range _param1 {
		_s = append(_s, _x)
	}
	_m.ctrl.Call(_m, "Fatal", _s...)
}

func (_mr *_MockLoggerRecorder) Fatal(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Fatal", _s...)
}

func (_m *MockLogger) SetLogLevel(_param0 LogLevel) {
	_m.ctrl.Call(_m, "SetLogLevel", _param0)
}

func (_mr *_MockLoggerRecorder) SetLogLevel(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetLogLevel", arg0)
}

func (_m *MockLogger) SetOnly(_param0 bool) {
	_m.ctrl.Call(_m, "SetOnly", _param0)
}

func (_mr *_MockLoggerRecorder) SetOnly(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetOnly", arg0)
}
