package main

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/jarxorg/gcsfs"
)

func main() {
	fsys := gcsfs.New("trycatch.jp")
	matches, err := fs.Glob(fsys, "*")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", matches)
}
