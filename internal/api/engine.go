package api

import (
	"context"
	"sync"

	"endfield-music/internal/config"
	"endfield-music/internal/download"
	"endfield-music/internal/service"
)

var (
	engineOnce sync.Once
	engine     *download.Engine
)

func InitDownloadEngine(cfg *config.Config, taskService *service.TaskService) *download.Engine {
	engineOnce.Do(func() {
		apiURL := cfg.Netease.APIURL
		if apiURL == "" {
			apiURL = "http://127.0.0.1:3000"
		}
		concurrency := cfg.Download.Concurrency
		if concurrency <= 0 {
			concurrency = 2
		}
		netease := service.NewNeteaseService(apiURL)
		engine = download.NewEngine(netease, taskService, concurrency)
		engine.Start(context.Background())
	})
	return engine
}

func getDownloadEngine() *download.Engine {
	return engine
}
