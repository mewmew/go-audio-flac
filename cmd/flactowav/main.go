// This tool converts a FLAC file into an identical aiff file and stores
// it in the same folder as the source.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/go-audio/flac"
)

var (
	flagPath = flag.String("path", "", "The path to the FLAC file to convert to aiff")
)

func main() {
	flag.Parse()
	if *flagPath == "" {
		fmt.Println("You must set the -path flag")
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		log.Println("Failed to get the user home directory")
		os.Exit(1)
	}

	sourcePath := *flagPath
	if sourcePath[:2] == "~/" {
		sourcePath = strings.Replace(sourcePath, "~", usr.HomeDir, 1)
	}

	f, err := os.Open(*flagPath)
	if err != nil {
		fmt.Println("Invalid path", *flagPath, err)
		os.Exit(1)
	}
	defer f.Close()

	d, err := flac.NewDecoder(f)
	if err != nil {
		fmt.Println("Unable to decode FLAC file", *flagPath, err)
		os.Exit(1)
	}

	outPath := sourcePath[:len(sourcePath)-len(filepath.Ext(sourcePath))] + ".aiff"
	of, err := os.Create(outPath)
	if err != nil {
		fmt.Println("Failed to create", outPath)
		panic(err)
	}
	defer of.Close()

	format := d.Format()
	e := aiff.NewEncoder(of, int(format.SampleRate), int(d.SampleBitDepth()), int(format.NumChannels))

	bufferSize := 1000000
	buf := &audio.IntBuffer{Data: make([]int, bufferSize), Format: format}
	var n int
	for err == nil {
		n, err = d.PCMBuffer(buf)
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
		if n != len(buf.Data) {
			buf.Data = buf.Data[:n]
		}
		if err := e.Write(buf); err != nil {
			panic(err)
		}
	}

	if err := e.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("FLAC file converted to %s\n", outPath)
}
