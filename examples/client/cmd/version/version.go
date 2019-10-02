// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version represents the client version.
const version = "0.1.0"

// versionCmd represents the command to be executed when called with the version
// subcommand.
var versionCmd = &cobra.Command{
	Args:  cobra.NoArgs,
	Use:   "version",
	Short: "Print the version",
	Long:  "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client version", version)
	},
}

// NewCommand returns the Command instance for the version subcommand.
func NewCommand() *cobra.Command {
	return versionCmd
}
