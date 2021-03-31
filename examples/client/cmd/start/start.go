// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package start

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vmware/hamlet/examples/client/pkg/lifecycle_v1alpha1"
	"github.com/vmware/hamlet/examples/client/pkg/lifecycle_v1alpha2"
)

// flagSet represents the flags available with the start subcommand.
type flagSet struct {
	RootCACert         string
	PeerCert           string
	PeerKey            string
	ServerAddr         string
	InsecureSkipVerify bool
	Context            string
	ApiVersion         string
}

// NewCommand returns a new Command instance for the start subcommand.
func NewCommand() *cobra.Command {
	flags := &flagSet{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "start",
		Short: "Start the client",
		Long:  "Start the client",
		Run: func(cmd *cobra.Command, args []string) {
			if flags.ApiVersion == "v1alpha1" {
				lifecycle_v1alpha1.Start(flags.RootCACert, flags.PeerCert,
					flags.PeerKey, flags.ServerAddr,
					flags.InsecureSkipVerify)
			} else if flags.ApiVersion == "v1alpha2" {
				lifecycle_v1alpha2.Start(flags.RootCACert, flags.PeerCert,
					flags.PeerKey, flags.ServerAddr,
					flags.InsecureSkipVerify, flags.Context)
			} else {
				fmt.Fprintf(os.Stderr, "Could not find version %s for client\n", flags.ApiVersion)
				os.Exit(1)
			}

		},
	}

	cmd.Flags().StringVar(&flags.RootCACert, "root-ca-cert", "", "the root CA certificate path")
	cmd.Flags().StringVar(&flags.PeerCert, "peer-cert", "", "the peer certificate path")
	cmd.Flags().StringVar(&flags.PeerKey, "peer-key", "", "the peer key path")
	cmd.Flags().StringVar(&flags.ServerAddr, "server-addr", "localhost:8000", "the server's address")
	cmd.Flags().BoolVar(&flags.InsecureSkipVerify, "insecure-skip-verify", false, "whether normal verification should be ignored")

	cmd.Flags().StringVar(&flags.Context, "context", "", "connection context")
	cmd.Flags().StringVar(&flags.ApiVersion, "api-version", "", "api version to use v1alpha1 or v1alpha2")

	cmd.MarkFlagRequired("root-ca-cert")
	cmd.MarkFlagRequired("peer-cert")
	cmd.MarkFlagRequired("peer-key")
	cmd.MarkFlagRequired("api-version")
	return cmd
}
