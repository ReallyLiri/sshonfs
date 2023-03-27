package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type serveFn func(*Config) (io.Closer, error)

func Execute(serve serveFn) {
	runCmd := &cobra.Command{
		Use:     "sshonfs",
		Short:   "access remote fs using ssh on top of nfs",
		Version: "1.0",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				log.Fatalf("could not bind flags: %v", err)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var config Config
			err := viper.Unmarshal(&config)
			if err != nil {
				log.Fatalf("failed to unmarshal config: %v", err)
			}

			var closer io.Closer

			closer, err = serve(&config)

			if err != nil {
				log.Fatalf("failed: %v", err)
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan
			log.Println("stopping")
			go func() {
				// time to gracefully close, otherwise, kill
				time.Sleep(3 * time.Second)
				panic("timeout exceeded")
			}()
			closer.Close()
			os.Exit(0)
		},
	}

	runCmd.Flags().StringP("address", "a", "127.0.0.1:22", "ssh server address")
	runCmd.Flags().StringP("username", "u", "root", "ssh username")
	runCmd.Flags().StringP("password", "p", "", "ssh password")
	runCmd.Flags().StringP("root", "r", "/opt", "ssh root")
	runCmd.Flags().StringP("private-key", "i", os.Getenv("HOME")+`/.ssh/id_rsa`, "path to private ssh key")
	runCmd.Flags().StringP("serve-port", "P", "2049", "local port to serve nfs server on")
	runCmd.Flags().BoolP("skip-mount", "s", false, "skip mount, only serve")
	runCmd.Flags().StringP("mount-options", "o", "", "options to mount with, default options are OS dependent")
	runCmd.Flags().StringP("mount-path", "m", ".", "path to mount the ssh fs on")

	if err := runCmd.Execute(); err != nil {
		log.Fatalf("error executing command: %v", err)
	}
}
