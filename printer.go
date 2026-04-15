package moonraker

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// PrinterService provides access to printer-related endpoints.
type PrinterService struct {
	client *Client
}

// NewPrinterService creates a new printer service.
func NewPrinterService(client *Client) *PrinterService {
	return &PrinterService{client: client}
}

// ListObjects returns all loaded Klipper printer objects.
func (s *PrinterService) ListObjects(ctx context.Context) ([]string, error) {
	var result struct {
		Objects []string `json:"objects"`
	}
	if err := s.client.doRequest(ctx, "GET", "/printer/objects/list", nil, &result); err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}
	return result.Objects, nil
}

// QueryObjectsWithParams is a flexible version for querying objects.
func (s *PrinterService) QueryObjectsWithParams(ctx context.Context, objectNames []string) (*PrinterObjectsQuery, error) {
	params := url.Values{}
	for _, obj := range objectNames {
		params.Add("objects", obj)
	}

	path := "/printer/objects/query?" + params.Encode()

	var result PrinterObjectsQuery
	if err := s.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to query objects: %w", err)
	}
	return &result, nil
}

// RunGCode executes a GCode command.
func (s *PrinterService) RunGCode(ctx context.Context, script string) error {
	body := map[string]string{"script": script}
	var result string
	if err := s.client.doRequest(ctx, "POST", "/printer/gcode/script", body, &result); err != nil {
		return fmt.Errorf("failed to run gcode: %w", err)
	}
	return nil
}

// EmergencyStop immediately halts the printer.
func (s *PrinterService) EmergencyStop(ctx context.Context) error {
	var result string
	if err := s.client.doRequest(ctx, "POST", "/printer/emergency_stop", nil, &result); err != nil {
		return fmt.Errorf("failed to emergency stop: %w", err)
	}
	return nil
}

// GetStatus fetches comprehensive printer status.
//
// Uses the direct object-name-as-query-param format
// (e.g. ?webhooks&print_stats&toolhead) which is required by many
// Moonraker installations (Creality K1, etc.) that return null for
// the standard ?objects=webhooks format.
func (s *PrinterService) GetStatus(ctx context.Context) (*PrinterStatus, error) {
	status := &PrinterStatus{}

	// Objects to query — using the object name as the query param key.
	objects := []string{
		"webhooks",
		"print_stats",
		"display_status",
		"virtual_sdcard",
		"toolhead",
		"fan",
		"extruder",
		"heater_bed",
	}

	params := url.Values{}
	for _, obj := range objects {
		params.Set(obj, "")
	}
	// Build bare param names (no `=`) — required by many Moonraker
	// installations (Creality K1, etc.) that reject `key=` format.
	path := "/printer/objects/query?" + strings.Join(objects, "&")

	var result map[string]interface{}
	if err := s.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to query printer status: %w", err)
	}

	// Parse webhooks
	if v, ok := result["webhooks"].(map[string]interface{}); ok {
		if state, ok := v["state"].(string); ok {
			status.WebhooksState = state
		}
	}

	// Parse print_stats
	if v, ok := result["print_stats"].(map[string]interface{}); ok {
		if state, ok := v["state"].(string); ok {
			status.PrintState = state
		}
		if filename, ok := v["filename"].(string); ok {
			status.Filename = filename
		}
		if message, ok := v["message"].(string); ok {
			status.Message = message
		}
		if pd, ok := v["print_duration"].(float64); ok {
			status.PrintDuration = pd
		}
		if td, ok := v["total_duration"].(float64); ok {
			status.TotalDuration = td
		}
		if fu, ok := v["filament_used"].(float64); ok {
			status.FilamentUsed = fu
		}
	}

	// Parse display_status
	if v, ok := result["display_status"].(map[string]interface{}); ok {
		if progress, ok := v["progress"].(float64); ok {
			status.Progress = float32(progress) * 100
		}
	}

	// Parse virtual_sdcard
	if v, ok := result["virtual_sdcard"].(map[string]interface{}); ok {
		if progress, ok := v["progress"].(float64); ok {
			status.VSDProgress = float32(progress)
		}
		if isActive, ok := v["is_active"].(bool); ok {
			status.VSDIsActive = isActive
		}
	}

	// Parse toolhead
	if v, ok := result["toolhead"].(map[string]interface{}); ok {
		if pos, ok := v["position"].([]interface{}); ok && len(pos) >= 3 {
			status.Position = make([]float64, len(pos))
			for i, p := range pos {
				if val, ok := p.(float64); ok {
					status.Position[i] = val
				}
			}
		}
		if est, ok := v["estimated_print_time"].(float64); ok {
			status.EstimatedPrintTime = est
		}
		if pt, ok := v["print_time"].(float64); ok {
			status.PrintTime = pt
		}
	}

	// Parse fan
	if v, ok := result["fan"].(map[string]interface{}); ok {
		if speed, ok := v["speed"].(float64); ok {
			status.FanSpeed = float32(speed) * 100
		}
	}

	// Parse extruder
	if v, ok := result["extruder"].(map[string]interface{}); ok {
		if temp, ok := v["temperature"].(float64); ok {
			status.NozzleTemperature = float32(temp)
		}
		if target, ok := v["target"].(float64); ok {
			status.NozzleTarget = float32(target)
		}
	}

	// Parse heater_bed
	if v, ok := result["heater_bed"].(map[string]interface{}); ok {
		if temp, ok := v["temperature"].(float64); ok {
			status.BedTemperature = float32(temp)
		}
		if target, ok := v["target"].(float64); ok {
			status.BedTarget = float32(target)
		}
	}

	if status.Progress > 0 && status.PrintDuration > 0 {
		totalTime := status.PrintDuration / float64(status.Progress/100)
		status.TimeLeft = int64(totalTime - status.PrintDuration)
		if status.TimeLeft < 0 {
			status.TimeLeft = 0
		}
	}

	return status, nil
}

// PrinterStatus represents comprehensive printer status.
type PrinterStatus struct {
	WebhooksState      string
	PrintState         string
	Filename           string
	Message            string
	Progress           float32 // Percentage 0-100
	PrintDuration      float64
	TotalDuration      float64
	FilamentUsed       float64
	TimeLeft           int64 // Seconds
	Position           []float64
	EstimatedPrintTime float64
	PrintTime          float64
	FanSpeed           float32 // Percentage 0-100
	BedTemperature     float32
	BedTarget          float32
	NozzleTemperature  float32
	NozzleTarget       float32
	VSDProgress        float32
	VSDIsActive        bool
}
