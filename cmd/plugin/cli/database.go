package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/wiremind/kubectl-db-restore/pkg/engine"
	"github.com/wiremind/kubectl-db-restore/pkg/k8screds"
	"github.com/wiremind/kubectl-db-restore/pkg/logger"
)

var (
	engineName   string
	backupName   string
	databaseName string
	namespace    string
	serviceName  string
	dryRun       bool
	osExit       = os.Exit
	secretRefs   []string
)

func runDatabaseRestore() error {
	logger.Global.Info("Restoring database '%s' from backup '%s' using engine '%s'", databaseName, backupName, engineName)

	eng, err := engine.GetEngine(engineName)
	if err != nil {
		logger.Global.Error(err)
		osExit(1)
	}
	parsedRefs := []k8screds.SecretKeyRef{}
	for _, ref := range secretRefs {
		parts := strings.SplitN(ref, "=", 2)
		if len(parts) != 2 {
			logger.Global.Error(fmt.Errorf("invalid --secret-ref format: %s", ref))
			osExit(1)
			return nil // add this for testability
		}

		secretParts := strings.SplitN(parts[1], ":", 2)
		if len(secretParts) != 2 {
			logger.Global.Error(fmt.Errorf("invalid secret/key in --secret-ref: %s", ref))
			osExit(1)
			return nil
		}

		parsedRefs = append(parsedRefs, k8screds.SecretKeyRef{
			EnvVarName: parts[0],
			SecretName: secretParts[0],
			Key:        secretParts[1],
		})
	}

	opts := engine.RestoreOptions{
		Namespace:     namespace,
		ServiceName:   serviceName,
		DryRun:        dryRun,
		SecretKeyRefs: parsedRefs,
	}

	err = eng.Restore(KubernetesConfigFlags, backupName, databaseName, opts)
	if err != nil {
		logger.Global.Error(err)
		osExit(1)
	}

	logger.Global.Info("Restore completed successfully")
	return nil
}
