package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/konkers/lacodex/ingest"

	"github.com/google/subcommands"
)

type processCmd struct{}

func (*processCmd) Name() string     { return "process" }
func (*processCmd) Synopsis() string { return "Process and image and output it's JSON record." }
func (*processCmd) Usage() string {
	return `process <imange>:
	Process and image and output it's JSON record."
  `
}
func (p *processCmd) SetFlags(f *flag.FlagSet) {}

func (p *processCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) != 1 {
		fmt.Printf("Only one file at a time is supported.\n")
		return subcommands.ExitFailure
	}

	img, err := openImage(f.Args()[0])
	if err != nil {
		fmt.Printf("%v\n", err)
		return subcommands.ExitFailure
	}

	record, err := ingest.IngestImage(img)
	if err != nil {
		fmt.Printf("Ingest error: %v\n", err)
		return subcommands.ExitFailure
	}

	b, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		fmt.Printf("Failed to encode record: %v\n", err)
		return subcommands.ExitFailure
	}

	os.Stdout.Write(b)
	return subcommands.ExitSuccess
}
