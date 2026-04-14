package moonraker

import (
	"context"
)

// ServerService provides access to server administration endpoints.
type ServerService struct {
	client *Client
}

// NewServerService creates a new server service.
func NewServerService(client *Client) *ServerService {
	return &ServerService{client: client}
}

// GetInfo fetches Moonraker server information.
func (s *ServerService) GetInfo(ctx context.Context) (*ServerInfo, error) {
	return s.client.GetServerInfo(ctx)
}
