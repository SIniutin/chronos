package main

import (
	"context"
	"net/http"

	"github.com/SIniutin/history-app-backend/internal/app"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		panic(err)
	}
	defer application.Close()

	addr := ":" + application.Port()
	application.Logger().Info("server started", zap.String("addr", addr))
	if err := http.ListenAndServe(addr, application.Router()); err != nil {
		application.Logger().Fatal("server failed", zap.Error(err))
	}
}
