package core

import (
	"github.com/pkg/errors"
	"github.com/reallyliri/sshonfs/cmd"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
)

var ciphers = []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"}

func sshConfig(config *cmd.Config) (*ssh.ClientConfig, error) {
	auth := []ssh.AuthMethod{
		ssh.Password(config.SshPassword),
	}

	publicKey, err := publicKeyFile(config.PrivateKeyFilePath)
	if err == nil {
		auth = append(auth, publicKey)
	} else {
		return nil, errors.Wrapf(err, "failed to parse key file at '%v'", config.PrivateKeyFilePath)
	}

	return &ssh.ClientConfig{
		User: config.SshUsername,
		Auth: auth,
		Config: ssh.Config{
			Ciphers: ciphers,
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}, nil
}

func publicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse file")
	}
	return ssh.PublicKeys(key), nil
}
