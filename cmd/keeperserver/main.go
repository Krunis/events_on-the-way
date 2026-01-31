package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Krunis/events_on-the-way/packages/common"
	"github.com/Krunis/events_on-the-way/packages/keeperserver"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	keeper := keeperserver.NewKeeperServerService(":8081")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	errCh := make(chan error, 1)

	go func(){
		if err := keeper.Start(common.GetDBConnectionString()); err != nil{
			errCh <- err
		}
	}()

	select{
	case <-stop:
		if err := keeper.Stop(); err != nil{
			log.Printf("Error while stopping: %s\n", err)
		}
	case err := <-errCh:
		if err != nil{
			log.Printf("Error while serving: %s", err)
		}

		if err := keeper.Stop(); err != nil{
			log.Printf("Error while stop: %s", err)
		}
	case <-ctx.Done():
		return
	}

}