package keeperserver

import (
	"context"
	"fmt"
)

func (k *KeeperServer) InsertEventInDB(ctx context.Context, event driverEvent) error {
	tx, err := k.dbPool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start transaction to DB: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `INSERT INTO drivers (id, driver_id, position, destination)
				  VALUES ($1, $2, $3, $4)
				  ON CONFLICT (id) DO NOTHING`, 
				  event.Trip_ID, event.Driver_ID, event.Trip_Position, event.Destination)
	if err != nil{
		return err
	}

	if tag.RowsAffected() == 0{
		return fmt.Errorf("incorrect event: %w", err)
	}

}
