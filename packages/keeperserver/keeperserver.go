package keeperserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/IBM/sarama"
	"github.com/Krunis/events_on-the-way/packages/common"
	"github.com/jackc/pgx/v4/pgxpool"
)

type driverEvent struct {
	Trip_ID            string `json:"trip_id"`
	Driver_ID     string `json:"driver_id"`
	Trip_Position string `json:"trip_position"`
	Destination   string `json:"destination"`
}

type KeeperServer struct {
	port       string
	httpServer *http.Server

	dbPool *pgxpool.Pool

	producer *sarama.SyncProducer // Do not use methods in this file

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(serverPort string) *KeeperServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &KeeperServer{port: serverPort, ctx: ctx, cancel: cancel}
}

func (k *KeeperServer) Start() error {
	pool, err := common.ConnectToDB(k.ctx, common.GetDBConnectionString())
	if err != nil {
		return err
	}

	k.dbPool = pool

	k.httpServer.Addr = k.port

	errCh := make(chan error, 1)

	go func() {
		if err := k.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-k.ctx.Done():
		return nil
	}
}

func (k *KeeperServer) driverEventHandler(w http.ResponseWriter, r *http.Request) {
	var event driverEvent

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	k.InsertEventInDB(, event)

}
