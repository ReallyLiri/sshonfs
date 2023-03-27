package main

import (
	"github.com/reallyliri/sshonfs/cmd"
	"github.com/reallyliri/sshonfs/core"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	cmd.Execute(core.Serve)
}
