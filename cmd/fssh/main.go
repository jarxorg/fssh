package main

import (
	"log"
	"os"

	"github.com/jarxorg/fssh"
	_ "github.com/jarxorg/fssh/command"
)

func main() {
	if err := fssh.Main(os.Args); err != nil {
		log.Fatal(err)
	}
}
