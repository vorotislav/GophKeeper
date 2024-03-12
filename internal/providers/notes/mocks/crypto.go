// Code generated by mockery v2.24.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Crypto is an autogenerated mock type for the crypto type
type Crypto struct {
	mock.Mock
}

type Crypto_Expecter struct {
	mock *mock.Mock
}

func (_m *Crypto) EXPECT() *Crypto_Expecter {
	return &Crypto_Expecter{mock: &_m.Mock}
}

// DecryptString provides a mock function with given fields: value
func (_m *Crypto) DecryptString(value string) (string, error) {
	ret := _m.Called(value)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(value)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(value)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(value)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Crypto_DecryptString_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DecryptString'
type Crypto_DecryptString_Call struct {
	*mock.Call
}

// DecryptString is a helper method to define mock.On call
//   - value string
func (_e *Crypto_Expecter) DecryptString(value interface{}) *Crypto_DecryptString_Call {
	return &Crypto_DecryptString_Call{Call: _e.mock.On("DecryptString", value)}
}

func (_c *Crypto_DecryptString_Call) Run(run func(value string)) *Crypto_DecryptString_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Crypto_DecryptString_Call) Return(_a0 string, _a1 error) *Crypto_DecryptString_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Crypto_DecryptString_Call) RunAndReturn(run func(string) (string, error)) *Crypto_DecryptString_Call {
	_c.Call.Return(run)
	return _c
}

// EncryptString provides a mock function with given fields: value
func (_m *Crypto) EncryptString(value string) (string, error) {
	ret := _m.Called(value)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(value)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(value)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(value)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Crypto_EncryptString_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EncryptString'
type Crypto_EncryptString_Call struct {
	*mock.Call
}

// EncryptString is a helper method to define mock.On call
//   - value string
func (_e *Crypto_Expecter) EncryptString(value interface{}) *Crypto_EncryptString_Call {
	return &Crypto_EncryptString_Call{Call: _e.mock.On("EncryptString", value)}
}

func (_c *Crypto_EncryptString_Call) Run(run func(value string)) *Crypto_EncryptString_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Crypto_EncryptString_Call) Return(_a0 string, _a1 error) *Crypto_EncryptString_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Crypto_EncryptString_Call) RunAndReturn(run func(string) (string, error)) *Crypto_EncryptString_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewCrypto interface {
	mock.TestingT
	Cleanup(func())
}

// NewCrypto creates a new instance of Crypto. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCrypto(t mockConstructorTestingTNewCrypto) *Crypto {
	mock := &Crypto{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}