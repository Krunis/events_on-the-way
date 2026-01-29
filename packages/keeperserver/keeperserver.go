package keeperserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/Krunis/events_on-the-way/packages/common"
	"github.com/jackc/pgx/v4/pgxpool"
)

type driverEvent struct {
	Trip_ID       string `json:"trip_id"`
	Driver_ID     string `json:"driver_id"`
	Trip_Position string `json:"trip_position"`
	Destination   string `json:"destination"`
}

type KeeperServer struct {
	port       string
	httpServer *http.Server
	mux        *http.ServeMux

	dbPool *pgxpool.Pool

	producer *sarama.SyncProducer // Do not use methods in this file

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(serverPort string) *KeeperServer {
	ctx, cancel := context.WithCancel(context.Background())

	mux := http.NewServeMux()

	return &KeeperServer{
		port:   serverPort,
		mux:    mux,
		ctx:    ctx,
		cancel: cancel}
}

func (k *KeeperServer) Start() error {
	var err error

	k.dbPool, err = common.ConnectToDB(k.ctx, common.GetDBConnectionString())
	if err != nil {
		return err
	}

	k.mux.HandleFunc("/new-event", k.newDriverEventHandler)

	k.httpServer.Handler = k.mux
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

func (k *KeeperServer) newDriverEventHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-k.ctx.Done():
		log.Println("Request cancelled (shutdown or client disconnected)")
		return
	default:
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var event driverEvent
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := k.InsertEventInDB(ctx, event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusCreated)
	}

}
