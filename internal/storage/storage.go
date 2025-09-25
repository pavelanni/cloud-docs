package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type Client struct {
	client     *storage.Client
	bucketName string
}

type FileInfo struct {
	Content     io.ReadCloser
	ContentType string
	Size        int64
}

func NewClient(ctx context.Context, bucketName string) (*Client, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}

	var client *storage.Client
	var err error

	// Try to use the same credentials as gcloud CLI
	// First, check if GOOGLE_APPLICATION_CREDENTIALS is set
	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credPath))
	} else {
		// Use application default credentials, which should pick up gcloud credentials
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &Client{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) GetFile(ctx context.Context, objectPath string) (*FileInfo, error) {
	objectPath = strings.TrimPrefix(objectPath, "/")
	
	obj := c.client.Bucket(c.bucketName).Object(objectPath)
	
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist || strings.Contains(err.Error(), "storage: object doesn't exist") {
			return nil, fmt.Errorf("file not found: %s", objectPath)
		}
		return nil, fmt.Errorf("failed to get object attributes: %w", err)
	}

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create object reader: %w", err)
	}

	contentType := attrs.ContentType
	if contentType == "" {
		contentType = detectContentType(objectPath)
	}

	return &FileInfo{
		Content:     reader,
		ContentType: contentType,
		Size:        attrs.Size,
	}, nil
}

func detectContentType(filename string) string {
	ext := filepath.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	
	if contentType == "" {
		switch strings.ToLower(ext) {
		case ".html", ".htm":
			return "text/html; charset=utf-8"
		case ".css":
			return "text/css; charset=utf-8"
		case ".js":
			return "application/javascript; charset=utf-8"
		case ".json":
			return "application/json; charset=utf-8"
		case ".xml":
			return "application/xml; charset=utf-8"
		case ".png":
			return "image/png"
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".gif":
			return "image/gif"
		case ".svg":
			return "image/svg+xml"
		case ".pdf":
			return "application/pdf"
		default:
			return "application/octet-stream"
		}
	}
	
	return contentType
}

func (c *Client) UploadFile(ctx context.Context, objectPath string, content io.Reader, contentType string) error {
	objectPath = strings.TrimPrefix(objectPath, "/")
	
	obj := c.client.Bucket(c.bucketName).Object(objectPath)
	writer := obj.NewWriter(ctx)
	
	if contentType == "" {
		contentType = detectContentType(objectPath)
	}
	writer.ContentType = contentType
	
	if _, err := io.Copy(writer, content); err != nil {
		writer.Close()
		return fmt.Errorf("failed to upload file %s: %w", objectPath, err)
	}
	
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer for %s: %w", objectPath, err)
	}
	
	return nil
}