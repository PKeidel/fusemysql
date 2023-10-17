package main

import (
	"context"
	"log"

	"github.com/PKeidel/goargs"
)

var (
	// version, commit, date are for goreleaser
    version string = "development"
    commit string = "?"
    date string = "?"
    app *FusemysqlApp

    args struct {
        mountpoint string
    }
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

    mountArg := goargs.WithStringS("mount", "m", &args.mountpoint)
    mountArg.Required = true

    goargs.Parse()

    app = NewFusemysqlApp()
}

func main() {
    log.Printf("Fusemysql in version %s (commit: %s, build on: %s)\n", version, commit, date)

	ctx := context.Background()

    errCh := make(chan error, 1)

    go func() {
        errCh <- app.Run(ctx)
    }()


	err := <-errCh
	if err != nil {
		log.Fatalf("Error: %#v", err)
	}
}
