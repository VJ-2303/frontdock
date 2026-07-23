package storage

import (
	"context"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client        *minio.Client
	BucketUploads string
	BuckerSites   string
}

func New(endpoint, access, secret string, useSSL bool, uploads, sites string) (*Storage, error) {
	c, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, secret, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Storage{
		client:        c,
		BucketUploads: uploads,
		BuckerSites:   sites,
	}, nil
}

func (s *Storage) Put(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *Storage) Stat(ctx context.Context, bucket, key string) (minio.ObjectInfo, error) {
	return s.client.StatObject(ctx, bucket, key,
		minio.StatObjectOptions{})
}

func (s *Storage) Delete(ctx context.Context, bucket, key string) error {
	return s.client.RemoveObject(ctx, bucket, key,
		minio.RemoveObjectOptions{})
}

func ContentTypeFor(name string) string {
	ct := mime.TypeByExtension(filepath.Ext(name))
	if ct != "" {
		return ct
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mjs":
		return "text/javascript; charset=utf-8"
	case ".webmanifest":
		return "application/manifest+json"
	case ".map":
		return "application/json"
	case ".woff2":
		return "font/woff2"
	case ".avif":
		return "image/avif"
	}
	return "application/octet-stream"
}
