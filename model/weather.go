package model

import "time"

type Weather struct {
	ID      uint64    `json:"id"`
	City    string    `json:"city"`
	MinT    float64   `json:"min_t"`
	MaxT    float64   `json:"max_t"`
	Period  string    `json:"period"`
	Date    time.Time `json:"date"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type WeatherPeriod struct {
	MinT float64 `json:"min_t"`
	MaxT float64 `json:"max_t"`
}

type WeatherResponse struct {
	AM *WeatherPeriod `json:"AM"`
	PM *WeatherPeriod `json:"PM"`
}
