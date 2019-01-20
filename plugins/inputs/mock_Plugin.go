package inputs

import (
	"github.com/lavaorg/telex"

	"github.com/stretchr/testify/mock"
)

// MockPlugin struct should be named the same as the Plugin
type MockPlugin struct {
	mock.Mock
}

// Gather defines what data the plugin will gather.
func (m *MockPlugin) Gather(_a0 telex.Accumulator) error {
	ret := m.Called(_a0)

	r0 := ret.Error(0)

	return r0
}
