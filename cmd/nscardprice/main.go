// Command nscardprice is a CLI for querying Nintendo Switch cartridge
// recycle prices across multiple merchants.
package main

import (
	"os"

	"nscardprice/internal/cli"
)

func main() {
	os.Exit(cli.Run())
}
