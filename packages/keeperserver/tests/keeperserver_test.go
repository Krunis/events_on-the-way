package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/Krunis/events_on-the-way/packages/keeperserver"
	"github.com/stretchr/testify/assert"
)

func TestValidateDriverRequest(t *testing.T) {
    tests := []struct {
        name    string
        request keeperserver.RegDriverRequest
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid request",
            request: keeperserver.RegDriverRequest{
                Name:     "John",
                Surname:  "Doe",
                Car_Type: "sedan",
            },
            wantErr: false,
        },
        {
            name: "empty name",
            request: keeperserver.RegDriverRequest{
                Name:     "",
                Surname:  "Doe",
                Car_Type: "sedan",
            },
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name: "only spaces in name",
            request: keeperserver.RegDriverRequest{
                Name:     "   ",
                Surname:  "Doe",
                Car_Type: "sedan",
            },
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name: "all fields empty",
            request: keeperserver.RegDriverRequest{
                Name:     "",
                Surname:  "",
                Car_Type: "",
            },
            wantErr: true,
            errMsg:  "name is required, surname is required, car_type is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := keeperserver.ValidateRegDriverRequest(tt.request)
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateDriverEvent(t *testing.T) {
    tests := []struct {
        name   string
        event  keeperserver.DriverEvent
        wantOk bool
    }{
        {
            name: "valid event",
            event: keeperserver.DriverEvent{
                Trip_ID:       "trip-123",
                Driver_ID:     "driver-456",
                Trip_Position: "prepared",
                Destination:   "Moscow",
            },
            wantOk: true,
        },
        {
            name: "missing trip_id",
            event: keeperserver.DriverEvent{
                Trip_ID:       "",
                Driver_ID:     "driver-456",
                Trip_Position: "prepared",
                Destination:   "Moscow",
            },
            wantOk: false,
        },
        {
            name: "missing destination",
            event: keeperserver.DriverEvent{
                Trip_ID:       "trip-123",
                Driver_ID:     "driver-456",
                Trip_Position: "prepared",
                Destination:   "",
            },
            wantOk: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := keeperserver.ValidateDriverEvent(tt.event)
            if tt.wantOk {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}

func TestFormatDBError(t *testing.T) {
    t.Run("with error", func(t *testing.T) {
        dbErr := errors.New("connection failed")
        resultErr := keeperserver.FormatDBError("insert data", dbErr)
        
        assert.Error(t, resultErr)
        assert.Contains(t, resultErr.Error(), "failed to insert data")
        assert.Contains(t, resultErr.Error(), "connection failed")
        assert.True(t, errors.Is(resultErr, dbErr))
    })
    
    t.Run("nil error", func(t *testing.T) {
        resultErr := keeperserver.FormatDBError("insert data", nil)
        assert.NoError(t, resultErr)
    })
}

func TestNowFunction(t *testing.T) {
    before := time.Now()
    result := keeperserver.Now()
    after := time.Now()
    
    assert.True(t, result.After(before) || result.Equal(before))
    assert.True(t, result.Before(after) || result.Equal(after))
}