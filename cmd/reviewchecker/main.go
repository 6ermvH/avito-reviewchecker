package main

import (
	"context"
	"os"

	"github.com/6ermvH/avito-reviewchecker/cmd/reviewchecker/app"
	"github.com/6ermvH/avito-reviewchecker/cmd/reviewchecker/config"
)

func main() {
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
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
