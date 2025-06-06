// Code generated by MockGen. DO NOT EDIT.
// Source: ./ (interfaces: Recaptcha)
//
// Generated by this command:
//
//	mockgen -destination mocks/recaptcha_mock.go -package mockrecaptcha ./ Recaptcha
//

// Package mockrecaptcha is a generated GoMock package.
package mockrecaptcha

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockRecaptcha is a mock of Recaptcha interface.
type MockRecaptcha struct {
	ctrl     *gomock.Controller
	recorder *MockRecaptchaMockRecorder
	isgomock struct{}
}

// MockRecaptchaMockRecorder is the mock recorder for MockRecaptcha.
type MockRecaptchaMockRecorder struct {
	mock *MockRecaptcha
}

// NewMockRecaptcha creates a new mock instance.
func NewMockRecaptcha(ctrl *gomock.Controller) *MockRecaptcha {
	mock := &MockRecaptcha{ctrl: ctrl}
	mock.recorder = &MockRecaptchaMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRecaptcha) EXPECT() *MockRecaptchaMockRecorder {
	return m.recorder
}

// SiteVerify mocks base method.
func (m *MockRecaptcha) SiteVerify(ctx context.Context, secret, token string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SiteVerify", ctx, secret, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// SiteVerify indicates an expected call of SiteVerify.
func (mr *MockRecaptchaMockRecorder) SiteVerify(ctx, secret, token any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SiteVerify", reflect.TypeOf((*MockRecaptcha)(nil).SiteVerify), ctx, secret, token)
}
