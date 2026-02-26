package main

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const maxUploadSize = 5 << 20 // 5 MB

func newMinioClient() (*minio.Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
}

// ensureBucket creates the bucket if it doesn't exist and sets a public-read policy.
func ensureBucket(ctx context.Context, client *minio.Client, bucket string) error {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("checking bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("creating bucket: %w", err)
		}
	}

	// Allow anonymous GET on every object in the bucket.
	policy := `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": ["s3:GetObject"],
    "Resource": ["arn:aws:s3:::` + bucket + `/*"]
  }]
}`
	return client.SetBucketPolicy(ctx, bucket, policy)
}

// uploadImage uploads a multipart file to MinIO under the given folder prefix.
// It returns the public URL of the uploaded object.
func (app *application) uploadImage(file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	if header.Size > maxUploadSize {
		return "", fmt.Errorf("file exceeds 5 MB limit")
	}

	// Detect content type from the first 512 bytes.
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	if !isAllowedImageType(contentType) {
		return "", fmt.Errorf("file type %q is not allowed; only JPEG, PNG, GIF and WebP images are accepted", contentType)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	objectName := fmt.Sprintf("%s/%d%s", folder, time.Now().UnixNano(), ext)

	_, err := app.minioClient.PutObject(
		context.Background(),
		app.minioBucket,
		objectName,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("uploading to MinIO: %w", err)
	}

	scheme := "http"
	if os.Getenv("MINIO_USE_SSL") == "true" {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", scheme, os.Getenv("MINIO_ENDPOINT"), app.minioBucket, objectName)
	return url, nil
}

func isAllowedImageType(ct string) bool {
	switch ct {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	}
	return false
}

const maxDocumentSize = 20 << 20 // 20 MB

// uploadDocument uploads any document file to MinIO and returns the public URL,
// detected content-type, and original filename. Allowed types: PDF, Word, Excel,
// PowerPoint, plain text, and CSV.
func (app *application) uploadDocument(file multipart.File, header *multipart.FileHeader, folder string) (url, contentType string, err error) {
	if header.Size > maxDocumentSize {
		return "", "", fmt.Errorf("file exceeds 20 MB limit")
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType = http.DetectContentType(buf[:n])
	if !isAllowedDocumentType(contentType, header.Filename) {
		return "", "", fmt.Errorf("file type %q is not allowed for documents", contentType)
	}
	if _, err = file.Seek(0, 0); err != nil {
		return "", "", err
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".bin"
	}
	objectName := fmt.Sprintf("%s/%d%s", folder, time.Now().UnixNano(), ext)

	_, err = app.minioClient.PutObject(
		context.Background(),
		app.minioBucket,
		objectName,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", "", fmt.Errorf("uploading to MinIO: %w", err)
	}

	scheme := "http"
	if os.Getenv("MINIO_USE_SSL") == "true" {
		scheme = "https"
	}
	fileURL := fmt.Sprintf("%s://%s/%s/%s", scheme, os.Getenv("MINIO_ENDPOINT"), app.minioBucket, objectName)
	return fileURL, contentType, nil
}

func isAllowedDocumentType(ct, filename string) bool {
	switch ct {
	case "application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
		"text/csv":
		return true
	}
	// Fallback: allow by extension for common office formats whose MIME may
	// be detected as application/octet-stream on some platforms.
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".csv":
		return true
	}
	return false
}
