package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/gif"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const maxUploadSize    = 5 << 20  // 5 MB  — target size after compression
const maxImageInputSize = 20 << 20 // 20 MB — maximum accepted before attempting compression

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

// uploadImage uploads a multipart image to MinIO under the given folder prefix.
// If the file exceeds maxUploadSize (5 MB) the image is automatically compressed
// (quality reduction then dimension halving) before uploading.
// Files larger than maxImageInputSize (20 MB) are rejected outright.
// Returns the public URL of the uploaded object.
func (app *application) uploadImage(file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	if header.Size > maxImageInputSize {
		return "", fmt.Errorf("file exceeds 20 MB — please use a smaller image")
	}

	// Detect content type from the first 512 bytes.
	buf512 := make([]byte, 512)
	n, _ := file.Read(buf512)
	contentType := http.DetectContentType(buf512[:n])
	if !isAllowedImageType(contentType) {
		return "", fmt.Errorf("file type %q is not allowed; only JPEG, PNG, GIF and WebP images are accepted", contentType)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}

	// Decide whether to compress or stream directly.
	var uploadReader io.Reader
	var uploadSize    int64
	uploadContentType := contentType
	ext := filepath.Ext(header.Filename)

	if header.Size > maxUploadSize {
		// Image is over 5 MB — try to compress it down.
		compressed, ct, err := compressImage(file, contentType)
		if err != nil {
			return "", err
		}
		uploadReader       = bytes.NewReader(compressed)
		uploadSize         = int64(len(compressed))
		uploadContentType  = ct
		ext                = ".jpg" // output is always JPEG after compression
	} else {
		uploadReader = file
		uploadSize   = header.Size
		if ext == "" {
			ext = ".jpg"
		}
	}

	objectName := fmt.Sprintf("%s/%d%s", folder, time.Now().UnixNano(), ext)

	_, err := app.minioClient.PutObject(
		context.Background(),
		app.minioBucket,
		objectName,
		uploadReader,
		uploadSize,
		minio.PutObjectOptions{ContentType: uploadContentType},
	)
	if err != nil {
		return "", fmt.Errorf("uploading to MinIO: %w", err)
	}

	scheme := "http"
	if os.Getenv("MINIO_USE_SSL") == "true" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, os.Getenv("MINIO_ENDPOINT"), app.minioBucket, objectName), nil
}

// compressImage re-encodes an image as JPEG at progressively lower quality
// and smaller dimensions until the result fits within maxUploadSize.
func compressImage(file multipart.File, contentType string) ([]byte, string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, "image/jpeg", fmt.Errorf("reading image data: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		// Likely WebP or an unsupported codec — give a clear message.
		return nil, "image/jpeg", fmt.Errorf(
			"image is over 5 MB and could not be automatically compressed " +
				"(unsupported format — please convert to JPEG or PNG first)")
	}

	// Remove transparency so JPEG encoding is lossless in terms of colour.
	img = flattenAlpha(img)

	// Phase 1: quality reduction at original dimensions.
	for _, q := range []int{82, 70, 58, 44, 30} {
		if out, err := jpegEncode(img, q); err == nil && int64(len(out)) <= maxUploadSize {
			return out, "image/jpeg", nil
		}
	}

	// Phase 2: scale to 50 % of original dimensions.
	b := img.Bounds()
	w, h := (b.Max.X-b.Min.X)/2, (b.Max.Y-b.Min.Y)/2
	half := nearestScale(img, w, h)
	for _, q := range []int{82, 70, 58} {
		if out, err := jpegEncode(half, q); err == nil && int64(len(out)) <= maxUploadSize {
			return out, "image/jpeg", nil
		}
	}

	// Phase 3: scale to 25 % of original dimensions.
	quarter := nearestScale(img, w/2, h/2)
	if out, err := jpegEncode(quarter, 80); err == nil && int64(len(out)) <= maxUploadSize {
		return out, "image/jpeg", nil
	}

	return nil, "image/jpeg", fmt.Errorf("image could not be compressed to under 5 MB — please use a smaller image")
}

// jpegEncode encodes img as JPEG at the given quality (1–100) and returns the bytes.
func jpegEncode(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	return buf.Bytes(), err
}

// flattenAlpha composites src over a white background, removing any alpha channel
// so that JPEG encoding (which has no transparency) produces correct colours.
func flattenAlpha(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	// Flood-fill white.
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	// Composite the source image on top.
	draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
	return dst
}

// nearestScale resizes src to newW×newH using nearest-neighbour sampling.
// This requires no external dependencies and is fast enough for one-time uploads.
func nearestScale(src image.Image, newW, newH int) image.Image {
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}
	bounds := src.Bounds()
	srcW := bounds.Max.X - bounds.Min.X
	srcH := bounds.Max.Y - bounds.Min.Y
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			dst.Set(x, y, src.At(bounds.Min.X+x*srcW/newW, bounds.Min.Y+y*srcH/newH))
		}
	}
	return dst
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
