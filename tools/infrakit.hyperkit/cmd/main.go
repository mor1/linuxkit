package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/plugin/metadata"
	instance_plugin "github.com/docker/infrakit/pkg/rpc/instance"
	metadata_plugin "github.com/docker/infrakit/pkg/rpc/metadata"
	instance_spi "github.com/docker/infrakit/pkg/spi/instance"
	"github.com/docker/infrakit/pkg/template"
)

const (
	// Default path when used with Docker for Mac
	defaultHyperKit = "/Applications/Docker.app/Contents/MacOS/com.docker.hyperkit"
)

var (
	// Version is the build release identifier.
	Version = "Unspecified"

	// Revision is the build source control revision.
	Revision = "Unspecified"

	// Default path to the VPNKit socket on Docker for Mac
	defaultVPNKitSock = "Library/Containers/com.docker.docker/Data/s50"
)

func main() {

	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "HyperKit instance plugin",
	}

	defaultVMDir := filepath.Join(getHome(), ".infrakit/hyperkit-vms")
	defaultVPNKitSock = path.Join(getHome(), defaultVPNKitSock)

	name := cmd.Flags().String("name", "instance-hyperkit", "Plugin name to advertise for discovery")
	logLevel := cmd.Flags().Int("log", cli.DefaultLogLevel, "Logging level. 0 is least verbose. Max is 5")

	vmDir := cmd.Flags().String("vm-dir", defaultVMDir, "Directory where to store VM state")
	hyperkit := cmd.Flags().String("hyperkit", defaultHyperKit, "Path to HyperKit executable")

	vpnkitSock := cmd.Flags().String("vpnkit-sock", defaultVPNKitSock, "Path to VPNKit UNIX domain socket")

	cmd.RunE = func(c *cobra.Command, args []string) error {
		opts := template.Options{
			SocketDir: discovery.Dir(),
		}
		thyper, err := template.NewTemplate("str://"+hyperkitArgs, opts)
		if err != nil {
			return err
		}
		tkern, err := template.NewTemplate("str://"+hyperkitKernArgs, opts)
		if err != nil {
			return err
		}

		os.MkdirAll(*vmDir, os.ModePerm)

		cli.SetLogLevel(*logLevel)
		cli.RunPlugin(*name,
			instance_plugin.PluginServer(NewHyperKitPlugin(*vmDir, *hyperkit, *vpnkitSock, thyper, tkern)),
			metadata_plugin.PluginServer(metadata.NewPluginFromData(
				map[string]interface{}{
					"version":    Version,
					"revision":   Revision,
					"implements": instance_spi.InterfaceSpec,
				},
			)),
		)
		return nil
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "print build version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			buff, err := json.MarshalIndent(map[string]interface{}{
				"version":  Version,
				"revision": Revision,
			}, "  ", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(buff))
			return nil
		},
	})

	if err := cmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func getHome() string {
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return os.Getenv("HOME")
}
