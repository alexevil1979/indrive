// Package storage — MinIO/S3 storage client for document uploads
package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config holds MinIO connection configuration.
type Config struct {
	Endpoint        string // e.g., "localhost:9000"
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	UseSSL          bool
	PublicURL       string // Public URL prefix for accessing files (optional)
}

// Client wraps MinIO client for document storage.
type Client struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

// New creates a new MinIO storage client.
func New(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("minio endpoint required")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio connect: %w", err)
	}

	c := &Client{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}

	return c, nil
}

// EnsureBucket creates bucket if it doesn't exist.
func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		err = c.client.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}
	return nil
}

// UploadResult contains info about uploaded file.
type UploadResult struct {
	Key         string
	Size        int64
	ContentType string
	URL         string
}

// Upload uploads a file to MinIO.
func (c *Client) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*UploadResult, error) {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	info, err := c.client.PutObject(ctx, c.bucket, key, reader, size, opts)
	if err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}

	return &UploadResult{
		Key:         key,
		Size:        info.Size,
		ContentType: contentType,
		URL:         c.GetPublicURL(key),
	}, nil
}

// Download returns a reader for the file.
func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	return obj, nil
}

// Delete removes a file from storage.
func (c *Client) Delete(ctx context.Context, key string) error {
	err := c.client.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

// GetPresignedURL returns a presigned URL for temporary access.
func (c *Client) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := c.client.PresignedGetObject(ctx, c.bucket, key, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}
	return presignedURL.String(), nil
}

// GetPublicURL returns public URL for the file (if publicURL configured).
func (c *Client) GetPublicURL(key string) string {
	if c.publicURL != "" {
		return c.publicURL + "/" + c.bucket + "/" + key
	}
	// Fallback to MinIO URL format
	return fmt.Sprintf("http://%s/%s/%s", c.client.EndpointURL().Host, c.bucket, key)
}

// GenerateKey generates a storage key for driver document.
func GenerateKey(userID, docType, filename string) string {
	ext := path.Ext(filename)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("drivers/%s/%s_%d%s", userID, docType, timestamp, ext)
}

// Stat returns file info without downloading.
func (c *Client) Stat(ctx context.Context, key string) (*minio.ObjectInfo, error) {
	info, err := c.client.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &info, nil
}
