package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/distributed_locker"
)

func init() {
	RootCmd.Flags().StringP("name", "n", "deploy", "lock name")
}

var RootCmd = &cobra.Command{
	Use:   "deploylock-client <server>",
	Short: "Get a distributed lock lease for deploying to Salesforce",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		server, err := url.Parse(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not parse server: %s", err.Error())
			os.Exit(1)
		}
		serverUrl := fmt.Sprintf("%s://%s", server.Scheme, server.Host)
		lockName, _ := cmd.Flags().GetString("name")
		backend := distributed_locker.NewHttpBackend(serverUrl)
		l := distributed_locker.NewDistributedLocker(backend)

		acquired, lockHandle, err := l.Acquire(lockName, lockgate.AcquireOptions{
			OnWaitFunc: func(lockName string, doWait func() error) error {
				fmt.Printf("WAITING FOR %s\n", lockName)
				if err := doWait(); err != nil {
					fmt.Printf("WAITING FOR %s FAILED: %s\n", lockName, err)
					return err
				} else {
					fmt.Printf("WAITING FOR %s DONE\n", lockName)
				}
				return nil
			},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error acquiring lock: %s", err.Error())
			os.Exit(1)
		} else if !acquired {
			fmt.Fprintf(os.Stderr, "lock %s not acquired", lockName)
			os.Exit(1)
		}

		fmt.Printf("acquired a lock: %#v\n", lockHandle)

		defer l.Release(lockHandle)

		select {
		case <-cmd.Context().Done():
			fmt.Fprintf(os.Stderr, cmd.Context().Err().Error())
		}
	},
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cancelUponSignal(cancel)
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
func cancelUponSignal(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		interuptsReceived := 0
		for {
			<-sigs
			if interuptsReceived > 0 {
				os.Exit(1)
			}
			log.Warnln("signal received.  cancelling.")
			cancel()
			interuptsReceived++
		}
	}()
}
