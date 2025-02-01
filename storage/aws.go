package storage

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Storage struct {
	Client     *s3.Client
	Downloader *manager.Downloader
	Prefix     string
	Uploader   *manager.Uploader
}

func NewS3Storage(prefix string, storageConfig map[string]string) *S3Storage {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	if profile, ok := storageConfig["profile"]; ok {
		slog.Debug(fmt.Sprintf("Using profile %s", profile))
		newCfg, err := config.LoadDefaultConfig(
			context.TODO(),
			config.WithSharedConfigProfile(profile),
		)
		if err == nil {
			cfg = newCfg
		}
	}
	var client *s3.Client
	if endpoint, ok := storageConfig["endpointurl"]; ok {
		useSSL, _ := strconv.ParseBool(storageConfig["use_ssl"])
		slog.Debug(fmt.Sprintf("Base endpoint: %s", endpoint))
		// cfg.BaseEndpoint = aws.String(endpoint)
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
			o.EndpointOptions = s3.EndpointResolverOptions{DisableHTTPS: !useSSL}
		})
	} else {
		client = s3.NewFromConfig(cfg)
	}
	downloader := manager.NewDownloader(client)
	uploader := manager.NewUploader(client)
	return &S3Storage{
		Client:     client,
		Downloader: downloader,
		Prefix:     prefix,
		Uploader:   uploader,
	}
}

func (s *S3Storage) CheckLocalObject(obj StorageObject) (bool, error) {
	file, err := os.Open(obj.LocalPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, err
	}
	localHash := hex.EncodeToString(hasher.Sum(nil))
	remoteHash, err := s.GetSHA256(obj)
	if err != nil {
		return false, err
	}
	return localHash == remoteHash, nil
}

func (s *S3Storage) Download(obj StorageObject) error {
	if !isValidURI(obj.RemotePath) {
		return os.ErrInvalid
	}
	bucket, key := parseURI(obj.RemotePath)
	file, err := os.Create(obj.LocalPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = s.Downloader.Download(context.TODO(), file, &s3.GetObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			err = noKey
		}
		return err
	}
	return nil
}

func (s *S3Storage) GetSHA256(obj StorageObject) (string, error) {
	if !isValidURI(obj.RemotePath) {
		return "", os.ErrInvalid
	}
	bucket, key := parseURI(obj.RemotePath)
	head, err := s.Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket:    bucket,
		Key:       key,
		VersionId: aws.String(obj.Version),
	})
	if err != nil {
		return "", err
	}
	return *head.ChecksumSHA256, nil

}

func (s *S3Storage) Upload(obj StorageObject) (*UploadResult, error) {
	if !isValidURI(obj.RemotePath) {
		err := errors.New(fmt.Sprintf("%s is not a valid S3 URI", obj.RemotePath))
		return nil, err
	}
	bucket, key := parseURI(obj.RemotePath)
	ver, err := s.Client.GetBucketVersioning(
		context.TODO(),
		&s3.GetBucketVersioningInput{Bucket: bucket},
	)
	if err != nil {
		return nil, err
	}
	// Don't upload file to a bucket that doesn't have versioning enabled
	if ver.Status != "Enabled" {
		slog.Debug(fmt.Sprintf("Bucket versioning: %+v", ver))
		err = errors.New(fmt.Sprintf(
			"Bucket %s does not have versioning enabled, to prevent data loss enable versioning on the bucket",
			*bucket,
		))
		return nil, err
	}
	// Check if current file is already uploaded
	versions, _ := s.Client.ListObjectVersions(
		context.TODO(),
		&s3.ListObjectVersionsInput{
			Bucket: bucket,
			Prefix: key,
		},
	)
	for _, v := range versions.Versions {
		v_ := v.VersionId
		head, _ := s.Client.HeadObject(
			context.TODO(),
			&s3.HeadObjectInput{
				Bucket:       bucket,
				Key:          key,
				VersionId:    v_,
				ChecksumMode: types.ChecksumModeEnabled,
			},
		)
		if *head.ChecksumSHA256 == hexToBase64(obj.SHA256) {
			return nil, nil
		}
	}
	file, err := os.Open(obj.LocalPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result, err := s.Uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:            bucket,
		Key:               key,
		Body:              file,
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
		ChecksumSHA256:    aws.String(hexToBase64(obj.SHA256)),
		Metadata:          map[string]string{"managedBy": "SDVC"},
	})
	if err != nil {
		return nil, err
	}
	slog.Debug(fmt.Sprintf("Upload result %+v", result))
	r := &UploadResult{
		ETag:    *result.ETag,
		Path:    result.Location,
		SHA256:  *result.ChecksumSHA256,
		Version: *result.VersionID,
	}
	return r, nil
}

func parseURI(uri string) (*string, *string) {
	u, _ := url.Parse(uri)
	return aws.String(u.Host), aws.String(u.Path[1:])
}

// Check if the given URI matches the form s3://<bucket>/<key>
func isValidURI(uri string) bool {
	s3Uri := regexp.MustCompile(`s3\:\/\/[a-zA-Z0-9\-\.]+[a-zA-Z]\/\S*?$`)
	return s3Uri.Match([]byte(uri))
}

func hexToBase64(objHash string) string {
	raw, _ := hex.DecodeString(objHash)
	encoded := base64.StdEncoding.EncodeToString(raw)
	return encoded
}
