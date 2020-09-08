package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ckeyer/tarofs/cmd/internal/inner"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debug  bool
	chExit = make(chan struct{})

	rootCmd = cobra.Command{
		Use:   "tarofs",
		Short: "",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// logrus.SetFormatter(&logrus.JSONFormatter{})
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			<-chExit
			logrus.Info("exit 0")
		},
	}
)

// Execute root command
func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "print debug message.")

	rootCmd.AddCommand(inner.Command()...)
	rootCmd.Execute()
}

func waitExec(f func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	startTimeout := false
	for {
		select {
		case <-sigChan:
			startTimeout = true
			logrus.Info("wait exit.")
			go func() {
				f()
				close(chExit)
			}()
		case <-time.Tick(time.Second * 10):
			if startTimeout {
				os.Exit(1)
			}
		}
	}
}
