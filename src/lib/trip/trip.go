package trip

import (
	"../geoloc"

	"fmt"
	"time"
)

type Trip struct {
	From        geoloc.Location
	To          geoloc.Location
	MidSteps    []geoloc.Location
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	VehicleType string
	FuelType    string
	ServiceName string
	CostInCents int
	ScrapedApp  string
}

func (t Trip) TimeTable() string {
	return fmt.Sprintf("Service: %6s, Departure: %10v, Arrival: %10v, Duration: %6v",
		t.ScrapedApp,
		t.StartTime.Format("02/01/2006, 15:04:05"),
		t.EndTime.Format("02/01/2006, 15:04:05"),
		t.Duration)
}