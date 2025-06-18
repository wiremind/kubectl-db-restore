package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wiremind/kubectl-restore/pkg/engine"
	"github.com/wiremind/kubectl-restore/pkg/k8screds"
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

// --- flag reset helper ---
func resetFlagsAndVars() {
	engineName = ""
	backupName = ""
	databaseName = ""
	namespace = ""
	serviceName = ""
	dryRun = false

	databaseCmd.ResetFlags()

	databaseCmd.Flags().StringVar(&engineName, "engine", "", "Database engine (clickhouse, postgres, ...)")
	databaseCmd.Flags().StringVar(&backupName, "backup-name", "", "Backup name")
	databaseCmd.Flags().StringVar(&databaseName, "database", "", "Database name")
	databaseCmd.Flags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	databaseCmd.Flags().StringVar(&serviceName, "service-name", "", "Kubernetes service name for DB")
	databaseCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run")
}

// --- tests ---
func TestDatabaseCommand_Success(t *testing.T) {
	resetFlagsAndVars()
	mock := &mockEngine{}
	engine.RegisterEngine(mock)

	databaseCmd.SetArgs([]string{
		"--engine=mock",
		"--backup-name=test-backup",
		"--database=test-db",
		"--namespace=test-ns",
		"--service-name=test-svc",
	})

	err := databaseCmd.Execute()
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

func TestDatabaseCommand_RestoreFails(t *testing.T) {
	resetFlagsAndVars()
	mock := &mockEngine{returnErr: errors.New("restore failed")}
	engine.RegisterEngine(mock)

	exitCalled := false
	oldExit := osExit
	osExit = func(code int) {
		exitCalled = true
	}
	defer func() { osExit = oldExit }()

	databaseCmd.SetArgs([]string{
		"--engine=mock",
		"--backup-name=test-backup",
		"--database=test-db",
		"--service-name=test-svc",
	})

	err := databaseCmd.Execute()
	assert.Nil(t, err)
	assert.True(t, exitCalled)
	assert.True(t, mock.restoreCalled)
}
