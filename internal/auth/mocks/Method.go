// Code generated by mockery v2.35.4. DO NOT EDIT.

package mocks

import (
	auth "github.com/jannfis/argocd-agent/internal/auth"
	mock "github.com/stretchr/testify/mock"
)

// Method is an autogenerated mock type for the Method type
type Method struct {
	mock.Mock
}

type Method_Expecter struct {
	mock *mock.Mock
}

func (_m *Method) EXPECT() *Method_Expecter {
	return &Method_Expecter{mock: &_m.Mock}
}

// Authenticate provides a mock function with given fields: credentials
func (_m *Method) Authenticate(credentials auth.Credentials) (string, error) {
	ret := _m.Called(credentials)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(auth.Credentials) (string, error)); ok {
		return rf(credentials)
	}
	if rf, ok := ret.Get(0).(func(auth.Credentials) string); ok {
		r0 = rf(credentials)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(auth.Credentials) error); ok {
		r1 = rf(credentials)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Method_Authenticate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Authenticate'
type Method_Authenticate_Call struct {
	*mock.Call
}

// Authenticate is a helper method to define mock.On call
//   - credentials auth.Credentials
func (_e *Method_Expecter) Authenticate(credentials interface{}) *Method_Authenticate_Call {
	return &Method_Authenticate_Call{Call: _e.mock.On("Authenticate", credentials)}
}

func (_c *Method_Authenticate_Call) Run(run func(credentials auth.Credentials)) *Method_Authenticate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(auth.Credentials))
	})
	return _c
}

func (_c *Method_Authenticate_Call) Return(_a0 string, _a1 error) *Method_Authenticate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Method_Authenticate_Call) RunAndReturn(run func(auth.Credentials) (string, error)) *Method_Authenticate_Call {
	_c.Call.Return(run)
	return _c
}

// NewMethod creates a new instance of Method. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMethod(t interface {
	mock.TestingT
	Cleanup(func())
}) *Method {
	mock := &Method{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
