package agent

import (
	"testing"

	"github.com/lavaorg/telex/internal/config"

	// needing to load the plugins
	_ "github.com/lavaorg/telex/plugins/inputs/all"
	// needing to load the outputs
	_ "github.com/lavaorg/telex/plugins/outputs/all"

	"github.com/stretchr/testify/assert"
)

func TestAgent_OmitHostname(t *testing.T) {
	c := config.NewConfig()
	c.Agent.OmitHostname = true
	_, err := NewAgent(c)
	assert.NoError(t, err)
	assert.NotContains(t, c.Tags, "host")
}

func TestAgent_LoadPlugin(t *testing.T) {
	c := config.NewConfig()
	c.InputFilters = []string{"cpu"}
	err := c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ := NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"foo"}
	err = c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"cpu", "foo"}
	err = c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"diskio", "exec"}
	err = c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"diskio", "foo", "cpu", "bar"}
	err = c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Inputs))
}

func TestAgent_LoadOutput(t *testing.T) {
	c := config.NewConfig()
	c.OutputFilters = []string{"file"}
	err := c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ := NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"file"}
	err = c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{}
	err = c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"foo"}
	err = c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"socket_writer", "foo"}
	err = c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"file", "foo", "socket_writer", "bar"}
	err = c.LoadConfig("../internal/config/testdata/telex-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Outputs))
}
