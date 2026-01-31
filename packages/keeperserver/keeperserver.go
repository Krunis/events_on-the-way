package keeperserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Krunis/events_on-the-way/packages/common"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type driverEvent struct {
	Trip_ID       string `json:"trip_id"`
	Driver_ID     string `json:"driver_id"`
	Trip_Position string `json:"trip_position"`
	Destination   string `json:"destination"`
}

type regDriverRequest struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Car_Type string `json:"car_type"`
}

type KeeperServerService struct {
	port       string
	httpServer *http.Server
	mux        *http.ServeMux

	dbPool *pgxpool.Pool

	lifecycle struct {
		ctx    context.Context
		cancel context.CancelFunc
	}

	stopOnce sync.Once
}

func NewKeeperServerService(serverPort string) *KeeperServerService {
	mux := http.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())

	return &KeeperServerService{
		port: serverPort,
		mux:  mux,
		lifecycle: struct {
			ctx    context.Context
			cancel context.CancelFunc
		}{
			ctx:    ctx,
			cancel: cancel},
	}
}

func (k *KeeperServerService) Start(dbConnectionString string) error {
	var err error

	k.dbPool, err = common.ConnectToDB(k.lifecycle.ctx, dbConnectionString)
	if err != nil {
		return err
	}

	k.mux.HandleFunc("/reg", k.regMockDriverHandler)
	k.mux.HandleFunc("/new-event", k.newDriverEventHandler)

	k.httpServer = &http.Server{}

	k.httpServer.Handler = k.mux
	k.httpServer.Addr = k.port

	errCh := make(chan error, 1)

	log.Println("Starting listening on :8081...")
	go func() {
		if err := k.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-k.lifecycle.ctx.Done():
		return nil
	}
}

func (k *KeeperServerService) regMockDriverHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-k.lifecycle.ctx.Done():
		log.Println("Request cancelled (shutdown or client disconnected)")
		return
	default:
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		regDriver := regDriverRequest{}

		err := json.NewDecoder(r.Body).Decode(&regDriver)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		log.Printf("Received reg request: %v\n", regDriver)

		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 2)
		defer cancel()

		id, err := k.RegDriverToDB(ctx, regDriver)
		if err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Println(id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(struct{ID uuid.UUID}{ID: id})
	}
}

func (k *KeeperServerService) newDriverEventHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-k.lifecycle.ctx.Done():
		log.Println("Request cancelled (shutdown or client disconnected)")
		return
	default:
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var event driverEvent

		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Received event: %v\n", event)

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
		defer cancel()

		if err := k.InsertEventInDB(ctx, event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}

}

func (k *KeeperServerService) Stop() error {
	var err error

	k.stopOnce.Do(func() {
		k.lifecycle.cancel()

		if k.httpServer != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			log.Println("Graceful shutdown...")
			if err = k.httpServer.Shutdown(shutdownCtx); err != nil {
				log.Printf("Graceful shutdown failed: %s\n", err)

				if err = k.httpServer.Close(); err != nil {
					log.Printf("Force close failed: %s\n", err)
					err = fmt.Errorf("shutdown failed: %v, close failed: %v", err, err)
					return
				}
				err = fmt.Errorf("shutdown failed: %v, forced close", err)
			}
		}
		log.Println("Shutdown completed")

		if k.dbPool != nil {
			k.dbPool.Close()
			log.Println("Database pool stopped")
		}

	})

	return err
}
