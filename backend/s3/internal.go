package s3

import (
	"bytes"
	"context"
	"io"

	"io/ioutil"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (s *S3Backend) DeleteKey(name string, ctx context.Context) error {
	delRequest := &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(name),
	}
	_, err := s.s3.DeleteObject(delRequest)
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Backend) WriteBytes(name string, data []byte, ctx context.Context) error {
	return s.WriteReader(name, bytes.NewReader(data), ctx)
}

func (s *S3Backend) WriteReader(name string, data io.ReadSeeker, ctx context.Context) error {
	putRequest := &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Body:   data,
		Key:    aws.String(name),
	}
	_, err := s.s3.PutObject(putRequest)
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Backend) ReadBytes(name string, ctx context.Context) ([]byte, error) {
	reader, err := s.ReadReader(name, ctx)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(reader)
	if err == io.EOF || err == nil {
		return data, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *S3Backend) ReadReader(name string, ctx context.Context) (io.ReadCloser, error) {
	getRequest := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(name),
	}
	getResponse, err := s.s3.GetObject(getRequest)
	if err != nil {
		return nil, err
	}
	return getResponse.Body, nil
}

func (s *S3Backend) PingFile(name string, ctx context.Context) (bool, interface{}, error) {
	getACLRequest := &s3.GetObjectAclInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(name),
	}
	getACLResponse, err := s.s3.GetObjectAcl(getACLRequest)
	if err != nil {
		return false, getACLResponse, err
	}
	return true, getACLResponse, nil
}

func (s *S3Backend) GetOptions() common.BackendOption {
	return backendOptions
}

func (s *S3Backend) GetFirstWith(options common.BackendOption) common.Backend {
	if options&backendOptions != options {
		return s
	}
	return nil
}

func (s *S3Backend) GetAllWith(options common.BackendOption) []common.Backend {
	return []common.Backend{s.GetFirstWith(options)}
}
