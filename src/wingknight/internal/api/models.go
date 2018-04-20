package api

import (
	"fmt"
	"time"

	rrule "github.com/teambition/rrule-go"
)

// WingFlavor is a type of wing that is available
type WingFlavor struct {
	Name string `json:"name"` // flavor name on the menu
	Heat int    `json:"heat"` // spiciness of the wing
}

func (f WingFlavor) String() string {
	return fmt.Sprintf("%s", f.Name)
}

// WingNight details a when a deal occurs at a resturant
type WingNight struct {
	Name        string `json:"name"`  // event name
	Description string `json:"desc"`  // detailed description about the wing night
	RRule       string `json:"rrule"` // calendar rule for when the wing night occurs
}

// IsTonightWingNight indicates if the deal applies to tonight so you can go have some wings
func (d WingNight) IsTonightWingNight() bool {
	rule, err := rrule.StrToRRule(d.RRule)
	if err != nil {
		return false
	}
	t := time.Now()
	// determine if a wing night is occuring today
	return len(rule.Between(
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Add(time.Hour*24),
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local),
		true,
	)) > 0
}

// Wing menu definition is used to list prices for quantities that that a restaurant will do
type Wing struct {
	Name  string  `json:"name"`  // name of the menu item
	Price float32 `json:"price"` // price per wing
}

// Restaurant objects are a geographic locations that serves wings and may have Wing Nights
type Restaurant struct {
	Name    string       `json:"name"`
	Address string       `json:"address"`
	Menu    []Wing       `json:"menu"`
	Deals   []WingNight  `json:"deals"`
	Flavors []WingFlavor `json:"flavors"`
}

// GeoJSON file representation of each of the restaurants and where they're located
// these are stored in out WingDB and are the keys in the geotable of redis
type GeoJSON struct {
	Type     string `json:"type"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	}
	Restaurant Restaurant `json:"properties"`
}
