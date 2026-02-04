package keeperserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DriverEvent struct {
	Trip_ID       string `json:"trip_id"`
	Driver_ID     string `json:"driver_id"`
	Trip_Position string `json:"trip_position"`
	Destination   string `json:"destination"`
}

type RegDriverRequest struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Car_Type string `json:"car_type"`
}

type KeeperServerService struct {
	port       string
	httpServer *http.Server
	mux        *http.ServeMux

	DBPool *pgxpool.Pool

	Lifecycle struct {
		Ctx    context.Context
		Cancel context.CancelFunc
	}

	stopOnce sync.Once
}

func NewKeeperServerService(serverPort string) *KeeperServerService {
	mux := http.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())

	return &KeeperServerService{
		port: serverPort,
		mux:  mux,
		Lifecycle: struct {
			Ctx    context.Context
			Cancel context.CancelFunc
		}{
			Ctx:    ctx,
			Cancel: cancel},
	}
}

func (k *KeeperServerService) Start() error {
	k.mux.HandleFunc("/reg", k.RegDriverHandler)
	k.mux.HandleFunc("/new-event", k.NewDriverEventHandler)

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
	case <-k.Lifecycle.Ctx.Done():
		return nil
	}
}

func (k *KeeperServerService) RegDriverHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-k.Lifecycle.Ctx.Done():
		log.Println("Request cancelled (shutdown or client disconnected)")
		return
	default:
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		regDriver := RegDriverRequest{}

		err := json.NewDecoder(r.Body).Decode(&regDriver)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		log.Printf("Received reg request: %v\n", regDriver)

		if err := ValidateRegDriverRequest(regDriver); err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

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

func (k *KeeperServerService) NewDriverEventHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-k.Lifecycle.Ctx.Done():
		log.Println("Request cancelled (shutdown or client disconnected)")
		return
	default:
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var event DriverEvent

		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Received event: %v\n", event)

		if err := ValidateDriverEvent(event); err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

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
		k.Lifecycle.Cancel()

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

		if k.DBPool != nil {
			k.DBPool.Close()
			log.Println("Database pool stopped")
		}

	})

	return err
}
