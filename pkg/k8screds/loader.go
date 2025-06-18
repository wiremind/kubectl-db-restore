package k8screds

import (
	"fmt"
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type SecretKeyRef struct {
	EnvVarName string // Logical name: e.g. CLICKHOUSE_USER
	SecretName string // Secret: e.g. ch-secret
	Key        string // Key in secret: e.g. user
}

type LoadedVar struct {
	FromEnv       *string       // if set, from os env
	FromSecretRef *SecretKeyRef // if set, to use in valueFrom
}

// LoadEnvVarsSmart loads credentials from explicit SecretRefs if available,
// otherwise falls back to env vars. Returns an error if any required key is missing.
func LoadSecretsVars(configFlags *genericclioptions.ConfigFlags, namespace string, refs []SecretKeyRef, requiredVars []string) (map[string]LoadedVar, error) {
	result := map[string]LoadedVar{}

	refMap := map[string]SecretKeyRef{}
	for _, ref := range refs {
		refMap[ref.EnvVarName] = ref
	}

	for _, key := range requiredVars {
		if ref, ok := refMap[key]; ok {
			result[key] = LoadedVar{
				FromSecretRef: &ref,
			}
		} else {
			envVal := os.Getenv(key)
			if envVal == "" {
				return nil, fmt.Errorf("missing required variable %q (not in secret nor env)", key)
			}
			result[key] = LoadedVar{
				FromEnv: &envVal,
			}
		}
	}

	return result, nil
}
