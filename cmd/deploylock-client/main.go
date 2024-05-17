package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/distributed_locker"
)

func init() {
	AcquireCmd.Flags().StringP("name", "n", "deploy", "lock name")
	RenewCmd.Flags().StringP("name", "n", "deploy", "lock name")
	RenewCmd.Flags().StringP("lease", "l", "", "lease uuid")

	RenewCmd.MarkFlagRequired("lease")

	RootCmd.AddCommand(AcquireCmd)
	RootCmd.AddCommand(RenewCmd)
}

var RootCmd = &cobra.Command{
	Use: "deploylock-client",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var RenewCmd = &cobra.Command{
	Use:   "renew --uuid <lease> <server>",
	Short: "Renew a distributed lock lease for deploying to Salesforce",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		server, err := url.Parse(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not parse server: %s", err.Error())
			os.Exit(1)
		}
		serverUrl := fmt.Sprintf("%s://%s", server.Scheme, server.Host)
		lockName, _ := cmd.Flags().GetString("name")
		uuid, _ := cmd.Flags().GetString("lease")
		backend := distributed_locker.NewHttpBackend(serverUrl)
		l := distributed_locker.NewDistributedLocker(backend)

		done := make(chan struct{})
		go func() {
			defer close(done)
			l.HoldLease(lockName, uuid)
			fmt.Fprintf(os.Stderr, "HoldLease finished\n")
		}()
		defer l.Release(lockgate.LockHandle{UUID: uuid, LockName: lockName})
		select {
		case <-cmd.Context().Done():
			fmt.Fprintf(os.Stderr, cmd.Context().Err().Error())
			return
		case <-done:
			fmt.Fprintf(os.Stderr, "done channel closed\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "returning from Run\n")
	},
}

var AcquireCmd = &cobra.Command{
	Use:   "acquire <server>",
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
			AcquirerId: uuid.New().String(),
			OnWaitFunc: func(lockName string, doWait func() error) error {
				done := make(chan struct{})
				ticker := time.NewTicker(3 * time.Second)
				defer ticker.Stop()
				go func() {
					for {
						fmt.Fprintf(os.Stderr, "WAITING FOR %s\n", lockName)
						select {
						case <-done:
						case <-ticker.C:
						}
					}
				}()
				defer close(done)
				if err := doWait(); err != nil {
					fmt.Fprintf(os.Stderr, "WAITING FOR %s FAILED: %s\n", lockName, err)
					return err
				} else {
					fmt.Fprintf(os.Stderr, "WAITING FOR %s DONE\n", lockName)
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

		fmt.Println(lockHandle.UUID)
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
