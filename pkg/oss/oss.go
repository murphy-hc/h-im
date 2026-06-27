package oss

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Config holds Alibaba Cloud OSS connection parameters.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
}

// Client wraps an OSS bucket.
type Client struct {
	bucket     *oss.Bucket
	BucketName string
	Endpoint   string
}

// NewClient creates a new OSS client.
func NewClient(cfg Config) (*Client, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("oss: connect: %w", err)
	}
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("oss: bucket %s: %w", cfg.BucketName, err)
	}
	return &Client{bucket: bucket, BucketName: cfg.BucketName, Endpoint: cfg.Endpoint}, nil
}

// Upload uploads data to the given object key and returns the public URL.
func (c *Client) Upload(key string, data io.Reader, contentType string) (string, error) {
	opts := []oss.Option{}
	if contentType != "" {
		opts = append(opts, oss.ContentType(contentType))
	}
	if err := c.bucket.PutObject(key, data, opts...); err != nil {
		return "", fmt.Errorf("oss: upload %s: %w", key, err)
	}
	return c.URL(key), nil
}

// Delete removes an object from the bucket.
func (c *Client) Delete(key string) error {
	if err := c.bucket.DeleteObject(key); err != nil {
		return fmt.Errorf("oss: delete %s: %w", key, err)
	}
	return nil
}

// PresignedUploadURL generates a pre-signed PUT URL that allows the client
// to upload directly to OSS without going through the media service.
func (c *Client) PresignedUploadURL(key string, contentType string, expires time.Duration) (string, error) {
	opts := []oss.Option{}
	if contentType != "" {
		opts = append(opts, oss.ContentType(contentType))
	}
	return c.bucket.SignURL(key, oss.HTTPPut, int64(expires.Seconds()), opts...)
}

// HeadObject checks if an object exists and returns its metadata.
func (c *Client) HeadObject(key string) (http.Header, error) {
	return c.bucket.GetObjectMeta(key)
}

// URL returns the public HTTPS URL for a given object key.
func (c *Client) URL(key string) string {
	return fmt.Sprintf("https://%s.%s/%s", c.BucketName, c.Endpoint, key)
}
