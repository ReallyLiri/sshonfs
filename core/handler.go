package core

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/dgraph-io/badger"
	"github.com/go-git/go-billy/v5"
	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/reallyliri/sshonfs/cmd"
	"github.com/willscott/go-nfs"
	"github.com/willscott/go-nfs/helpers"
	"golang.org/x/crypto/ssh"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const rootPathPlaceholder = "/"

type nfsHandler struct {
	nfs.Handler
	db *badger.DB // stores both ways: path <---> fileHandleId , both as []byte
	fs billy.Filesystem
}

var _ nfs.Handler = &nfsHandler{}

func newHandler(config *cmd.Config) (*nfsHandler, error) {
	sshConf, err := sshConfig(config)
	if err != nil {
		return nil, err
	}

	sshAddress := config.SshAddress
	if !strings.Contains(sshAddress, ":") {
		sshAddress = sshAddress + ":22"
	}
	conn, err := ssh.Dial("tcp", sshAddress, sshConf)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ssh dial to '%v'", sshAddress)
	}

	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sftp client")
	}

	homeDirPath, err := homedir.Dir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get home dir path")
	}

	badgerPath := filepath.Join(homeDirPath, ".sshonfs", hash(sshAddress), hash(config.SshRootPath))
	os.MkdirAll(badgerPath, 0777)
	opts := badger.DefaultOptions(badgerPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	bfs := &sshFs{
		sftpClient: sftpClient,
		sshRoot:    config.SshRootPath,
	}

	return &nfsHandler{
		Handler: helpers.NewNullAuthHandler(bfs),
		db:      db,
		fs:      bfs,
	}, nil
}

func hash(str string) string {
	h := sha256.Sum256([]byte(str))
	return hex.EncodeToString(h[:8])
}

func newFileHandle() ([]byte, error) {
	return uuid.New().MarshalBinary()
}

func (handler *nfsHandler) lookup(key []byte, addIfMissing bool, valueCreator func() ([]byte, error)) (value []byte, err error) {
	err = handler.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			switch err {
			case badger.ErrKeyNotFound:
				item = nil
			default:
				return err
			}
		}

		if item != nil {
			return item.Value(func(storedValue []byte) error {
				value = make([]byte, len(storedValue))
				copy(value, storedValue)
				return nil
			})
		} else if !addIfMissing {
			return nil
		}

		value, err = valueCreator()
		if err != nil {
			return err
		}

		err = txn.Set(key, value)
		if err != nil {
			return err
		}

		err = txn.Set(value, key)
		return err
	})
	return
}

func (handler *nfsHandler) ToHandle(_ billy.Filesystem, path []string) []byte {
	fullPath := filepath.Join(path...)
	if len(fullPath) == 0 {
		fullPath = rootPathPlaceholder
	}
	handle, err := handler.lookup([]byte(fullPath), true, newFileHandle)
	if err != nil || handle == nil {
		log.Printf("handler.ToHandle: failed for '%v': %v", fullPath, err)
		return nil
	}
	return handle
}

func (handler *nfsHandler) FromHandle(handle []byte) (fs billy.Filesystem, path []string, err error) {
	fs = handler.fs

	var fullPath []byte
	fullPath, err = handler.lookup(handle, false, nil)
	if err != nil || fullPath == nil {
		log.Printf("handler.FromHandle: could not resolve handle '%v': %v", handle, err)
		return nil, []string{}, &nfs.NFSStatusError{NFSStatus: nfs.NFSStatusStale}
	}

	fullPathStr := string(fullPath)
	if fullPathStr == rootPathPlaceholder {
		path = []string{""}
	} else {
		path = strings.Split(fullPathStr, string(filepath.Separator))
	}
	return
}

func (handler *nfsHandler) HandleLimit() int {
	return math.MaxInt32
}
