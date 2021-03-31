package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/qumine/ingress-controller/internal/api"
	"github.com/qumine/ingress-controller/internal/k8s"
	"github.com/qumine/ingress-controller/internal/server"
	"github.com/qumine/ingress-controller/pkg/build"
	"github.com/qumine/ingress-controller/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "ingress-controller",
		Short:   "A Kubernetes ingress controller for minecraft servers",
		Long:    "A Kubernetes ingress controller for minecraft servers",
		Version: build.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cliOptions := config.GetCliOptions()
			logrus.SetLevel(cliOptions.LogLevel)
		},
		Run: func(cmd *cobra.Command, args []string) {
			api := api.NewAPI(config.GetAPIOptions())
			k8s := k8s.NewK8S(config.GetK8SOptions())
			srv := server.NewServer(config.GetIngressOptions())

			context, cancel := context.WithCancel(context.Background())
			defer cancel()

			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

			go api.Start(context, k8s, srv)
			go k8s.Start(context)
			go srv.Start(context)

			<-c
		},
	}
	rootCmd.PersistentFlags().AddFlagSet(config.GetCliFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(config.GetK8SFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(config.GetAPIFlagSet())
	rootCmd.PersistentFlags().AddFlagSet(config.GetIngressFlagSet())
	return rootCmd
}