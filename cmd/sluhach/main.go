package main

import (
	"flag"
	"fmt"
	"os"

	"sluhach/internal/sluhach"
)

var (
	modelpath = flag.String("model", "vosk-model-small-ru-0.22", "path to model dir")
)

func main() {
	flag.Parse()

	_sluhach, err := sluhach.New(*modelpath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := _sluhach.Start(); err != nil {
		os.Exit(2)
	}
}
