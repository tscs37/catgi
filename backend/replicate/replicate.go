package replicate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"

	msgpack "gopkg.in/vmihailenco/msgpack.v2"

	"golang.org/x/crypto/blake2b"

	"github.com/labstack/gommon/log"

	"git.timschuster.info/rls.moe/catgi/backend"
	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
)

type Replication struct {
	DriverSet []struct {
		Driver       string                 `cgc:"driver"`
		DriverConfig map[string]interface{} `cgc:"params"`
	} `cgc:"backends"`

	drivers []common.Backend
	// MinWrites dictates how many writes must complete
	// for a successful write.
	// If set to -1 then all writes must complete.
	MinWrites int `cgc:"min_writes"`
	// ReadMode specifies how the backend should treat
	// reads. By default it uses "random"
	ReadMode string `cgc:"read_mode"`
}

const driverName = "replicate"
const packageName = "backend/replicate"

func init() {
	//backend.NewDriver(driverName, nil)
}

func NewReplicateBackend(params map[string]interface{}, ctx context.Context) (common.Backend, error) {
	var repl = &Replication{}
	{
		err := common.DecodeConfig(repl, params, ctx,
			common.ConfigMustHave("drivers"),
			common.ConfigDefault("min_writes", -1),
			common.ConfigDefault("read_mode", "random"),
		)
		if err != nil {
			return nil, err
		}
	}

	for k := range repl.DriverSet {
		node, err := backend.NewBackend(
			repl.DriverSet[k].Driver,
			repl.DriverSet[k].DriverConfig,
			ctx)
		if err != nil {
			return nil, err
		}
		repl.drivers = append(repl.drivers, node)
	}

	return repl, nil
}

const (
	unknownReadMode int = -(iota + 1)
	readAll
	readAllChecksum
)

func (r *Replication) selectBackend() int {
	if r.ReadMode == "random" {
		return rand.Int() % len(r.drivers)
	} else if r.ReadMode == "read-all" {
		return readAll
	} else if r.ReadMode == "read-all-checksum" {
		return readAllChecksum
	}
	return unknownReadMode
}

func mapError(errorMap map[string]error) error {
	if len(errorMap) == 0 {
		return nil
	}
	return fmt.Errorf("%d errors: %#+v", len(errorMap), errorMap)
}

func mergeFileMap(fileMap map[string]*common.File, checksum bool) (*common.File, error) {
	if fileMap == nil {
		return nil, nil
	}
	if len(fileMap) == 0 {
		return nil, nil
	}
	if len(fileMap) == 1 {
		for k := range fileMap {
			return fileMap[k], nil
		}
	}

	var hashSum = &[64]byte{}
	var selectedFile *common.File
	var errorMap = map[string]error{}
	for k := range fileMap {
		encoded, err := msgpack.Marshal(fileMap[k])
		if err != nil {
			errorMap[k] = err
		}
		nextHashSum := blake2b.Sum512(encoded)
		if hashSum == nil {
			hashSum = &nextHashSum
			selectedFile = fileMap[k]
		}
		if !bytes.Equal((*hashSum)[:], nextHashSum[:]) {
			errorMap[k] = errors.New("Replication Hash Mismatch")
		}
	}
	return selectedFile, mapError(errorMap)
}

func (r *Replication) Name() string { return driverName }

func (r *Replication) Upload(flake string, file *common.File, ctx context.Context) error {
	if file == nil {
		return common.ErrorSerializationFailure
	}
	if file.Flake != flake {
		file.Flake = flake
	}

	var remainingWrites = r.MinWrites
	if remainingWrites < 0 {
		remainingWrites = len(r.DriverSet)
	}

	var errorMap = map[string]error{}
	for k := range r.drivers {
		err := r.drivers[k].Upload(flake, file, ctx)
		if err != nil {
			log.Errorf("Error on Backend Nr %d[%s]: %s", k, r.drivers[k].Name(), err)
			errorMap[r.drivers[k].Name()] = err
		} else {
			remainingWrites--
		}
	}

	if remainingWrites > 0 {
		return mapError(errorMap)
	}
	return nil
}

func (r *Replication) Exists(flake string, ctx context.Context) error {
	log := logger.LogFromCtx(packageName+".Exists", ctx)

	n := r.selectBackend()

	if n >= 0 {
		return r.drivers[n].Exists(flake, ctx)
	} else if n == readAll || n == readAllChecksum {
		var existErrors = map[string]error{}
		for k := range r.drivers {
			err := r.drivers[k].Exists(flake, ctx)
			if err != nil {
				log.Errorf("Error on Backend Nr %d[%s]: %s", k, r.drivers[k].Name(), err)
				existErrors[r.drivers[k].Name()] = err
			}
		}
		return mapError(existErrors)
	}

	return common.ErrorNotImplemented
}

func (r *Replication) Get(flake string, ctx context.Context) (*common.File, error) {
	log := logger.LogFromCtx(packageName+".Get", ctx)

	n := r.selectBackend()

	if n >= 0 {
		return r.drivers[n].Get(flake, ctx)
	} else if n == readAll || n == readAllChecksum {
		var files = map[string]*common.File{}
		var errors = map[string]error{}
		for k := range r.drivers {
			file, err := r.drivers[k].Get(flake, ctx)
			if err != nil {
				errors[r.drivers[k].Name()] = err
			} else {
				files[r.drivers[k].Name()] = file
			}
		}
		file, err := mergeFileMap(files, n == readAllChecksum)
		if err != nil {
			return nil, err
		}
		return file, mapError(errors)
	}
	return nil, common.ErrorNotImplemented
}

func (r *Replication) Delete(flake string, ctx context.Context) error {
	return common.ErrorNotImplemented
}

func (r *Replication) ListGlob(ctx context.Context, prefix string) ([]*common.File, error) {
	return nil, common.ErrorNotImplemented
}

func (r *Replication) RunGc(ctx context.Context) ([]common.File, error) {
	return nil, common.ErrorNotImplemented
}
