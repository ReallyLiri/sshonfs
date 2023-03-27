package core

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/reallyliri/sshonfs/cmd"
	"github.com/willscott/go-nfs"
	"github.com/willscott/go-nfs/helpers"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

func Serve(config *cmd.Config) (io.Closer, error) {
	listenAddress := ":" + config.LocalServePort
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %v: %v", listenAddress, err)
	}

	handler, err := newHandler(config)
	if err != nil {
		return nil, err
	}

	log.Printf("nfs server running at %s and serving '%v'", listener.Addr(), config.SshAddress)

	go func() {
		err = nfs.Serve(listener, handler)
		fmt.Printf("serve failed: %v", err)
	}()

	mountPath := config.MountPath
	if config.SkipMount {
		mountPath = ""
	} else {
		log.Printf("mounting at '%v'", mountPath)
		os.MkdirAll(mountPath, 0777)
		umount(mountPath)
		err = mount(config.MountOptions, mountPath, config.LocalServePort)
		if err != nil {
			return nil, err
		}
	}

	return &closer{mountPoint: mountPath, listener: listener}, nil
}

func newHandler(config *cmd.Config) (nfs.Handler, error) {
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

	bfs := &sshFs{
		sftpClient: sftpClient,
		sshRoot:    config.SshRootPath,
	}

	return helpers.NewCachingHandler(helpers.NewNullAuthHandler(bfs), 1024*1024), nil
}

type closer struct {
	listener   net.Listener
	mountPoint string
}

var _ io.Closer = &closer{}

func (u *closer) Close() error {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		u.listener.Close()
	}()
	go func() {
		defer wg.Done()
		if u.mountPoint != "" {
			umount(u.mountPoint)
		}
	}()
	wg.Wait()
	return nil
}
