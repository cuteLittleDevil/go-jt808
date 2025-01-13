package file

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"path"
	"time"
)

var _s3Client *Minio

type Minio struct {
	*minio.Client
	bucket string
}

func Init(url string, appKey, appSecret string, bucket string) error {
	client, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(appKey, appSecret, ""),
		Secure: false,
	})
	if err != nil {
		return err
	}
	ok, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return err
	}
	if !ok {
		if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{
			Region:        "cn-north-1",
			ObjectLocking: false,
		}); err != nil {
			return err
		}
	}
	_s3Client = &Minio{
		Client: client,
		bucket: bucket,
	}
	return nil
}

func Default() *Minio {
	return _s3Client
}

func (m *Minio) Upload(name string, data []byte) (string, error) {
	contentType := "binary/octet-stream"
	switch path.Ext(name) {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	}
	uploadInfo, err := m.Client.PutObject(context.Background(),
		m.bucket, name,
		bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
			ContentType: contentType,
		})
	if err != nil {
		return "", err
	}
	tmpGetURL, err := m.Client.PresignedGetObject(context.Background(), m.bucket,
		uploadInfo.Key, time.Hour, nil)
	if err != nil {
		return "", err
	}
	return tmpGetURL.String(), nil
}
