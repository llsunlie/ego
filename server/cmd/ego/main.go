package main

import (
	"ego-server/internal/bootstrap"
	"ego-server/internal/config"
)

func main() {
	cfg := config.Load()

	p, err := bootstrap.InitPlatform(cfg)
	if err != nil {
		panic("init platform: " + err.Error())
	}
	defer p.Close()

	p.Logger.Info("ego server starting",
		"web_port", cfg.WebPort,
		"web_tls_port", cfg.WebTLSPort,
		"grpc_port", cfg.GRPCPort,
		"log_level", cfg.LogLevel,
		"log_format", cfg.LogFormat,
	)

	identityHandler := bootstrap.NewIdentityHandler(p)
	writingHandler := bootstrap.NewWritingHandler(p)
	timelineHandler := bootstrap.NewTimelineHandler(p)
	starmapHandler := bootstrap.NewStarmapHandler(p)
	chatHandler := bootstrap.NewChatHandler(p)
	settingHandler := bootstrap.NewSettingHandler(p)
	handler := bootstrap.NewEgoHandler(identityHandler, writingHandler, timelineHandler, starmapHandler, chatHandler, settingHandler)
	server := bootstrap.NewServer(cfg, p, handler)

	if err := server.Serve(); err != nil {
		p.Logger.Error("serve failed", "error", err)
		panic("serve: " + err.Error())
	}
}
