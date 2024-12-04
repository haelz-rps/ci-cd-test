// Code generated by mockery v2.26.1. DO NOT EDIT.

package feed

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockPostRepository is an autogenerated mock type for the PostRepository type
type MockPostRepository struct {
	mock.Mock
}

type MockPostRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPostRepository) EXPECT() *MockPostRepository_Expecter {
	return &MockPostRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, post
func (_m *MockPostRepository) Create(ctx context.Context, post *Post) error {
	ret := _m.Called(ctx, post)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *Post) error); ok {
		r0 = rf(ctx, post)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockPostRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockPostRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - post *Post
func (_e *MockPostRepository_Expecter) Create(ctx interface{}, post interface{}) *MockPostRepository_Create_Call {
	return &MockPostRepository_Create_Call{Call: _e.mock.On("Create", ctx, post)}
}

func (_c *MockPostRepository_Create_Call) Run(run func(ctx context.Context, post *Post)) *MockPostRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*Post))
	})
	return _c
}

func (_c *MockPostRepository_Create_Call) Return(_a0 error) *MockPostRepository_Create_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockPostRepository_Create_Call) RunAndReturn(run func(context.Context, *Post) error) *MockPostRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, feedID
func (_m *MockPostRepository) List(ctx context.Context, feedID int) ([]Post, error) {
	ret := _m.Called(ctx, feedID)

	var r0 []Post
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int) ([]Post, error)); ok {
		return rf(ctx, feedID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int) []Post); ok {
		r0 = rf(ctx, feedID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Post)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, feedID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockPostRepository_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type MockPostRepository_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - feedID int
func (_e *MockPostRepository_Expecter) List(ctx interface{}, feedID interface{}) *MockPostRepository_List_Call {
	return &MockPostRepository_List_Call{Call: _e.mock.On("List", ctx, feedID)}
}

func (_c *MockPostRepository_List_Call) Run(run func(ctx context.Context, feedID int)) *MockPostRepository_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int))
	})
	return _c
}

func (_c *MockPostRepository_List_Call) Return(_a0 []Post, _a1 error) *MockPostRepository_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPostRepository_List_Call) RunAndReturn(run func(context.Context, int) ([]Post, error)) *MockPostRepository_List_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewMockPostRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockPostRepository creates a new instance of MockPostRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockPostRepository(t mockConstructorTestingTNewMockPostRepository) *MockPostRepository {
	mock := &MockPostRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}