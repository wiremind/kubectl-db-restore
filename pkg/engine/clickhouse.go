package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/wiremind/kubectl-db-restore/pkg/job"
	"github.com/wiremind/kubectl-db-restore/pkg/k8screds"
	"github.com/wiremind/kubectl-db-restore/pkg/logger"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ClickhouseEngine struct{}

func (c *ClickhouseEngine) Name() string {
	return "clickhouse"
}

func (c *ClickhouseEngine) Restore(configFlags *genericclioptions.ConfigFlags, backupName, databaseName string, opts RestoreOptions) error {
	requiredVars := []string{
		"CLICKHOUSE_USER",
		"CLICKHOUSE_PASSWORD",
		"CLICKHOUSE_AWS_S3_ENDPOINT_URL_BACKUP",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
	}

	// Load secrets
	resolvedVars, err := k8screds.LoadSecretsVars(configFlags, opts.Namespace, opts.SecretKeyRefs, requiredVars)
	if err != nil {
		return fmt.Errorf("failed to load secret vars: %w", err)
	}

	// Convert to environment variables
	var envSources []job.EnvVarSource
	for name, lv := range resolvedVars {
		env := job.EnvVarSource{Name: name}
		if lv.FromSecretRef != nil {
			env.SecretRef = lv.FromSecretRef
		} else if lv.FromEnv != nil {
			env.Value = lv.FromEnv
		}
		envSources = append(envSources, env)
	}

	if opts.DryRun {
		logger.Global.Info("🔍 [Dry Run] Initiating validation for restore process...")
		logger.Global.Info("[Dry Run] Target database: '%s'", databaseName)
		logger.Global.Info("[Dry Run] Backup source: '%s'", backupName)
		logger.Global.Info("[Dry Run] Service name (ClickHouse host): '%s'", opts.ServiceName)
		logger.Global.Info("[Dry Run] Namespace: '%s'", opts.Namespace)
		logger.Global.Info("[Dry Run] Validated secret keys: %v", requiredVars)

		for _, env := range envSources {
			switch {
			case env.SecretRef != nil:
				logger.Global.Info("[Dry Run] Would load secret for var: %s", env.Name)
			case env.Value != nil:
				logger.Global.Info("[Dry Run] Would use env var '%s' with direct value (masked)", env.Name)
			default:
				logger.Global.Info("[Dry Run] ⚠️ Missing or unresolved value for env var: %s", env.Name)
			}
		}

		logger.Global.Info("[Dry Run] Would create 3 sequential Kubernetes jobs:")
		logger.Global.Info("  - 🗑️ Job: Drop database '%s' (if it exists)", databaseName)
		logger.Global.Info("  - 🏗️ Job: Create new database '%s'", databaseName)
		logger.Global.Info("  - 📦 Job: Restore database '%s' from S3 path '%s/%s'", databaseName, os.Getenv("CLICKHOUSE_AWS_S3_ENDPOINT_URL_BACKUP"), backupName)

		logger.Global.Info("✅ [Dry Run] Validation completed successfully. No changes were made.")
		return nil
	}

	logger.Global.Info("🚀 Starting ClickHouse restore sequence for database: %s", databaseName)

	timestamp := time.Now().Unix()

	// SQL job phases
	jobs := []struct {
		Name           string
		Script         string
		SuccessMessage string
		FailureHeader  string
	}{
		{
			Name: "clickhouse-drop-db",
			Script: fmt.Sprintf(`clickhouse-client --host %s \
--user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
--query "DROP DATABASE IF EXISTS %s ON CLUSTER default SYNC"`, opts.ServiceName, databaseName),
			SuccessMessage: fmt.Sprintf("🗑️ Successfully dropped database '%s' (if it existed)", databaseName),
			FailureHeader:  "🛑 Failed to drop existing database",
		},
		{
			Name: "clickhouse-create-db",
			Script: fmt.Sprintf(`clickhouse-client --host %s \
--user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
--query "CREATE DATABASE %s ON CLUSTER default"`, opts.ServiceName, databaseName),
			SuccessMessage: fmt.Sprintf("🏗️ Successfully created database '%s'", databaseName),
			FailureHeader:  "❌ Failed to create new database",
		},
		{
			Name: "clickhouse-restore",
			Script: fmt.Sprintf(`clickhouse-client --host %s \
--user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
--query "RESTORE DATABASE %s FROM S3('$CLICKHOUSE_AWS_S3_ENDPOINT_URL_BACKUP/%s', '$AWS_ACCESS_KEY_ID', '$AWS_SECRET_ACCESS_KEY')"`,
				opts.ServiceName, databaseName, backupName),
			SuccessMessage: fmt.Sprintf("✅ Successfully restored database '%s' from backup '%s'", databaseName, backupName),
			FailureHeader:  "💣 ClickHouse restore job failed",
		},
	}

	for i, jobCfg := range jobs {
		jobSpec := job.JobSpec{
			Namespace:         opts.Namespace,
			JobName:           fmt.Sprintf("%s-%d", jobCfg.Name, timestamp+int64(i)),
			Image:             "clickhouse/clickhouse-server:25.5-alpine",
			Command:           []string{"/bin/sh"},
			Args:              []string{"-c", jobCfg.Script},
			EnvVars:           envSources,
			JobSuccessMessage: jobCfg.SuccessMessage,
			JobFailureHeader:  jobCfg.FailureHeader,
		}

		if err := job.CreateJob(configFlags, jobSpec); err != nil {
			return fmt.Errorf("failed to create %s job: %w", jobCfg.Name, err)
		}
	}

	logger.Global.Info("🎉 All jobs for ClickHouse restore sequence completed successfully!")
	return nil
}

func init() {
	RegisterEngine(&ClickhouseEngine{})
}
