package keeperserver

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func insertInOutbox(ctx context.Context, tx pgx.Tx, event DriverEvent) error {
	_, err := tx.Exec(ctx, `INSERT INTO outbox (event_id, trip_id, driver_id, trip_position, trip_destination, created_at)
							VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), event.Trip_ID, event.Driver_ID, event.Trip_Position, event.Destination, time.Now())
	return err
}

func (k *KeeperServerService) InsertEventInDB(ctx context.Context, event DriverEvent) error {
	if !IsValidTripPosition(event.Trip_Position) {
		return errors.New("trip_position is incorrect")
	}

	tx, err := k.DBPool.Begin(ctx)
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
		return FormatDBError("insert in table trips", err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}

	if err := insertInOutbox(ctx, tx, event); err != nil {
		return FormatDBError("insert in outbox", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return FormatDBError("commit transaction", err)
	}

	return nil
}

func (k *KeeperServerService) RegDriverToDB(ctx context.Context, regDriver RegDriverRequest) (uuid.UUID, error){
	row := k.DBPool.QueryRow(ctx, `INSERT INTO drivers (name_dr, surname_dr, registered_at, car_type)
						VALUES ($1, $2, $3, $4)
						RETURNING id`, regDriver.Name, regDriver.Surname, time.Now(), regDriver.Car_Type)

	var id uuid.UUID

	if err := row.Scan(&id); err != nil{
		return uuid.Nil, FormatDBError("insert data", err)
	}

	return id, nil
}