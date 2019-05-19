package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/subcommands"
	"github.com/konkers/lacodex/ingest"
)

type gamecropCmd struct{}

func (*gamecropCmd) Name() string     { return "gamecrop" }
func (*gamecropCmd) Synopsis() string { return "Crop image to game size." }
func (*gamecropCmd) Usage() string {
	return `gamecrop <file>...:
	Crop file to game size.
  `
}
func (p *gamecropCmd) SetFlags(f *flag.FlagSet) {
}

func openImage(fileName string) (image.Image, error) {
	reader, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Can't open %s: %v", fileName, err)
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("Can't decode %s: %v", fileName, err)
	}

	return img, nil
}

func writeImage(fileName string, suffix string, img image.Image) error {
	ext := filepath.Ext(fileName)
	fileName = fmt.Sprintf(strings.TrimSuffix(fileName, ext) + "-" + suffix + ext)

	writer, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", fileName, err)
	}

	err = png.Encode(writer, img)
	if err != nil {
		return fmt.Errorf("failed to encode %s as png: %v", fileName, err)
	}

	return nil
}

func (p *gamecropCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	for _, fileName := range f.Args() {
		img, err := openImage(fileName)
		if err != nil {
			fmt.Printf("%v\n", err)
			return subcommands.ExitFailure
		}

		croppedImg := ingest.UtilCropGameImage(img)

		err = writeImage(fileName, "game", croppedImg)
		if err != nil {
			fmt.Printf("%v\n", err)
			return subcommands.ExitFailure
		}

	}
	return subcommands.ExitSuccess
}
