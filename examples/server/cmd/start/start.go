// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package start

import (
	"github.com/spf13/cobra"
	"github.com/vmware/hamlet/examples/server/pkg/lifecycle"
)

// flagSet represents the flags available with the start subcommand.
type flagSet struct {
	RootCACerts []string
	PeerCert    string
	PeerKey     string
	Port        uint32
}

// NewCommand returns a new Command instance for the start subcommand.
func NewCommand() *cobra.Command {
	flags := &flagSet{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "start",
		Short: "Start the server",
		Long:  "Start the server",
		Run: func(cmd *cobra.Command, args []string) {
			lifecycle.Start(flags.RootCACerts, flags.PeerCert,
				flags.PeerKey, flags.Port)
		},
	}

	cmd.Flags().StringArrayVar(&flags.RootCACerts, "root-ca-cert", []string{}, "a root CA certificate path")
	cmd.Flags().StringVar(&flags.PeerCert, "peer-cert", "", "the peer certificate path")
	cmd.Flags().StringVar(&flags.PeerKey, "peer-key", "", "the peer key path")
	cmd.Flags().Uint32Var(&flags.Port, "port", 8000, "the port to listen for requests on")

	cmd.MarkFlagRequired("root-ca-cert")
	cmd.MarkFlagRequired("peer-cert")
	cmd.MarkFlagRequired("peer-key")
	return cmd
}
