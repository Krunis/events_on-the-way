package keeperserver

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func insertInOutbox(ctx context.Context, tx pgx.Tx, event driverEvent) error {
	_, err := tx.Exec(ctx, `INSERT INTO outbox (trip_id, driver_id, trip_position, trip_destination)
							VALUES ($1, $2, $3, $4)`,
		event.Trip_ID, event.Driver_ID, event.Trip_Position, event.Destination)

	return err
}

func (k *KeeperServer) InsertEventInDB(ctx context.Context, event driverEvent) error {
	tx, err := k.dbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `INSERT INTO drivers (id, driver_id, position, destination)
				  			  VALUES ($1, $2, $3, $4)
				  			  ON CONFLICT (id) 
							  DO UPDATE 
							  SET position=EXCLUDED.position
							  WHERE trips.position 
							  IS DISTINCT FROM EXCLUDED.position`,
		event.Trip_ID, event.Driver_ID, event.Trip_Position, event.Destination)
	if err != nil {
		return fmt.Errorf("failed to insert in table trips: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}

	if err := insertInOutbox(ctx, tx, event); err != nil {
		return fmt.Errorf("failed to insert in outbox: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
