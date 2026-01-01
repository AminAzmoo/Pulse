package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHostInfo(t *testing.T) {
	info := GetHostInfo()
	assert.NotNil(t, info)
	assert.NotEmpty(t, info.Os)
	assert.NotEmpty(t, info.Arch)
	assert.NotEmpty(t, info.Hostname)
	assert.NotEmpty(t, info.AgentVersion)
	
	// Platform might be empty on some test environments, but usually populated
	t.Logf("Host Info: %+v", info)
}
