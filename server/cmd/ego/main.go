package main

import (
	"log"

	"ego-server/internal/bootstrap"
	"ego-server/internal/config"
)

func main() {
	cfg := config.Load()

	p, err := bootstrap.InitPlatform(cfg)
	if err != nil {
		log.Fatalf("init platform: %v", err)
	}
	defer p.Close()

	identityHandler := bootstrap.NewIdentityHandler(p)
	writingHandler := bootstrap.NewWritingHandler(p)
	timelineHandler := bootstrap.NewTimelineHandler(p)
	handler := bootstrap.NewEgoHandler(identityHandler, writingHandler, timelineHandler)
	server := bootstrap.NewServer(cfg, p, handler)

	if err := server.Serve(); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
