// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"

// KVDB is an autogenerated mock type for the KVDB type
type KVDB struct {
	mock.Mock
}

// BatchGet provides a mock function with given fields: keys
func (_m *KVDB) BatchGet(keys [][]byte) ([][]byte, error) {
	ret := _m.Called(keys)

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func([][]byte) [][]byte); ok {
		r0 = rf(keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([][]byte) error); ok {
		r1 = rf(keys)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Begin provides a mock function with given fields:
func (_m *KVDB) Begin() {
	_m.Called()
}

// Commit provides a mock function with given fields:
func (_m *KVDB) Commit() {
	_m.Called()
}

// Get provides a mock function with given fields: key
func (_m *KVDB) Get(key []byte) ([]byte, error) {
	ret := _m.Called(key)

	var r0 []byte
	if rf, ok := ret.Get(0).(func([]byte) []byte); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: prefix, key, count, direction
func (_m *KVDB) List(prefix []byte, key []byte, count int32, direction int32) ([][]byte, error) {
	ret := _m.Called(prefix, key, count, direction)

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func([]byte, []byte, int32, int32) [][]byte); ok {
		r0 = rf(prefix, key, count, direction)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte, []byte, int32, int32) error); ok {
		r1 = rf(prefix, key, count, direction)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PrefixCount provides a mock function with given fields: prefix
func (_m *KVDB) PrefixCount(prefix []byte) int64 {
	ret := _m.Called(prefix)

	var r0 int64
	if rf, ok := ret.Get(0).(func([]byte) int64); ok {
		r0 = rf(prefix)
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// Rollback provides a mock function with given fields:
func (_m *KVDB) Rollback() {
	_m.Called()
}

// Set provides a mock function with given fields: key, value
func (_m *KVDB) Set(key []byte, value []byte) error {
	ret := _m.Called(key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, []byte) error); ok {
		r0 = rf(key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
