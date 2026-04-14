package moonraker

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://192.168.1.100")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.baseURL != "http://192.168.1.100" {
		t.Errorf("expected baseURL=http://192.168.1.100, got %s", c.baseURL)
	}
	if c.httpClient.Timeout.String() != "30s" {
		t.Errorf("expected 30s timeout, got %s", c.httpClient.Timeout)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	c := NewClient("http://192.168.1.100",
		WithAPIKey("test-key"),
		WithTimeout(60*time.Second),
	)
	if c.apiKey != "test-key" {
		t.Errorf("expected apiKey=test-key, got %s", c.apiKey)
	}
	if c.httpClient.Timeout.String() != "1m0s" {
		t.Errorf("expected 1m0s timeout, got %s", c.httpClient.Timeout)
	}
}

func TestNewMoonraker(t *testing.T) {
	m := New("http://192.168.1.100")
	if m == nil {
		t.Fatal("expected non-nil Moonraker")
	}
	if m.Client == nil {
		t.Error("expected non-nil Client")
	}
	if m.Printer == nil {
		t.Error("expected non-nil Printer service")
	}
	if m.Files == nil {
		t.Error("expected non-nil Files service")
	}
	if m.Commands == nil {
		t.Error("expected non-nil Commands service")
	}
	if m.Server == nil {
		t.Error("expected non-nil Server service")
	}
}

func TestEncodeQueryParams(t *testing.T) {
	params := QueryParams{
		"name":   "test",
		"count":  5,
		"active": true,
	}
	encoded := EncodeQueryParams(params)
	if encoded == "" {
		t.Error("expected non-empty encoded params")
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "not found"}
	expected := "moonraker API error: HTTP 404 - not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestConstants(t *testing.T) {
	if KlippyStateReady != "ready" {
		t.Errorf("KlippyStateReady = %q, want %q", KlippyStateReady, "ready")
	}
	if PrintStatePrinting != "printing" {
		t.Errorf("PrintStatePrinting = %q, want %q", PrintStatePrinting, "printing")
	}
	if CommandPause != "pause" {
		t.Errorf("CommandPause = %q, want %q", CommandPause, "pause")
	}
}
