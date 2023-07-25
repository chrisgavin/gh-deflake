package main

import (
	"context"
	"errors"
	"os"

	"github.com/AlecAivazis/survey/v2/terminal"
	log "github.com/sirupsen/logrus"

	"github.com/chrisgavin/gh-deflake/cmd"
)

func main() {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	if err := cmd.Execute(ctx); err != nil {
		if errors.Is(err, cmd.SilentErr) {
			os.Exit(1)
		}
		if errors.Is(err, terminal.InterruptErr) {
			os.Exit(128 + 2)
		}
		log.Fatalf("%+v", err)
	}
}
