package moonraker

import (
	"context"
	"time"
)

// Moonraker is the main facade that provides access to all services.
type Moonraker struct {
	Client   *Client
	Printer  *PrinterService
	Files    *FileService
	Commands *CommandService
	Server   *ServerService
}

// New creates a new Moonraker facade.
func New(baseURL string, opts ...ClientOption) *Moonraker {
	client := NewClient(baseURL, opts...)

	return &Moonraker{
		Client:   client,
		Printer:  NewPrinterService(client),
		Files:    NewFileService(client),
		Commands: NewCommandService(client),
		Server:   NewServerService(client),
	}
}

// Connect tests connection to Moonraker.
func (m *Moonraker) Connect(ctx context.Context) error {
	_, err := m.Server.GetInfo(ctx)
	return err
}

// WaitForKlippy waits for Klipper to be ready with timeout.
func (m *Moonraker) WaitForKlippy(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		info, err := m.Server.GetInfo(ctx)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		switch info.KlippyState {
		case KlippyStateReady:
			return nil
		case KlippyStateError, KlippyStateShutdown, KlippyStateDisconnected:
			return &APIError{
				StatusCode: 500,
				Message:    "klipper in state: " + info.KlippyState,
			}
		case KlippyStateStartup:
			time.Sleep(2 * time.Second)
			continue
		default:
			time.Sleep(2 * time.Second)
		}
	}

	return &APIError{
		StatusCode: 504,
		Message:    "timeout waiting for klipper to be ready",
	}
}

// GetPrinterStatus returns comprehensive printer status.
func (m *Moonraker) GetPrinterStatus(ctx context.Context) (*PrinterStatus, error) {
	return m.Printer.GetStatus(ctx)
}

// ExecuteCommand executes a print command.
func (m *Moonraker) ExecuteCommand(ctx context.Context, cmd PrintCommand) error {
	return m.Commands.ExecuteCommand(ctx, cmd)
}

// ListFiles lists GCode files.
func (m *Moonraker) ListFiles(ctx context.Context) (*FileListResponse, error) {
	return m.Files.ListGCodeFiles(ctx)
}

// GetFileMetadata gets metadata for a file.
func (m *Moonraker) GetFileMetadata(ctx context.Context, filename string) (*FileMetadata, error) {
	return m.Files.GetFileMetadata(ctx, filename)
}

// UploadAndPrint uploads and starts printing.
func (m *Moonraker) UploadAndPrint(ctx context.Context, opts UploadOptions) error {
	opts.StartPrint = true
	_, err := m.Files.Upload(ctx, opts)
	return err
}

// RunGCode executes GCode.
func (m *Moonraker) RunGCode(ctx context.Context, script string) error {
	return m.Commands.RunGCode(ctx, script)
}

// EmergencyStop immediately halts the printer.
func (m *Moonraker) EmergencyStop(ctx context.Context) error {
	return m.Commands.EmergencyStop(ctx)
}
