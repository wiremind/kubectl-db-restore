package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// DummyEngine is a test double for the Engine interface
type DummyEngine struct {
	calledRestore bool
}

func (d *DummyEngine) Name() string {
	return "dummy"
}

func (d *DummyEngine) Restore(_ *genericclioptions.ConfigFlags, _ string, _ string, _ RestoreOptions) error {
	d.calledRestore = true
	return nil
}

func TestRegisterAndGetEngine(t *testing.T) {
	engine := &DummyEngine{}

	// Register the engine
	RegisterEngine(engine)

	// Retrieve it
	retrieved, err := GetEngine("dummy")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	// Ensure it matches the original
	assert.Equal(t, "dummy", retrieved.Name())

	// Check that Restore works
	err = retrieved.Restore(nil, "backup1", "db1", RestoreOptions{})
	assert.NoError(t, err)

	// Cast back to DummyEngine to check internal state
	dummy, ok := retrieved.(*DummyEngine)
	require.True(t, ok)
	assert.True(t, dummy.calledRestore)
}

func TestGetEngine_Unknown(t *testing.T) {
	e, err := GetEngine("nonexistent")
	require.Error(t, err)
	assert.Nil(t, e)
	assert.Contains(t, err.Error(), "unknown engine")
}
