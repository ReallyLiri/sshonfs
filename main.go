package main

import (
	"github.com/reallyliri/sshonfs/cmd"
	"github.com/reallyliri/sshonfs/core"
	"io"
	"log"
	"os"
)

const isDebug = true

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	if isDebug {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.Discard)
	}

	cmd.Execute(core.Serve)
}
