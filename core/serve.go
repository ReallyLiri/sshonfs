package core

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/reallyliri/sshonfs/cmd"
	"github.com/willscott/go-nfs"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
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

func mount(mountOptions string, mountPoint string, servePort string) error {
	if mountOptions == "" {
		switch runtime.GOOS {
		case "windows":
			mountOptions = fmt.Sprintf("port=%v,mountport=%v", servePort, servePort)
		case "darwin":
			fmt.Println("Running on macOS")
			mountOptions = fmt.Sprintf("port=%v,mountport=%v", servePort, servePort)
		default:
			mountOptions = fmt.Sprintf("port=%v,mountport=%v,nfsvers=3,noacl,tcp", servePort, servePort)
		}
	}
	return runCommand("mount", "-o", mountOptions, "-t", "nfs", "localhost:/", mountPoint)
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

func umount(mountPoint string) error {
	return runCommand("umount", mountPoint)
}

func runCommand(name string, args ...string) error {
	log.Printf("running '%v %v'", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "%v command failed to execute\n%v", name, string(output))
	}
	return nil
}
