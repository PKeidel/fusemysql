package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func NewFusemysqlApp() *FusemysqlApp {
	return &FusemysqlApp{}
}

type FusemysqlApp struct {
	// your stuff
}

func (app *FusemysqlApp) Run(ctx context.Context) error {
	c, err := fuse.Mount(
		args.mountpoint,
		fuse.FSName("mysql"),
		fuse.Subtype("pkeidel.mysqlfs"),
		fuse.ReadOnly(),
	)
	if err != nil {
		return err
	}
	defer c.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
        sig := <-sigs
		log.Println("Signal received:", sig)

		fuse.Unmount(args.mountpoint)
    }()

	db := OpenConnection()
	defer db.Close()

	rootFs := NewFS(db)

	err = fs.Serve(c, rootFs)
	if err != nil {
		return err
	}

	return nil
}
