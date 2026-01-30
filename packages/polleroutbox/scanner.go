package polleroutbox

import (
	"context"
	"log"
	"time"

	"github.com/Krunis/events_on-the-way/packages/types"
)

func (p *PollerOutboxService) GetRowsFromOutbox() ([]types.Row, error) {
	ctx, cancel := context.WithTimeout(p.lifecycle.ctx, time.Second*5)
	defer cancel()

	rows, err := p.dbPool.Query(ctx, `SELECT event_id, trip_id, driver_id, trip_position, trip_destination
						 			  FROM outbox
						 			  WHERE event_status="NEW"
						 			  ORDER BY created_at
						 			  LIMIT $1
						 			  FOR UPDATE SKIP LOCKED`,
									  p.cfg.batchSize)
	if err != nil{
		return nil, err
	}
	defer rows.Close()

	result := []types.Row{}

	for rows.Next(){
		row := types.Row{}

		err := rows.Scan(row.Event_ID, row.Trip_ID, row.Driver_ID, row.Trip_Position, row.Trip_Destination)
		if err != nil{
			log.Printf("Error while scanning row with event_id=%s: %s", row.Event_ID, err)
			continue
		}

		result = append(result, row)

	}

	return result, nil
}
