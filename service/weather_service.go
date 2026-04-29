package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/huimingz/hyweb-assessment/model"
)

const (
	cwaBaseURL          = "https://opendata.cwa.gov.tw/api/v1/rest/datastore/F-C0032-001"
	weatherLocationName = "新北市"
)

type cwaResp struct {
	Records struct {
		Location []cwaLocation `json:"location"`
	} `json:"records"`
}

type cwaLocation struct {
	LocationName   string       `json:"locationName"`
	WeatherElement []cwaElement `json:"weatherElement"`
}

type cwaElement struct {
	ElementName string    `json:"elementName"`
	Time        []cwaTime `json:"time"`
}

type cwaTime struct {
	StartTime string `json:"startTime"`
	Parameter struct {
		ParameterName string `json:"parameterName"`
	} `json:"parameter"`
}

type WeatherService struct {
	db     *sql.DB
	logger *slog.Logger
	apiKey string
	client *http.Client
}

func NewWeatherService(db *sql.DB, logger *slog.Logger, apiKey string, client *http.Client) *WeatherService {
	return &WeatherService{db: db, logger: logger, apiKey: apiKey, client: client}
}

func (s *WeatherService) FetchAndStore(ctx context.Context) {
	if s.apiKey == "" {
		s.logger.WarnContext(ctx, "weather fetch skipped: WEATHER_API_KEY is not set")
		return
	}

	url := fmt.Sprintf("%s?Authorization=%s&locationName=%s&elementName=MinT&elementName=MaxT",
		cwaBaseURL, s.apiKey, weatherLocationName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.logger.ErrorContext(ctx, "weather fetch: failed to build request", "error", err)
		return
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.ErrorContext(ctx, "weather fetch: http request failed", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.ErrorContext(ctx, "weather fetch: unexpected status", "status", resp.StatusCode)
		return
	}

	var cwa cwaResp
	if err := json.NewDecoder(resp.Body).Decode(&cwa); err != nil {
		s.logger.ErrorContext(ctx, "weather fetch: failed to decode response", "error", err)
		return
	}

	var loc *cwaLocation
	for i := range cwa.Records.Location {
		if cwa.Records.Location[i].LocationName == weatherLocationName {
			loc = &cwa.Records.Location[i]
			break
		}
	}
	if loc == nil {
		s.logger.WarnContext(ctx, "weather fetch: 新北市 not found in response")
		return
	}

	elements := make(map[string][]cwaTime, len(loc.WeatherElement))
	for _, el := range loc.WeatherElement {
		elements[el.ElementName] = el.Time
	}

	periods := []struct {
		label string
		idx   int
	}{
		{"AM", 0},
		{"PM", 1},
	}

	for _, p := range periods {
		minTimes, okMin := elements["MinT"]
		maxTimes, okMax := elements["MaxT"]
		if !okMin || !okMax || p.idx >= len(minTimes) || p.idx >= len(maxTimes) {
			s.logger.WarnContext(ctx, "weather fetch: missing data for period", "period", p.label)
			continue
		}

		minT, err := strconv.ParseFloat(minTimes[p.idx].Parameter.ParameterName, 64)
		if err != nil {
			s.logger.ErrorContext(ctx, "weather fetch: failed to parse MinT", "period", p.label, "error", err)
			continue
		}
		maxT, err := strconv.ParseFloat(maxTimes[p.idx].Parameter.ParameterName, 64)
		if err != nil {
			s.logger.ErrorContext(ctx, "weather fetch: failed to parse MaxT", "period", p.label, "error", err)
			continue
		}

		_, err = s.db.ExecContext(ctx, `
			INSERT INTO weather (city, min_t, max_t, period, date)
			VALUES ('新北市', ?, ?, ?, CURDATE())
			ON DUPLICATE KEY UPDATE min_t = VALUES(min_t), max_t = VALUES(max_t), updated = NOW()
		`, minT, maxT, p.label)
		if err != nil {
			s.logger.ErrorContext(ctx, "weather fetch: upsert failed", "period", p.label, "error", err)
			continue
		}
	}

	s.logger.InfoContext(ctx, "weather fetch: stored successfully", "city", weatherLocationName)
}

func (s *WeatherService) GetTodayWeather(ctx context.Context) (*model.WeatherResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT period, min_t, max_t FROM weather WHERE city = '新北市' AND date = CURDATE()",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &model.WeatherResponse{}
	for rows.Next() {
		var period string
		var minT, maxT float64
		if err := rows.Scan(&period, &minT, &maxT); err != nil {
			return nil, err
		}
		wp := &model.WeatherPeriod{MinT: minT, MaxT: maxT}
		switch period {
		case "AM":
			result.AM = wp
		case "PM":
			result.PM = wp
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
