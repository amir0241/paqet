package run

import (
	"fmt"
	"log"
	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/pkg/buffer"

	"github.com/spf13/cobra"
)

var confPath string

func init() {
	Cmd.Flags().StringVarP(&confPath, "config", "c", "config.yaml", "Path to the configuration file.")
}

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the client or server based on the config file.",
	Long:  `The 'run' command reads the specified YAML configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := conf.LoadFromFile(confPath)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		if err := initialize(cfg); err != nil {
			log.Fatalf("Failed to initialize: %v", err)
		}

		switch cfg.Role {
		case "client":
			startClient(cfg)
			return
		case "server":
			startServer(cfg)
			return
		}

		log.Fatalf("Failed to load configuration")
	},
}

func initialize(cfg *conf.Conf) error {
	flog.SetLevel(cfg.Log.Level)
	if err := buffer.Initialize(cfg.Transport.TCPBuf, cfg.Transport.UDPBuf); err != nil {
		return fmt.Errorf("failed to initialize buffers: %w", err)
	}
	return nil
}
