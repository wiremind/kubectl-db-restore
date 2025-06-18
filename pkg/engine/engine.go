package engine

import (
	"fmt"

	"github.com/wiremind/kubectl-restore/pkg/k8screds"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type RestoreOptions struct {
	Namespace     string
	ServiceName   string
	DryRun        bool
	SecretKeyRefs []k8screds.SecretKeyRef
}

type Engine interface {
	Name() string
	Restore(configFlags *genericclioptions.ConfigFlags, backupName string, databaseName string, opts RestoreOptions) error
}

var registry = map[string]Engine{}

func RegisterEngine(e Engine) {
	registry[e.Name()] = e
}

func GetEngine(name string) (Engine, error) {
	e, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown engine: %s", name)
	}
	return e, nil
}
