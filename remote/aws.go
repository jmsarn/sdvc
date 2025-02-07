package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/jmsarn/sdvc/utils"
	"github.com/schollz/progressbar/v3"
)

type S3Remote struct {
	Client     *s3.Client
	Downloader *manager.Downloader
	Prefix     string
	Uploader   *manager.Uploader
}

func wrapWithProgress(reader io.Reader, size int64) io.Reader {
	bar := progressbar.NewOptions(int(size),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("Uploading..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Println("\nUpload completed!")
		}),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionThrottle(65*time.Millisecond),
	)
	r := progressbar.NewReader(reader, bar)
	return &r
}

func NewS3Remote(prefix string, storageConfig map[string]string) *S3Remote {
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
	return &S3Remote{
		Client:     client,
		Downloader: downloader,
		Prefix:     prefix,
		Uploader:   uploader,
	}
}

func (r *S3Remote) CheckLocalObject(obj FileObject) (bool, error) {
	localHash, _ := utils.FileSHA256(obj.LocalPath)
	remoteHash, err := r.GetSHA256(obj)
	if err != nil {
		return false, err
	}
	return localHash == remoteHash, nil
}

func (r *S3Remote) Download(obj FileObject) error {
	remotePath, _ := url.JoinPath(r.Prefix, obj.LocalPath)
	if !isValidS3URI(remotePath) {
		return errors.New(fmt.Sprintf("%s is not a valid S3 URI", remotePath))
	}
	remoteSHA256, err := r.GetSHA256(obj)
	if err != nil {
		return err
	}
	if obj.SHA256 != remoteSHA256 {
		return errors.New("Hash between remote object and pointer file do not match")
	}
	bucket, key := parseURI(remotePath)
	file, err := os.Create(obj.LocalPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = r.Downloader.Download(context.TODO(), file, &s3.GetObjectInput{
		Bucket:    bucket,
		Key:       key,
		VersionId: aws.String(obj.Version),
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

func (r *S3Remote) GetSHA256(obj FileObject) (string, error) {
	remotePath, _ := url.JoinPath(r.Prefix, obj.LocalPath)
	if !isValidS3URI(remotePath) {
		return "", os.ErrInvalid
	}
	bucket, key := parseURI(remotePath)
	head, err := r.Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket:    bucket,
		Key:       key,
		VersionId: aws.String(obj.Version),
	})
	if err != nil {
		return "", err
	}
	return head.Metadata["sha256"], nil
}

func (r *S3Remote) Upload(obj FileObject) (*UploadResult, error) {
	remotePath, _ := url.JoinPath(r.Prefix, obj.LocalPath)
	if !isValidS3URI(remotePath) {
		err := errors.New(fmt.Sprintf("%s is not a valid S3 URI", remotePath))
		return nil, err
	}
	bucket, key := parseURI(remotePath)
	ver, err := r.Client.GetBucketVersioning(
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
	versions, _ := r.Client.ListObjectVersions(
		context.TODO(),
		&s3.ListObjectVersionsInput{
			Bucket: bucket,
			Prefix: key,
		},
	)
	for _, v := range versions.Versions {
		v_ := v.VersionId
		head, _ := r.Client.HeadObject(
			context.TODO(),
			&s3.HeadObjectInput{
				Bucket:    bucket,
				Key:       key,
				VersionId: v_,
			},
		)
		slog.Debug(fmt.Sprintf("%+v", *head))
		if head.Metadata["sha256"] == obj.SHA256 {
			return nil, nil
		}
	}
	file, err := os.Open(obj.LocalPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	body := wrapWithProgress(file, fileInfo.Size())
	result, err := r.Uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:   bucket,
		Key:      key,
		Body:     body,
		Metadata: map[string]string{"managedBy": "SDVC", "sha256": obj.SHA256},
	})
	if err != nil {
		return nil, err
	}
	slog.Debug(fmt.Sprintf("Upload result %+v", result))
	return &UploadResult{
		ETag:    *result.ETag,
		Path:    result.Location,
		SHA256:  obj.SHA256,
		Version: *result.VersionID,
	}, nil
}

func parseURI(uri string) (*string, *string) {
	u, _ := url.Parse(uri)
	return aws.String(u.Host), aws.String(u.Path[1:])
}

// Check if the given URI matches the form s3://<bucket>/<key>
func isValidS3URI(uri string) bool {
	s3Uri := regexp.MustCompile(`s3\:\/\/[a-zA-Z0-9\-\.]+[a-zA-Z]\/\S*?$`)
	return s3Uri.Match([]byte(uri))
}
