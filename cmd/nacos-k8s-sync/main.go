package main

import (
	"fmt"
	"os"

	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"

	"github.com/nacos-group/nacos-k8s-sync/cmd"
	"github.com/nacos-group/nacos-k8s-sync/pkg/bootstrap"
	"github.com/nacos-group/nacos-k8s-sync/pkg/logger"
	"github.com/nacos-group/nacos-k8s-sync/pkg/model"
)

var (
	options       bootstrap.Options
	loggerOptions = logger.DefaultOptions()

	rootCmd = &cobra.Command{
		Use:     "nacos-k8s-sync",
		Short:   "Sync service information between k8s and nacos.",
		Args:    cobra.NoArgs,
		PreRunE: configureLogging,
		RunE: func(c *cobra.Command, args []string) error {
			cmd.PrintFlags(c.Flags())

			stop := make(chan struct{})
			server, err := bootstrap.NewServer(options)
			if err != nil {
				return err
			}
			server.Run(stop)

			cmd.WaitSignal(stop)
			return nil
		},
	}
)

func configureLogging(_ *cobra.Command, _ []string) error {
	if err := logger.Configure(loggerOptions); err != nil {
		return err
	}
	return nil
}

func init() {
	// TODO Use configmap to avoid so many cmd line args.
	rootCmd.Flags().StringVar(&options.KubeOptions.KubeConfig, "kubeconfig", "",
		"Use a Kubernetes configuration file instead of in-cluster configuration.")

	rootCmd.Flags().StringVarP(&options.KubeOptions.WatchedNamespace, "appNamespace", "a", v1.NamespaceAll,
		"Specify the namespace in where the service source should be synced to nacos.")

	rootCmd.Flags().StringVar(&options.NacosOptions.Namespace, "nacosNamespace", constant.DEFAULT_NAMESPACE_ID,
		"Specify the namespace to which the service in naocs should be stored.")

	rootCmd.Flags().StringSliceVar(&options.NacosOptions.ServersIP, "serversIP", nil,
		"serversIP are explicitly specified to be connected to nacos by client.")

	// TODO set default port
	rootCmd.Flags().Uint64Var(&options.NacosOptions.ServerPort, "serverPort", 0,
		"serverPort are explicitly specified to be used when the client connects to nacos.")

	rootCmd.Flags().StringVar((*string)(&options.Direction), "direction", string(model.ToNacos),
		"Specify the direction of sync which can be to-nacos, to-k8s, or both")

	loggerOptions.AttachCobraFlags(rootCmd)

	cmd.AddFlags(rootCmd)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
