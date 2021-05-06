// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package start

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vmware/hamlet/examples/server/pkg/lifecycle_v1alpha1"
	"github.com/vmware/hamlet/examples/server/pkg/lifecycle_v1alpha2"
)

// flagSet represents the flags available with the start subcommand.
type flagSet struct {
	RootCACerts []string
	PeerCert    string
	PeerKey     string
	Port        uint32
	ApiVersion  string
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
			if flags.ApiVersion == "v1alpha1" {
				lifecycle_v1alpha1.Start(flags.RootCACerts, flags.PeerCert,
					flags.PeerKey, flags.Port)
			} else if flags.ApiVersion == "v1alpha2" {
				lifecycle_v1alpha2.Start(flags.RootCACerts, flags.PeerCert,
					flags.PeerKey, flags.Port, "token-1234")
			} else {
				fmt.Fprintf(os.Stderr, "Could not find version %s for client\n", flags.ApiVersion)
				os.Exit(1)
			}

		},
	}

	cmd.Flags().StringArrayVar(&flags.RootCACerts, "root-ca-cert", []string{}, "a root CA certificate path")
	cmd.Flags().StringVar(&flags.PeerCert, "peer-cert", "", "the peer certificate path")
	cmd.Flags().StringVar(&flags.PeerKey, "peer-key", "", "the peer key path")
	cmd.Flags().Uint32Var(&flags.Port, "port", 8000, "the port to listen for requests on")
	cmd.Flags().StringVar(&flags.ApiVersion, "api-version", "", "api version to use v1alpha1 or v1alpha2")

	cmd.MarkFlagRequired("root-ca-cert")
	cmd.MarkFlagRequired("peer-cert")
	cmd.MarkFlagRequired("peer-key")
	cmd.MarkFlagRequired("api-version")
	return cmd
}
