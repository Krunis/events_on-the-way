package polleroutbox

import (
	"context"
	"log"

	"github.com/Krunis/events_on-the-way/packages/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func (p *PollerOutboxService) GetRowsFromOutbox(ctx context.Context, tx pgx.Tx) ([]*types.Row, error) {
	rows, err := tx.Query(ctx, `SELECT event_id, trip_id, driver_id, trip_position, trip_destination
						 		FROM outbox
						 		WHERE event_status='NEW'
						 		ORDER BY created_at
						 		LIMIT $1
						 		FOR UPDATE SKIP LOCKED`,
		p.cfg.BatchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*types.Row{}

	for rows.Next() {
		row := types.Row{}

		err := rows.Scan(&row.Event_ID, &row.Trip_ID, &row.Driver_ID, &row.Trip_Position, &row.Trip_Destination)
		if err != nil {
			log.Printf("Error while scanning row with event_id=%s: %s", row.Event_ID, err)
			continue
		}

		result = append(result, &row)

	}

	return result, nil
}

func (p *PollerOutboxService) MarkAsSentOutbox(ctx context.Context, tx pgx.Tx, rows []*types.Row) error {
	ids := make([]uuid.UUID, len(rows))

	for i, row := range rows {
		ids[i] = row.Event_ID
	}

	log.Println(ids)

	_, err := tx.Exec(ctx, `UPDATE outbox
						   		  SET event_status='SENT'
						   		  WHERE event_id=ANY($1)`,
		ids)
	if err != nil {
		return err
	}
	
	log.Println("Updated")

	return nil
}
