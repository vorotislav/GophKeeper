// Code generated by mockery v2.24.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	models "GophKeeper/internal/models"
)

// Storage is an autogenerated mock type for the storage type
type Storage struct {
	mock.Mock
}

type Storage_Expecter struct {
	mock *mock.Mock
}

func (_m *Storage) EXPECT() *Storage_Expecter {
	return &Storage_Expecter{mock: &_m.Mock}
}

// MediaCreate provides a mock function with given fields: ctx, m, userID
func (_m *Storage) MediaCreate(ctx context.Context, m models.Media, userID int) error {
	ret := _m.Called(ctx, m, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, models.Media, int) error); ok {
		r0 = rf(ctx, m, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_MediaCreate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MediaCreate'
type Storage_MediaCreate_Call struct {
	*mock.Call
}

// MediaCreate is a helper method to define mock.On call
//   - ctx context.Context
//   - m models.Media
//   - userID int
func (_e *Storage_Expecter) MediaCreate(ctx interface{}, m interface{}, userID interface{}) *Storage_MediaCreate_Call {
	return &Storage_MediaCreate_Call{Call: _e.mock.On("MediaCreate", ctx, m, userID)}
}

func (_c *Storage_MediaCreate_Call) Run(run func(ctx context.Context, m models.Media, userID int)) *Storage_MediaCreate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(models.Media), args[2].(int))
	})
	return _c
}

func (_c *Storage_MediaCreate_Call) Return(_a0 error) *Storage_MediaCreate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_MediaCreate_Call) RunAndReturn(run func(context.Context, models.Media, int) error) *Storage_MediaCreate_Call {
	_c.Call.Return(run)
	return _c
}

// MediaDelete provides a mock function with given fields: ctx, id, userID
func (_m *Storage) MediaDelete(ctx context.Context, id int, userID int) error {
	ret := _m.Called(ctx, id, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int, int) error); ok {
		r0 = rf(ctx, id, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_MediaDelete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MediaDelete'
type Storage_MediaDelete_Call struct {
	*mock.Call
}

// MediaDelete is a helper method to define mock.On call
//   - ctx context.Context
//   - id int
//   - userID int
func (_e *Storage_Expecter) MediaDelete(ctx interface{}, id interface{}, userID interface{}) *Storage_MediaDelete_Call {
	return &Storage_MediaDelete_Call{Call: _e.mock.On("MediaDelete", ctx, id, userID)}
}

func (_c *Storage_MediaDelete_Call) Run(run func(ctx context.Context, id int, userID int)) *Storage_MediaDelete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int), args[2].(int))
	})
	return _c
}

func (_c *Storage_MediaDelete_Call) Return(_a0 error) *Storage_MediaDelete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_MediaDelete_Call) RunAndReturn(run func(context.Context, int, int) error) *Storage_MediaDelete_Call {
	_c.Call.Return(run)
	return _c
}

// MediaUpdate provides a mock function with given fields: ctx, m, userID
func (_m *Storage) MediaUpdate(ctx context.Context, m models.Media, userID int) error {
	ret := _m.Called(ctx, m, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, models.Media, int) error); ok {
		r0 = rf(ctx, m, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_MediaUpdate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MediaUpdate'
type Storage_MediaUpdate_Call struct {
	*mock.Call
}

// MediaUpdate is a helper method to define mock.On call
//   - ctx context.Context
//   - m models.Media
//   - userID int
func (_e *Storage_Expecter) MediaUpdate(ctx interface{}, m interface{}, userID interface{}) *Storage_MediaUpdate_Call {
	return &Storage_MediaUpdate_Call{Call: _e.mock.On("MediaUpdate", ctx, m, userID)}
}

func (_c *Storage_MediaUpdate_Call) Run(run func(ctx context.Context, m models.Media, userID int)) *Storage_MediaUpdate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(models.Media), args[2].(int))
	})
	return _c
}

func (_c *Storage_MediaUpdate_Call) Return(_a0 error) *Storage_MediaUpdate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_MediaUpdate_Call) RunAndReturn(run func(context.Context, models.Media, int) error) *Storage_MediaUpdate_Call {
	_c.Call.Return(run)
	return _c
}

// Medias provides a mock function with given fields: ctx, userID
func (_m *Storage) Medias(ctx context.Context, userID int) ([]models.Media, error) {
	ret := _m.Called(ctx, userID)

	var r0 []models.Media
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int) ([]models.Media, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int) []models.Media); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Media)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_Medias_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Medias'
type Storage_Medias_Call struct {
	*mock.Call
}

// Medias is a helper method to define mock.On call
//   - ctx context.Context
//   - userID int
func (_e *Storage_Expecter) Medias(ctx interface{}, userID interface{}) *Storage_Medias_Call {
	return &Storage_Medias_Call{Call: _e.mock.On("Medias", ctx, userID)}
}

func (_c *Storage_Medias_Call) Run(run func(ctx context.Context, userID int)) *Storage_Medias_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int))
	})
	return _c
}

func (_c *Storage_Medias_Call) Return(_a0 []models.Media, _a1 error) *Storage_Medias_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_Medias_Call) RunAndReturn(run func(context.Context, int) ([]models.Media, error)) *Storage_Medias_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewStorage interface {
	mock.TestingT
	Cleanup(func())
}

// NewStorage creates a new instance of Storage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStorage(t mockConstructorTestingTNewStorage) *Storage {
	mock := &Storage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
