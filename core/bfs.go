package core

import (
	"github.com/go-git/go-billy/v5"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"os"
	"path/filepath"
)

type sshFs struct {
	sftpClient *sftp.Client
	sshRoot    string
}

var _ billy.Filesystem = &sshFs{}

type sshFile struct {
	*sftp.File
}

var _ billy.File = &sshFile{}

func (fs *sshFs) Capabilities() billy.Capability {
	return billy.ReadAndWriteCapability | billy.SeekCapability | billy.TruncateCapability
}

func (fs *sshFs) Create(filename string) (billy.File, error) {
	file, err := fs.sftpClient.Create(filepath.Join(fs.sshRoot, filename))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create file at '%v'", filename)
	}
	return &sshFile{file}, nil
}

func (fs *sshFs) Open(filename string) (billy.File, error) {
	file, err := fs.sftpClient.Open(filepath.Join(fs.sshRoot, filename))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file at '%v'", filename)
	}
	return &sshFile{file}, nil
}

func (fs *sshFs) OpenFile(filename string, flag int, _ os.FileMode) (billy.File, error) {
	file, err := fs.sftpClient.OpenFile(filepath.Join(fs.sshRoot, filename), flag)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file at '%v' with flag %v", filename, flag)
	}
	return &sshFile{file}, nil
}

func (fs *sshFs) Stat(filename string) (os.FileInfo, error) {
	return fs.sftpClient.Stat(filepath.Join(fs.sshRoot, filename))
}

func (fs *sshFs) Rename(oldpath, newpath string) error {
	return fs.sftpClient.Rename(filepath.Join(fs.sshRoot, oldpath), filepath.Join(fs.sshRoot, newpath))
}

func (fs *sshFs) Remove(filename string) error {
	return fs.sftpClient.Remove(filepath.Join(fs.sshRoot, filename))
}

func (fs *sshFs) Join(elem ...string) string {
	return fs.sftpClient.Join(elem...)
}

func (fs *sshFs) TempFile(dir, prefix string) (billy.File, error) {
	return nil, errors.New("TempFile is not supported")
}

func (fs *sshFs) ReadDir(path string) ([]os.FileInfo, error) {
	return fs.sftpClient.ReadDir(filepath.Join(fs.sshRoot, path))
}

func (fs *sshFs) MkdirAll(filename string, _ os.FileMode) error {
	return fs.sftpClient.MkdirAll(filepath.Join(fs.sshRoot, filename))
}

func (fs *sshFs) Lstat(filename string) (os.FileInfo, error) {
	return fs.sftpClient.Lstat(filepath.Join(fs.sshRoot, filename))
}

func (fs *sshFs) Symlink(target, link string) error {
	return errors.New("Symlink is not supported")
}

func (fs *sshFs) Readlink(link string) (string, error) {
	return "", errors.New("Readlink is not supported")
}

func (fs *sshFs) Chroot(path string) (billy.Filesystem, error) {
	return nil, errors.New("Chroot is not supported")
}

func (fs *sshFs) Root() string {
	return "/"
}

func (f *sshFile) Lock() error {
	return errors.New("Lock is not supported")
}

func (f *sshFile) Unlock() error {
	return errors.New("Unlock is not supported")
}
