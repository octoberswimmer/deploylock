package main

import (
	"fmt"
	"os"

	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/lockgate/pkg/distributed_locker/optimistic_locking_store"
)

func run(port string) error {
	store := optimistic_locking_store.NewInMemoryStore()
	backend := distributed_locker.NewOptimisticLockingStorageBasedBackend(store)
	return distributed_locker.RunHttpBackendServer("", port, backend)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	if err := run(port); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
