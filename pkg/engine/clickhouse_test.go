package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func setRequiredEnv(t *testing.T, env map[string]string) {
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestClickhouseEngine_Name(t *testing.T) {
	e := &ClickhouseEngine{}
	assert.Equal(t, "clickhouse", e.Name())
}

func TestClickhouseEngine_Restore_DryRun(t *testing.T) {
	e := &ClickhouseEngine{}
	env := map[string]string{
		"CLICKHOUSE_USER":                       "user",
		"CLICKHOUSE_PASSWORD":                   "pass",
		"CLICKHOUSE_AWS_S3_ENDPOINT_URL_BACKUP": "http://s3.example.com",
		"AWS_ACCESS_KEY_ID":                     "AKIA...",
		"AWS_SECRET_ACCESS_KEY":                 "secret",
	}
	setRequiredEnv(t, env)

	err := e.Restore(&genericclioptions.ConfigFlags{}, "backup1", "mydb", RestoreOptions{
		ServiceName: "clickhouse-service",
		Namespace:   "default",
		DryRun:      true,
	})
	assert.NoError(t, err)
}
