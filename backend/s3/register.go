package s3

import (
	"context"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Backend struct {
	config *s3Config
	sess   *session.Session
	s3     *s3.S3
}

type s3Config struct {
	Region    string `cgc:"region"`
	AccessKey string `cgc:"access_key"`
	SecretKey string `cgc:"secret_key"`
	Bucket    string `cgc:"bucket"`
	Prefix    string `cgc:"prefix"`
}

const packageName = "s3_backend"
const driverName = "s3"

const backendOptions = 0 |
	common.BackendOptionDirectBytesIO |
	common.BackendOptionDirectReaderIO |
	common.BackendOptionPingFile

func init() {
	backend.NewDriver(driverName, NewS3Backend)
}

func NewS3Backend(params map[string]interface{}, ctx context.Context) (common.Backend, error) {
	log := logger.LogFromCtx(packageName+".NewS3Backend", ctx)

	log.Debug("Loading Configuration")
	var config = &s3Config{}
	err := common.DecodeConfig(config, params, ctx,
		common.ConfigDefault("prefix", "/"),
		common.ConfigMustHave(
			"region",
			"access_key", "secret_key",
			"bucket",
		),
	)
	if err != nil {
		return nil, err
	}

	log.Debug("Connecting to AWS")
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(
				config.AccessKey,
				config.SecretKey,
				"",
			),
			Region: aws.String(config.Region),
		},
	})
	if err != nil {

		return nil, err
	}
	svc := s3.New(sess)

	s3b := S3Backend{
		config: config,
		sess:   sess,
		s3:     svc,
	}

	log.Debug("S3B: ", s3b)

	return nil, common.ErrorNotImplemented
}
