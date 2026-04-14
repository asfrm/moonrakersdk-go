package moonraker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// FileService provides access to file management endpoints.
type FileService struct {
	client *Client
}

// NewFileService creates a new file service.
func NewFileService(client *Client) *FileService {
	return &FileService{client: client}
}

// ListFiles lists available GCode files.
func (s *FileService) ListFiles(ctx context.Context, root string) (*FileListResponse, error) {
	path := "/server/files/list"
	if root != "" {
		path += "?root=" + root
	}

	var response FileListResponse
	if err := s.client.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return &response, nil
}

// ListGCodeFiles lists GCode files in the gcodes root.
func (s *FileService) ListGCodeFiles(ctx context.Context) (*FileListResponse, error) {
	return s.ListFiles(ctx, "gcodes")
}

// GetFileMetadata retrieves metadata about a specific file.
func (s *FileService) GetFileMetadata(ctx context.Context, filename string) (*FileMetadata, error) {
	path := fmt.Sprintf("/server/files/metadata?filename=%s", filename)

	var metadata FileMetadata
	if err := s.client.doRequest(ctx, "GET", path, nil, &metadata); err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	return &metadata, nil
}

// UploadOptions represents options for file upload.
type UploadOptions struct {
	Filename   string
	FilePath   string    // Local file path
	FileData   io.Reader // Or provide data directly
	StartPrint bool      // Start printing immediately
	Room       string    // Target directory
}

// Upload uploads a GCode file to Moonraker.
func (s *FileService) Upload(ctx context.Context, opts UploadOptions) (*FileMetadata, error) {
	var fileReader io.Reader
	var fileName string

	if opts.FilePath != "" {
		file, err := os.Open(opts.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer func() { _ = file.Close() }()
		fileReader = file
		fileName = filepath.Base(opts.FilePath)
	} else if opts.FileData != nil {
		fileReader = opts.FileData
		fileName = opts.Filename
	} else {
		return nil, fmt.Errorf("either FilePath or FileData must be provided")
	}

	if fileName == "" {
		return nil, fmt.Errorf("filename is required")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, fileReader); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	if opts.StartPrint {
		if err := writer.WriteField("print", "true"); err != nil {
			return nil, fmt.Errorf("failed to write print field: %w", err)
		}
	}

	if opts.Room != "" {
		if err := writer.WriteField("root", opts.Room); err != nil {
			return nil, fmt.Errorf("failed to write root field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	var metadata FileMetadata
	if err := s.client.doRequestForm(ctx, "POST", "/server/files/upload", body, writer.FormDataContentType(), &metadata); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &metadata, nil
}

// DeleteFile deletes a file from Moonraker.
func (s *FileService) DeleteFile(ctx context.Context, filename string) error {
	path := fmt.Sprintf("/server/files/%s", filename)
	var result string
	if err := s.client.doRequest(ctx, "DELETE", path, nil, &result); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
