package localfs

import (
	"context"
	"path/filepath"

	"git.timschuster.info/rls.moe/catgi/backend/common"

	"os"

	"sync"

	"io/ioutil"

	"strings"

	"git.timschuster.info/rls.moe/catgi/logger"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// LocalFSBackend offers storage into a local flatfile fs,
// abusing the nature of POSIX file systems as Key-Value DB.
type LocalFSBackend struct {
	// Root; The Root Path of the localfs backend. This will be
	// forced relative unless AbsoluteRoot is set.
	Root string `mapstructure:"root"`
	// AbosluteRoot, if set true, the root will be treated as absolute path
	// instead of forcing a relative path.
	AbsoluteRoot bool `mapstructure:"abs_root"`
	// DirMode Sets the mode used for all directories
	DirMode os.FileMode `mapstructure:"dir_mode"`
	// FileMode Sets the mode used for all files
	FileMode os.FileMode `mapstructure:"file_mode"`
	// rwlock is used to protect write access
	rwlock sync.RWMutex
}

// Name returns localfs
func (l *LocalFSBackend) Name() string { return "localfs" }

// Upload saves the given file into the root fs after splitting it up
// to avoid polluting a folder too much.
func (l *LocalFSBackend) Upload(name string, file *common.File, ctx context.Context) error {
	name = common.EscapeName(name)

	if file == nil {
		return common.ErrorSerializationFailure
	}

	log := logger.LogFromCtx(packageName+".Upload", ctx)

	if file.Flake != name {
		log.Debug("Flake mismatch, correcting flake in file")
		file.Flake = name
	}

	filePath := l.getPath(name)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return common.ErrorFileExists
	}

	l.rwlock.Lock()
	defer l.rwlock.Unlock()

	os.MkdirAll(strings.TrimSuffix(filePath, "file.msgpack"), 0700)
	f, err := os.Create(filePath)
	defer f.Close()

	if err != nil {
		os.Remove(filePath)
		return err
	}

	dat, err := msgpack.Marshal(*file)

	if err != nil {
		f.Close()
		os.Remove(filePath)
		return err
	}

	_, err = f.Write(dat)

	if err != nil {
		f.Close()
		os.Remove(filePath)
		return err
	}

	return nil
}

// Exists calls os.Stat on the localfs.
func (l *LocalFSBackend) Exists(name string, ctx context.Context) error {
	l.rwlock.RLock()
	defer l.rwlock.RUnlock()

	filePath := l.getPath(name)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return common.NewErrorFileNotExists(name, err)
	}
	return nil
}

func (l *LocalFSBackend) Get(name string, ctx context.Context) (*common.File, error) {
	l.rwlock.RLock()
	defer l.rwlock.RUnlock()

	filePath := l.getPath(name)

	var file = &common.File{}

	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = msgpack.Unmarshal(dat, file)
	if err != nil {
		return file, err
	}

	if file.Data == nil {
		file.Data = []byte{}
	}

	return file, nil
}

func (l *LocalFSBackend) Delete(name string, ctx context.Context) error {
	l.rwlock.Lock()
	defer l.rwlock.Unlock()

	return os.Remove(l.getPath(name))
}

func (l *LocalFSBackend) ListGlob(ctx context.Context, glob string) ([]*common.File, error) {
	return nil, common.ErrorNotImplemented
}

func (l *LocalFSBackend) RunGC(ctx context.Context) ([]common.File, error) {
	return nil, common.ErrorNotImplemented
}

// pingFS checks if the root exists and is writable.
func (l *LocalFSBackend) pingFS() error {
	filePath := filepath.Join(l.Root, "/ping.lock")

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer f.Close()

	n, err := f.Write([]byte{0, 0, 0, 0})

	if err != nil {
		return err
	}

	if n != 4 {
		return common.ErrorIncompleteWrite
	}

	f.Close()

	err = os.Remove(filePath)

	if err != nil {
		return err
	}

	return nil
}

func (l *LocalFSBackend) getPath(name string) string {
	name = common.EscapeName(name)

	fileName := common.FileName(name, "msgpack", 2)

	return filepath.Join(l.Root, fileName)
}
