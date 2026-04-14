// Package moonraker provides a comprehensive client for the Moonraker API
package moonraker

// ServerInfo represents Moonraker server information.
type ServerInfo struct {
	KlippyConnected  bool     `json:"klippy_connected"`
	KlippyState      string   `json:"klippy_state"`
	Components       []string `json:"components"`
	FailedComponents []string `json:"failed_components"`
	RegisteredDirs   []string `json:"registered_directories"`
	Warnings         []string `json:"warnings"`
	WebsocketCount   int      `json:"websocket_count"`
	MoonrakerVersion string   `json:"moonraker_version"`
	APIVersion       []int    `json:"api_version"`
	APIVersionString string   `json:"api_version_string"`
}

// PrinterInfo represents Klipper printer information.
type PrinterInfo struct {
	State           string `json:"state"`
	StateMessage    string `json:"state_message"`
	HostName        string `json:"hostname"`
	KlipperPath     string `json:"klipper_path"`
	PythonPath      string `json:"python_path"`
	ProcessID       int    `json:"process_id"`
	UserID          int    `json:"user_id"`
	GroupID         int    `json:"group_id"`
	LogFile         string `json:"log_file"`
	ConfigFile      string `json:"config_file"`
	SoftwareVersion string `json:"software_version"`
	CPUInfo         string `json:"cpu_info"`
}

// PrinterObjectsQuery represents the result of querying printer objects.
type PrinterObjectsQuery struct {
	EventTime float64                `json:"eventtime"`
	Status    map[string]interface{} `json:"status"`
}

// FileMetadata represents extracted file metadata.
type FileMetadata struct {
	Filename      string          `json:"filename"`
	Size          int64           `json:"size"`
	Modified      float64         `json:"modified"`
	Slicer        string          `json:"slicer"`
	EstimatedTime float64         `json:"estimated_time"`
	FilamentTotal float64         `json:"filament_total"`
	LayerCount    int             `json:"layer_count"`
	LayerHeight   float32         `json:"layer_height"`
	ObjectHeight  float32         `json:"object_height"`
	ThumbnailPath string          `json:"thumbnail_path"`
	Thumbnails    []FileThumbnail `json:"thumbnails"`
}

// FileThumbnail represents a file thumbnail.
type FileThumbnail struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Size         int    `json:"size"`
	RelativePath string `json:"relative_path"`
}

// FileInfo represents basic file information.
type FileInfo struct {
	Path     string  `json:"path"`
	Filename string  `json:"filename"`
	Size     int64   `json:"size"`
	Modified float64 `json:"modified"`
}

// FileListResponse represents the response from file list endpoint.
type FileListResponse struct {
	Files []FileInfo `json:"files"`
}

// TemperatureStoreEntry represents a temperature sensor history.
type TemperatureStoreEntry struct {
	Temperatures []float32 `json:"temperatures"`
	Targets      []float32 `json:"targets,omitempty"`
	Powers       []float32 `json:"powers,omitempty"`
	Speeds       []float32 `json:"speeds,omitempty"`
}

// GCodeStoreEntry represents a gcode log entry.
type GCodeStoreEntry struct {
	Message string  `json:"message"`
	Time    float64 `json:"time"`
	Type    string  `json:"type"`
}

// EndstopStatus represents endstop status.
type EndstopStatus map[string]string

// KlippyState constants.
const (
	KlippyStateReady        = "ready"
	KlippyStateStartup      = "startup"
	KlippyStateError        = "error"
	KlippyStateShutdown     = "shutdown"
	KlippyStateDisconnected = "disconnected"
)

// PrintState constants.
const (
	PrintStateStandby  = "standby"
	PrintStatePrinting = "printing"
	PrintStatePaused   = "paused"
	PrintStateError    = "error"
	PrintStateComplete = "complete"
)
