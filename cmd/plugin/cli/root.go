package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj/go-spin"
	"github.com/wiremind/kubectl-db-restore/pkg/logger"
	"github.com/wiremind/kubectl-db-restore/pkg/plugin"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

func shouldRunRestore() bool {
	return engineName != "" && backupName != "" && databaseName != "" && serviceName != ""
}

func validateRestoreFlags() error {
	missing := []string{}

	if engineName == "" {
		missing = append(missing, "--engine")
	}
	if backupName == "" {
		missing = append(missing, "--backup-name")
	}
	if databaseName == "" {
		missing = append(missing, "--database")
	}
	if serviceName == "" {
		missing = append(missing, "--service-name")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required flag(s) to run restore job: %s", strings.Join(missing, ", "))
	}
	return nil
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "kubectl-db-restore",
		Short:         "",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				panic(fmt.Errorf("failed to bind flags: %w", err))
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateRestoreFlags(); err == nil {
				return runDatabaseRestore()
			} else if engineName != "" || backupName != "" || databaseName != "" || serviceName != "" {
				// Some flags were set, but not all — show helpful error
				return err
			}
			// If no flags and no args, show help
			if len(os.Args) == 1 {
				return cmd.Help()
			}
			log := logger.NewLogger()
			log.Info("")

			s := spin.New()
			finishedCh := make(chan bool, 1)
			namespaceName := make(chan string, 1)
			go func() {
				lastNamespaceName := ""
				for {
					select {
					case <-finishedCh:
						fmt.Printf("\r")
						return
					case n := <-namespaceName:
						lastNamespaceName = n
					case <-time.After(time.Millisecond * 100):
						if lastNamespaceName == "" {
							fmt.Printf("\r  \033[36mSearching for namespaces\033[m %s", s.Next())
						} else {
							fmt.Printf("\r  \033[36mSearching for namespaces\033[m %s (%s)", s.Next(), lastNamespaceName)
						}
					}
				}
			}()
			defer func() {
				finishedCh <- true
			}()

			if err := plugin.RunPlugin(KubernetesConfigFlags, namespaceName); err != nil {
				return errors.Unwrap(err)
			}

			log.Info("")

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cmd.Flags().StringVar(&engineName, "engine", "", "Database engine (clickhouse, postgres, ...)")
	cmd.Flags().StringVar(&backupName, "backup-name", "", "Backup name")
	cmd.Flags().StringVar(&databaseName, "database", "", "Database name")
	cmd.Flags().StringVar(&serviceName, "service-name", "", "Kubernetes service name for DB")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run")
	cmd.Flags().StringSliceVar(&secretRefs, "secret-ref", nil, "Secret reference in the format VAR=secretName:key (can be repeated)")

	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		osExit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
