package run

import (
	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/gfwresist"
	"paqet/internal/server"
)

func startServer(cfg *conf.Conf) {
	flog.Infof("Starting server...")

	if cfg.GFWResist.AutoIPTables {
		mgr := gfwresist.NewIPTablesManager(cfg.Listen.Addr.Port)
		if err := mgr.Apply(); err != nil {
			flog.Warnf("GFW-resist: failed to apply iptables rules: %v", err)
		} else {
			defer mgr.Cleanup()
		}
	}

	server, err := server.New(cfg)
	if err != nil {
		flog.Fatalf("Failed to initialize server: %v", err)
	}
	if err := server.Start(); err != nil {
		flog.Fatalf("Server encountered an error: %v", err)
	}
}
