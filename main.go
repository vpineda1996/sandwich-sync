// The following is required by scotiafetch since the scotia server seems to break HTTP/2 streams
// For more info look at: https://github.com/googleapis/google-cloud-go/issues/7440
//
//go:debug http2client=0
package main

import (
	"fmt"
	"os"

	"github.com/vpnda/sandwich-sync/cmd/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
