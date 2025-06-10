package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wiremind/kubectl-restore/pkg/engine"
	"github.com/wiremind/kubectl-restore/pkg/k8screds"
	"github.com/wiremind/kubectl-restore/pkg/logger"
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

var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Restore a database",
	Run: func(cmd *cobra.Command, args []string) {
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
			}
			secretParts := strings.SplitN(parts[1], ":", 2)
			if len(secretParts) != 2 {
				logger.Global.Error(fmt.Errorf("invalid secret/key in --secret-ref: %s", ref))
				osExit(1)
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
	},
}

func init() {
	databaseCmd.Flags().StringVar(&engineName, "engine", "", "Database engine (clickhouse, postgres, ...)")
	databaseCmd.Flags().StringVar(&backupName, "backup-name", "", "Backup name")
	databaseCmd.Flags().StringVar(&databaseName, "database", "", "Database name")
	databaseCmd.Flags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	databaseCmd.Flags().StringVar(&serviceName, "service-name", "", "Kubernetes service name for DB")
	databaseCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run")
	databaseCmd.Flags().StringSliceVar(
		&secretRefs,
		"secret-ref",
		nil,
		"Secret reference in the format VAR=secretName:key (can be repeated)",
	)

	if err := databaseCmd.MarkFlagRequired("engine"); err != nil {
		panic(err)
	}
	if err := databaseCmd.MarkFlagRequired("backup-name"); err != nil {
		panic(err)
	}
	if err := databaseCmd.MarkFlagRequired("database"); err != nil {
		panic(err)
	}
	if err := databaseCmd.MarkFlagRequired("service-name"); err != nil {
		panic(err)
	}
}
