package s3

import (
	"context"
	"time"

	msgpack "gopkg.in/vmihailenco/msgpack.v2"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const skipSize = 2
const metaFormat = "msgpack"

func (s *S3Backend) Name() string {
	return driverName
}

func (s *S3Backend) Upload(flake string, file *common.File, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Upload", ctx)
	log.Debug("Creating object '", flake, "'")
	file.CreatedAt = common.FromTime(time.Now().UTC())
	dataName := common.DataName(flake, skipSize)
	metaName := common.MetaName(flake, skipSize, metaFormat)

	oldData := file.Data
	file.Data = []byte{}

	// Write Meta then Data to avoid inconsistency

	dat, err := msgpack.Marshal(*file)
	if err != nil {
		log.Error("Error encoding file metadata", err)
		return err
	}

	if err := s.WriteBytes(metaName, dat, ctx); err != nil {
		log.Error("Error writing metadata ", err)
		return err
	}

	file.Data = oldData

	if err := s.WriteBytes(dataName, file.Data, ctx); err != nil {
		log.Error("Error writing data ", err)
		return err
	}

	return nil
}

func (s *S3Backend) Exists(flake string, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Exists", ctx)

	log.Debug("Getting metadata")

	metaName := common.MetaName(flake, skipSize, metaFormat)
	exists, _, err := s.PingFile(metaName, ctx)
	if !exists || err != nil {
		return common.NewErrorFileNotExists(flake, err)
	}

	dataName := common.DataName(flake, skipSize)
	exists, _, err = s.PingFile(dataName, ctx)
	if !exists || err != nil {
		return common.NewErrorFileNotExists(flake, err)
	}
	return nil
}

func (s *S3Backend) Get(flake string, ctx context.Context) (*common.File, error) {
	log := logger.LogFromCtx(packageName+".Get", ctx).WithField("object", flake)
	var file = &common.File{}
	dataName := common.DataName(flake, skipSize)
	metaName := common.MetaName(flake, skipSize, metaFormat)

	{
		log.Debug("Loading MetaFile")
		dat, err := s.ReadBytes(flake, ctx)
		if err != nil {
			return nil, err
		}
		log.Debug("Unmarshalling MetaFile")
		err = msgpack.Unmarshal(dat, file)
		if err != nil {
			return nil, err
		}
		log.Debug("Checking if File is expired")
		if file.DeleteAt.TTL() == 0 {
			log.Info("Attempted to get expired file, deleting...")
			err := s.DeleteKey(dataName, ctx)
			if err != nil {
				log.Error("Error while deleting data file: ", err)
				return nil, err
			}
			err = s.DeleteKey(metaName, ctx)
			if err != nil {
				log.Error("Error while deleting meta file: ", err)
				return nil, err
			}
			return nil, common.ErrorExpired
		}
	}

	{
		log.Debug("Loading Data for File")
		dat, err := s.ReadBytes(dataName, ctx)
		if err != nil {
			log.Error("Error while reading data file: ", err)
			return nil, err
		}
		file.Data = dat
	}

	return file, nil
}

func (s *S3Backend) Delete(flake string, ctx context.Context) error {
	dataName := common.DataName(flake, skipSize)
	metaName := common.MetaName(flake, skipSize, metaFormat)

	err := s.DeleteKey(dataName, ctx)
	if err != nil {
		return err
	}

	err = s.DeleteKey(metaName, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Backend) ListGlob(ctx context.Context, prefix string) ([]*common.File, error) {
	log := logger.LogFromCtx(packageName+".ListGlob", ctx)
	reqInput := &s3.ListObjectsInput{
		Bucket: aws.String(s.config.Bucket),
		Prefix: aws.String(prefix),
	}
	response, err := s.s3.ListObjects(reqInput)
	if err != nil {
		return nil, err
	}
	var retList = []*common.File{}
	for k := range response.Contents {
		dat, err := s.ReadBytes(common.MetaName(
			*response.Contents[0].Key, skipSize, metaFormat), ctx)
		if err != nil {
			log.Error("Error on Read for Key ", *response.Contents[k].Key, ":", err)
			continue
		}
		var file = &common.File{}
		err = msgpack.Unmarshal(dat, file)
		if err != nil {
			log.Error("Error on Unpack for Key ", *response.Contents[k].Key, ":", err)
			continue
		}
		retList = append(retList, file)
	}
	return retList, nil
}

func (s *S3Backend) RunGC(ctx context.Context) ([]common.File, error) {
	return common.GenericGC(s, nil, nil, ctx)
}
