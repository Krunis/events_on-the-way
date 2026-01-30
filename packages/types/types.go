package types

type Row struct {
	Event_ID      string `json:"event_id"`
	Trip_ID       string `json:"trip_id"`
	Driver_ID     string `json:"driver_id"`
	Trip_Position string `json:"trip_position"`
	Trip_Destination   string `json:"destination"`
}