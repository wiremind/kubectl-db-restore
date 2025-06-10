package engine

import (
	"github.com/wiremind/kubectl-restore/pkg/logger"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type PostgresEngine struct{}

func (p *PostgresEngine) Name() string {
	return "postgres"
}

func (p *PostgresEngine) Restore(configFlags *genericclioptions.ConfigFlags, backupName string, databaseName string, opts RestoreOptions) error {
	logger.Global.Info("PostgresEngine.Restore() not implemented yet")
	return nil
}

func init() {
	RegisterEngine(&PostgresEngine{})
}
