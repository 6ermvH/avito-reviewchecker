package main

import (
	"context"

	"github.com/6ermvH/avito-reviewchecker/cmd/reviewchecker/app"
	"github.com/6ermvH/avito-reviewchecker/cmd/reviewchecker/config"
)

func main() {
	cfg, err := config.Load("configs/developer.yaml")
	if err != nil {
		panic(err)
	}

	application, err := app.New(*cfg)
	if err != nil {
		panic(err)
	}

	if err := application.Run(context.Background()); err != nil {
		panic(err)
	}
}
