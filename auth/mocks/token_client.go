// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	v1 "github.com/hantdev/mitras/api/grpc/token/v1"
)

// TokenServiceClient is an autogenerated mock type for the TokenServiceClient type
type TokenServiceClient struct {
	mock.Mock
}

type TokenServiceClient_Expecter struct {
	mock *mock.Mock
}

func (_m *TokenServiceClient) EXPECT() *TokenServiceClient_Expecter {
	return &TokenServiceClient_Expecter{mock: &_m.Mock}
}

// Issue provides a mock function with given fields: ctx, in, opts
func (_m *TokenServiceClient) Issue(ctx context.Context, in *v1.IssueReq, opts ...grpc.CallOption) (*v1.Token, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Issue")
	}

	var r0 *v1.Token
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.IssueReq, ...grpc.CallOption) (*v1.Token, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.IssueReq, ...grpc.CallOption) *v1.Token); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Token)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.IssueReq, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TokenServiceClient_Issue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Issue'
type TokenServiceClient_Issue_Call struct {
	*mock.Call
}

// Issue is a helper method to define mock.On call
//   - ctx context.Context
//   - in *v1.IssueReq
//   - opts ...grpc.CallOption
func (_e *TokenServiceClient_Expecter) Issue(ctx interface{}, in interface{}, opts ...interface{}) *TokenServiceClient_Issue_Call {
	return &TokenServiceClient_Issue_Call{Call: _e.mock.On("Issue",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *TokenServiceClient_Issue_Call) Run(run func(ctx context.Context, in *v1.IssueReq, opts ...grpc.CallOption)) *TokenServiceClient_Issue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*v1.IssueReq), variadicArgs...)
	})
	return _c
}

func (_c *TokenServiceClient_Issue_Call) Return(_a0 *v1.Token, _a1 error) *TokenServiceClient_Issue_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TokenServiceClient_Issue_Call) RunAndReturn(run func(context.Context, *v1.IssueReq, ...grpc.CallOption) (*v1.Token, error)) *TokenServiceClient_Issue_Call {
	_c.Call.Return(run)
	return _c
}

// Refresh provides a mock function with given fields: ctx, in, opts
func (_m *TokenServiceClient) Refresh(ctx context.Context, in *v1.RefreshReq, opts ...grpc.CallOption) (*v1.Token, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Refresh")
	}

	var r0 *v1.Token
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.RefreshReq, ...grpc.CallOption) (*v1.Token, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.RefreshReq, ...grpc.CallOption) *v1.Token); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Token)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.RefreshReq, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TokenServiceClient_Refresh_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Refresh'
type TokenServiceClient_Refresh_Call struct {
	*mock.Call
}

// Refresh is a helper method to define mock.On call
//   - ctx context.Context
//   - in *v1.RefreshReq
//   - opts ...grpc.CallOption
func (_e *TokenServiceClient_Expecter) Refresh(ctx interface{}, in interface{}, opts ...interface{}) *TokenServiceClient_Refresh_Call {
	return &TokenServiceClient_Refresh_Call{Call: _e.mock.On("Refresh",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *TokenServiceClient_Refresh_Call) Run(run func(ctx context.Context, in *v1.RefreshReq, opts ...grpc.CallOption)) *TokenServiceClient_Refresh_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*v1.RefreshReq), variadicArgs...)
	})
	return _c
}

func (_c *TokenServiceClient_Refresh_Call) Return(_a0 *v1.Token, _a1 error) *TokenServiceClient_Refresh_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TokenServiceClient_Refresh_Call) RunAndReturn(run func(context.Context, *v1.RefreshReq, ...grpc.CallOption) (*v1.Token, error)) *TokenServiceClient_Refresh_Call {
	_c.Call.Return(run)
	return _c
}

// NewTokenServiceClient creates a new instance of TokenServiceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTokenServiceClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *TokenServiceClient {
	mock := &TokenServiceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
