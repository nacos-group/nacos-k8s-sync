package cmd

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/nacos-group/nacos-k8s-sync/pkg/logger"
)

// WaitSignal awaits for SIGINT or SIGTERM and closes the channel
func WaitSignal(stop chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	close(stop)
	logger.Sync()
}

// AddFlags adds all command line flags to the given command.
func AddFlags(rootCmd *cobra.Command) {
	rootCmd.Flags().AddGoFlagSet(flag.CommandLine)
}

// PrintFlags logs the flags in the flagset
func PrintFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		logger.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}
