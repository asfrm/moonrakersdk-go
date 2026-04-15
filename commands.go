package moonraker

import (
	"context"
	"fmt"
)

// CommandService provides access to print control commands.
type CommandService struct {
	client *Client
}

// NewCommandService creates a new command service.
func NewCommandService(client *Client) *CommandService {
	return &CommandService{client: client}
}

// StartPrint starts printing a file.
func (s *CommandService) StartPrint(ctx context.Context, filename string) error {
	body := map[string]string{"filename": filename}
	var result string
	if err := s.client.doRequest(ctx, "POST", "/printer/print/start", body, &result); err != nil {
		return fmt.Errorf("failed to start print: %w", err)
	}
	return nil
}

// PausePrint pauses the current print.
func (s *CommandService) PausePrint(ctx context.Context) error {
	var result string
	if err := s.client.doRequest(ctx, "POST", "/printer/print/pause", nil, &result); err != nil {
		return fmt.Errorf("failed to pause print: %w", err)
	}
	return nil
}

// ResumePrint resumes a paused print.
func (s *CommandService) ResumePrint(ctx context.Context) error {
	var result string
	if err := s.client.doRequest(ctx, "POST", "/printer/print/resume", nil, &result); err != nil {
		return fmt.Errorf("failed to resume print: %w", err)
	}
	return nil
}

// CancelPrint cancels the current print.
func (s *CommandService) CancelPrint(ctx context.Context) error {
	var result string
	if err := s.client.doRequest(ctx, "POST", "/printer/print/cancel", nil, &result); err != nil {
		return fmt.Errorf("failed to cancel print: %w", err)
	}
	return nil
}

// CommandType represents the type of print command.
type CommandType string

const (
	CommandPause   CommandType = "pause"
	CommandResume  CommandType = "resume"
	CommandStop    CommandType = "stop"
	CommandEStop   CommandType = "estop"
	CommandHome    CommandType = "home"
	CommandSetFan  CommandType = "set_fan"
	CommandSetTemp CommandType = "set_temp"
	CommandGCode   CommandType = "gcode"
)

// PrintCommand represents a print control command.
type PrintCommand struct {
	Type   CommandType
	Params map[string]interface{}
}

// ExecuteCommand executes a print command.
func (s *CommandService) ExecuteCommand(ctx context.Context, cmd PrintCommand) error {
	switch cmd.Type {
	case CommandPause:
		return s.PausePrint(ctx)
	case CommandResume:
		return s.ResumePrint(ctx)
	case CommandStop:
		return s.CancelPrint(ctx)
	case CommandEStop:
		return s.emergencyStop(ctx)
	case CommandHome:
		return s.Home(ctx)
	case CommandSetFan:
		speed, _ := cmd.Params["speed"].(float64)
		return s.SetFanSpeed(ctx, speed)
	case CommandSetTemp:
		target, _ := cmd.Params["target"].(float64)
		heater, _ := cmd.Params["heater"].(string)
		if heater == "" {
			heater = "extruder"
		}
		return s.SetTemperature(ctx, heater, target)
	case CommandGCode:
		script, _ := cmd.Params["script"].(string)
		if script == "" {
			return fmt.Errorf("gcode script is required")
		}
		return s.runGCode(ctx, script)
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}

// SetFanSpeed sets the fan speed (0-100).
func (s *CommandService) SetFanSpeed(ctx context.Context, speed float64) error {
	speedInt := int(speed * 255 / 100)
	if speedInt < 0 {
		speedInt = 0
	}
	if speedInt > 255 {
		speedInt = 255
	}

	script := fmt.Sprintf("M106 S%d", speedInt)
	return s.runGCode(ctx, script)
}

// SetTemperature sets a heater target temperature.
func (s *CommandService) SetTemperature(ctx context.Context, heater string, temp float64) error {
	var script string
	switch heater {
	case "extruder", "nozzle":
		script = fmt.Sprintf("M104 S%.0f", temp)
	case "heater_bed", "bed":
		script = fmt.Sprintf("M140 S%.0f", temp)
	default:
		script = fmt.Sprintf("SET_HEATER_TEMPERATURE HEATER=%s TARGET=%.0f", heater, temp)
	}
	return s.runGCode(ctx, script)
}

// runGCode executes a GCode script (internal helper).
func (s *CommandService) runGCode(ctx context.Context, script string) error {
	body := map[string]string{"script": script}
	var result string
	if err := s.client.doRequest(ctx, "POST", "/printer/gcode/script", body, &result); err != nil {
		return fmt.Errorf("failed to run gcode: %w", err)
	}
	return nil
}

// RunGCode executes a GCode script (public for facade).
func (s *CommandService) RunGCode(ctx context.Context, script string) error {
	return s.runGCode(ctx, script)
}

// emergencyStop immediately halts the printer (internal).
func (s *CommandService) emergencyStop(ctx context.Context) error {
	var result string
	if err := s.client.doRequest(ctx, "POST", "/printer/emergency_stop", nil, &result); err != nil {
		return fmt.Errorf("failed to emergency stop: %w", err)
	}
	return nil
}

// EmergencyStop immediately halts the printer (public for facade).
func (s *CommandService) EmergencyStop(ctx context.Context) error {
	return s.emergencyStop(ctx)
}

// Home sends a G28 home command to the printer.
func (s *CommandService) Home(ctx context.Context) error {
	return s.runGCode(ctx, "G28")
}
