package storage

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/mattermost/backup-restore-tool/pkg/backuprestore"
	s3 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	"github.com/pkg/errors"
)

type S3FileBackend struct {
	endpoint  string
	accessKey string
	secretKey string
	secure    bool
	region    string
	bucket    string
	isBifrost bool

	client *s3.Client
}

func NewS3FileBackend(config backuprestore.StorageConfig) (*S3FileBackend, error) {
	backend := &S3FileBackend{
		endpoint:  config.Endpoint,
		accessKey: config.AccessKey,
		secretKey: config.SecretKey,
		secure:    config.EnableTLS,
		region:    config.Region,
		bucket:    config.Bucket,
		isBifrost: config.Bifrost,
	}

	err := backend.initClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to init client")
	}

	return backend, nil
}

func (b *S3FileBackend) initClient() error {
	var creds *credentials.Credentials

	if b.isBifrost {
		creds = credentials.New(customProvider{})
	} else {
		creds = credentials.NewStaticV4(b.accessKey, b.secretKey, "")
	}

	opts := s3.Options{
		Creds:  creds,
		Secure: b.secure,
		Region: b.region,
	}

	// If using Bifrost, override the default transport.
	if b.isBifrost {
		tr, err := s3.DefaultTransport(b.secure)
		if err != nil {
			return err
		}
		scheme := "http"
		if b.secure {
			scheme = "https"
		}
		opts.Transport = &customTransport{
			base:   tr,
			host:   b.endpoint,
			scheme: scheme,
		}
	}

	s3Clnt, err := s3.New(b.endpoint, &opts)
	if err != nil {
		return err
	}

	b.client = s3Clnt

	return nil
}

// UploadFile uploads a single file to file store
func (c *S3FileBackend) UploadFile(objectName, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "failed to open file: %q", path)
	}
	defer file.Close()

	var contentType string
	contentType = "binary/octet-stream"

	options := s3PutOptions(true, contentType)
	_, err = c.client.PutObject(context.Background(), c.bucket, objectName, file, -1, options)
	if err != nil {
		return errors.Wrap(err, "failed to upload file to file storage")
	}

	return nil
}

// DownloadFile downloads file from file store
func (c *S3FileBackend) DownloadFile(objectKey, path string) error {
	minioObject, err := c.client.GetObject(context.Background(), c.bucket, objectKey, s3.GetObjectOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to init storage client")
	}
	defer minioObject.Close()

	file, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file: %q", path)
	}
	defer file.Close()

	_, err = io.Copy(file, minioObject)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	return nil
}

func s3PutOptions(encrypted bool, contentType string) s3.PutObjectOptions {
	options := s3.PutObjectOptions{}
	if encrypted {
		options.ServerSideEncryption = encrypt.NewSSE()
	}
	options.ContentType = contentType
	// We set the part size to the minimum allowed value of 5MBs
	// to avoid an excessive allocation in minio.PutObject implementation.
	options.PartSize = 1024 * 1024 * 5

	return options
}

// customTransport is used to point the request to a different server.
// This is helpful in situations where a different service is handling AWS S3 requests
// from multiple applications, and applications themselves do not have any S3 credentials.
type customTransport struct {
	base   http.RoundTripper
	host   string
	scheme string
	client http.Client
}

// RoundTrip implements the http.Roundtripper interface.
func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rountrippers should not modify the original request.
	newReq := req.Clone(context.Background())
	*newReq.URL = *req.URL
	req.URL.Scheme = t.scheme
	req.URL.Host = t.host
	return t.client.Do(req)
}

// customProvider is a dummy credentials provider for the minio client to work
// without actually providing credentials. This is needed with a custom transport
// in cases where the minio client does not actually have credentials with itself,
// rather needs responses from another entity.
//
// It satisfies the credentials.Provider interface.
type customProvider struct{}

// Retrieve returns empty credentials.
func (cp customProvider) Retrieve() (credentials.Value, error) {
	sign := credentials.SignatureV4
	return credentials.Value{
		SignerType: sign,
	}, nil
}

// IsExpired always returns false.
func (cp customProvider) IsExpired() bool { return false }
