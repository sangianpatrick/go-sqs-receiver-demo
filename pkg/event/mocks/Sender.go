// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Sender is an autogenerated mock type for the Sender type
type Sender struct {
	mock.Mock
}

// Send provides a mock function with given fields: topic, key, headers, body
func (_m *Sender) Send(topic string, key string, headers map[string]string, body []byte) {
	_m.Called(topic, key, headers, body)
}
