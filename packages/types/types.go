package types

import (
	"github.com/google/uuid"
)

type Row struct {
	Event_ID         uuid.UUID `json:"event_id"`
	Trip_ID          string `json:"trip_id"`
	Driver_ID        uuid.UUID `json:"driver_id"`
	Trip_Position    string `json:"trip_position"`
	Trip_Destination string `json:"destination"`
}