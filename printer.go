package moonraker

import (
	"context"
	"fmt"
	"net/url"
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
func (s *PrinterService) GetStatus(ctx context.Context) (*PrinterStatus, error) {
	objects := []string{
		"webhooks",
		"print_stats",
		"display_status",
		"virtual_sdcard",
		"toolhead",
		"fan",
	}

	result, err := s.QueryObjectsWithParams(ctx, objects)
	if err != nil {
		return nil, err
	}

	status := &PrinterStatus{}

	if webhooks, ok := result.Status["webhooks"]; ok {
		if wh, ok := webhooks.(map[string]interface{}); ok {
			if state, ok := wh["state"].(string); ok {
				status.WebhooksState = state
			}
		}
	}

	if printStats, ok := result.Status["print_stats"]; ok {
		if ps, ok := printStats.(map[string]interface{}); ok {
			if state, ok := ps["state"].(string); ok {
				status.PrintState = state
			}
			if filename, ok := ps["filename"].(string); ok {
				status.Filename = filename
			}
			if message, ok := ps["message"].(string); ok {
				status.Message = message
			}
			if printDuration, ok := ps["print_duration"].(float64); ok {
				status.PrintDuration = printDuration
			}
			if totalDuration, ok := ps["total_duration"].(float64); ok {
				status.TotalDuration = totalDuration
			}
			if filamentUsed, ok := ps["filament_used"].(float64); ok {
				status.FilamentUsed = filamentUsed
			}
		}
	}

	if displayStatus, ok := result.Status["display_status"]; ok {
		if ds, ok := displayStatus.(map[string]interface{}); ok {
			if progress, ok := ds["progress"].(float64); ok {
				status.Progress = float32(progress) * 100
			}
		}
	}

	if vsd, ok := result.Status["virtual_sdcard"]; ok {
		if v, ok := vsd.(map[string]interface{}); ok {
			if progress, ok := v["progress"].(float64); ok {
				status.VSDProgress = float32(progress)
			}
			if isActive, ok := v["is_active"].(bool); ok {
				status.VSDIsActive = isActive
			}
		}
	}

	if toolhead, ok := result.Status["toolhead"]; ok {
		if th, ok := toolhead.(map[string]interface{}); ok {
			if pos, ok := th["position"].([]interface{}); ok && len(pos) >= 3 {
				status.Position = make([]float64, len(pos))
				for i, p := range pos {
					if val, ok := p.(float64); ok {
						status.Position[i] = val
					}
				}
			}
			if estTime, ok := th["estimated_print_time"].(float64); ok {
				status.EstimatedPrintTime = estTime
			}
			if printTime, ok := th["print_time"].(float64); ok {
				status.PrintTime = printTime
			}
		}
	}

	if fan, ok := result.Status["fan"]; ok {
		if f, ok := fan.(map[string]interface{}); ok {
			if speed, ok := f["speed"].(float64); ok {
				status.FanSpeed = float32(speed) * 100
			}
		}
	}

	heaterResult, _ := s.QueryObjectsWithParams(ctx, []string{"heater_bed"})
	if heaterResult != nil {
		if heaterBed, ok := heaterResult.Status["heater_bed"]; ok {
			if hb, ok := heaterBed.(map[string]interface{}); ok {
				if temp, ok := hb["temperature"].(float64); ok {
					status.BedTemperature = float32(temp)
				}
				if target, ok := hb["target"].(float64); ok {
					status.BedTarget = float32(target)
				}
			}
		}
	}

	extruderResult, _ := s.QueryObjectsWithParams(ctx, []string{"extruder"})
	if extruderResult != nil {
		if extruder, ok := extruderResult.Status["extruder"]; ok {
			if ex, ok := extruder.(map[string]interface{}); ok {
				if temp, ok := ex["temperature"].(float64); ok {
					status.NozzleTemperature = float32(temp)
				}
				if target, ok := ex["target"].(float64); ok {
					status.NozzleTarget = float32(target)
				}
			}
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
