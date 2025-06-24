package cli

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wiremind/kubectl-db-restore/pkg/engine"
	"github.com/wiremind/kubectl-db-restore/pkg/k8screds"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// --- mock engine ---
type mockEngine struct {
	restoreCalled bool
	returnErr     error
	lastArgs      struct {
		backup   string
		database string
		opts     engine.RestoreOptions
	}
}

func (m *mockEngine) Name() string {
	return "mock"
}

func (m *mockEngine) Restore(
	_ *genericclioptions.ConfigFlags,
	backup, database string,
	opts engine.RestoreOptions,
) error {
	m.restoreCalled = true
	m.lastArgs.backup = backup
	m.lastArgs.database = database
	m.lastArgs.opts = opts
	return m.returnErr
}

// --- test helpers ---
func resetVars() {
	engineName = ""
	backupName = ""
	databaseName = ""
	namespace = ""
	serviceName = ""
	dryRun = false
	secretRefs = nil
	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
}

// --- tests ---
func TestRunDatabaseRestore_Success(t *testing.T) {
	resetVars()
	mock := &mockEngine{}
	engine.RegisterEngine(mock)

	engineName = "mock"
	backupName = "test-backup"
	databaseName = "test-db"
	namespace = "test-ns"
	serviceName = "test-svc"
	dryRun = false
	secretRefs = []string{}

	err := runDatabaseRestore()

	assert.NoError(t, err)
	assert.True(t, mock.restoreCalled)
	assert.Equal(t, "test-backup", mock.lastArgs.backup)
	assert.Equal(t, "test-db", mock.lastArgs.database)
	assert.Equal(t, engine.RestoreOptions{
		Namespace:     "test-ns",
		ServiceName:   "test-svc",
		DryRun:        false,
		SecretKeyRefs: []k8screds.SecretKeyRef{},
	}, mock.lastArgs.opts)
}

func TestRunDatabaseRestore_RestoreFails(t *testing.T) {
	resetVars()
	mock := &mockEngine{returnErr: errors.New("restore failed")}
	engine.RegisterEngine(mock)

	engineName = "mock"
	backupName = "test-backup"
	databaseName = "test-db"
	serviceName = "test-svc"

	// Capture osExit
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
	}
	defer func() { osExit = os.Exit }()

	err := runDatabaseRestore()
	assert.NoError(t, err)

	assert.True(t, exitCalled)
	assert.True(t, mock.restoreCalled)
}

func TestRunDatabaseRestore_InvalidSecretRef(t *testing.T) {
	resetVars()
	mock := &mockEngine{}
	engine.RegisterEngine(mock)

	engineName = "mock"
	backupName = "test-backup"
	databaseName = "test-db"
	serviceName = "test-svc"
	secretRefs = []string{"INVALID_FORMAT"}

	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
	}
	defer func() { osExit = os.Exit }()

	err := runDatabaseRestore()
	assert.NoError(t, err)

	assert.True(t, exitCalled)
}
