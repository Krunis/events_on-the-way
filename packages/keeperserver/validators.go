package keeperserver

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Krunis/events_on-the-way/packages/common"
)


func ValidateRegDriverRequest(req RegDriverRequest) error {
    var errorSlice []string
    
    if strings.TrimSpace(req.Name) == "" {
        errorSlice = append(errorSlice, "name is required")
    }
    if strings.TrimSpace(req.Surname) == "" {
        errorSlice = append(errorSlice, "surname is required")
    }
    if strings.TrimSpace(req.Car_Type) == "" {
        errorSlice = append(errorSlice, "car_type is required")
    }
    
    if len(errorSlice) > 0 {
        return errors.New(strings.Join(errorSlice, ", "))
    }
    return nil
}

func ValidateDriverEvent(event DriverEvent) error {
    if strings.TrimSpace(event.Trip_ID) == "" {
        return fmt.Errorf("trip_id is required")
    }
    if strings.TrimSpace(event.Driver_ID) == "" {
        return fmt.Errorf("driver_id is required")
    }
    if strings.TrimSpace(event.Destination) == "" {
        return fmt.Errorf("destination is required")
    }
    return nil
}

func FormatDBError(operation string, err error) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("failed to %s: %w", operation, err)
}

func IsValidTripPosition(pos string)bool{
	switch pos{
	case common.Prepared, common.OnTheWay, common.Arrived, common.Completed:
		return true
	default:
		return false
	}
}

var Now = time.Now

