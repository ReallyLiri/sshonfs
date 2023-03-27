package core

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

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
