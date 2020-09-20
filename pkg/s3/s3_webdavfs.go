// Package s3 provides an S3 backend implementation of the webdav Filesystem interface
package s3

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/nicolaspernoud/vestibule/pkg/log"

	"golang.org/x/net/webdav"
)

// WdFs provides a webdav FileSystem interface implementation
type WdFs struct {
	*Fs
}

// NewWdFs returns a new S3 backed webdav filesystem
func NewWdFs(endpoint string, region string, bucket string, accessKeyID string, secretAccessKey string) WdFs {
	// Configure to use S3 compatible
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(s3Config)
	if err != nil {
		log.Logger.Printf("Error: %v\n", err)
	}

	// Initialize the file system
	s3Fs := NewFs(bucket, sess)
	return WdFs{s3Fs}
}

// Mkdir implements the wedav filesystem interface for a s3 bucket
func (sw WdFs) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return sw.Fs.Mkdir(name, perm)
}

// OpenFile implements the wedav filesystem interface for a s3 bucket
func (sw WdFs) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return sw.Fs.OpenFile(name, flag, perm)
}

// RemoveAll implements the wedav filesystem interface for a s3 bucket
func (sw WdFs) RemoveAll(ctx context.Context, name string) error {
	return sw.Fs.RemoveAll(name)
}

// Rename implements the wedav filesystem interface for a s3 bucket
func (sw WdFs) Rename(ctx context.Context, oldName, newName string) error {
	return sw.Fs.Rename(oldName, newName)
}

// Stat implements the wedav filesystem interface for a s3 bucket
func (sw WdFs) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return sw.Fs.Stat(name)
}
