package keeperserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Krunis/events_on-the-way/packages/common"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func insertInOutbox(ctx context.Context, tx pgx.Tx, event driverEvent) error {
	_, err := tx.Exec(ctx, `INSERT INTO outbox (event_id, trip_id, driver_id, trip_position, trip_destination, created_at)
							VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), event.Trip_ID, event.Driver_ID, event.Trip_Position, event.Destination, time.Now())
	return err
}

func (k *KeeperServerService) InsertEventInDB(ctx context.Context, event driverEvent) error {
	if !common.IsValidTripPosition(event.Trip_Position) {
		return errors.New("trip_position is incorrect")
	}
	tx, err := k.dbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `INSERT INTO trips (id, driver_id, position, destination, started_at)
				  			  VALUES ($1, $2, $3, $4, $5)
				  			  ON CONFLICT (id) 
							  DO UPDATE SET 
							  	position=EXCLUDED.position
								finished_at=CASE 
            						WHEN EXCLUDED.position = 'completed' AND trips.finished_at IS NULL
            						THEN CURRENT_TIMESTAMP
        						END
							  WHERE trips.position 
							  IS DISTINCT FROM EXCLUDED.position`,
		event.Trip_ID, event.Driver_ID, event.Trip_Position, event.Destination, time.Now())
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

func (k *KeeperServerService) RegDriverToDB(ctx context.Context, regDriver regDriverRequest) (uuid.UUID, error){
	row := k.dbPool.QueryRow(ctx, `INSERT INTO drivers (name_dr, surname_dr, registered_at, car_type)
						VALUES ($1, $2, $3, $4)
						RETURNING id`, regDriver.Name, regDriver.Surname, time.Now(), regDriver.Car_Type)

	var id uuid.UUID

	if err := row.Scan(&id); err != nil{
		return uuid.Nil, fmt.Errorf("failed to insert data: %s", err)
	}

	return id, nil
}