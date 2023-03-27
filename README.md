# ssh-on-nfs

Cross-platform solution to mount a file system accessible via ssh on a local path.

(Though Windows still don't work...)

![icon](sshonfs.png)

## Install

```shell
go install github.com/reallyliri/sshonfs@latest
```

## Pre-requisites

### Linux

Make sure you have `nfs-common` installed (i.e `apt install nfs-common`).

### Mac

Enable NFS: `sudo nfsd enable`.

### Windows

Server is up and running, but I couldn't make mount work, help is appreciated.

## Usage

```shell
sshonfs --help

Usage:
  sshonfs [flags]

Flags:
  -a, --address string         ssh server address (default "127.0.0.1:22")
  -h, --help                   help for sshonfs
  -o, --mount-options string   options to mount with, default options are OS dependent
  -m, --mount-path string      path to mount the ssh fs on (default ".")
  -p, --password string        ssh password
  -i, --private-key string     path to private ssh key (default "~/.ssh/id_rsa")
  -r, --root string            ssh root (default "/opt")
  -P, --serve-port string      local port to serve nfs server on (default "2049")
  -s, --skip-mount             skip mount, only serve
  -u, --username string        ssh username (default "root")
      --version                version for sshonfs
```

i.e

```shell
sshonfs -a aws-server -m ~/efs-test -i ~/.ssh/aws.pem -r /mnt/efs/ -P 2049 -u ubuntu
```

## Build

```shell
go build -o bin/sshonfs .
```
