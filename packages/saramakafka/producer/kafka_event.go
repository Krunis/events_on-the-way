package producer

import (
	"encoding/json"
	"time"

	"github.com/Krunis/events_on-the-way/packages/types"
)
type KafkaEvent struct {
	*types.Row
	Timestamp     time.Time
}

func NewKafkaEvent(row *types.Row) *KafkaEvent{
	return &KafkaEvent{
		Row: row,
		Timestamp: time.Now(),
	}	
}

func (e *KafkaEvent) Key() []byte{
	return []byte(e.Trip_ID)
}

func (e *KafkaEvent) Value() []byte{
	data, _ := json.Marshal(e.Row)
	return data
}