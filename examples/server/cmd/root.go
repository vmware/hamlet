// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vmware/hamlet/examples/server/cmd/start"
	"github.com/vmware/hamlet/examples/server/cmd/version"
)

// RootCmd represents the base command to be executed when called without any
// command line arguments.
var RootCmd = &cobra.Command{
	Use:   "server",
	Short: "Server Example",
	Long:  "Server Example",
}

// init initializes the root command instance.
func init() {
	RootCmd.AddCommand(start.NewCommand())
	RootCmd.AddCommand(version.NewCommand())
}
