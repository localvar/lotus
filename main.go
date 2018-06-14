package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/localvar/go-utils/config"
	"github.com/localvar/go-utils/log"
	"github.com/localvar/lotus/app"
	"github.com/localvar/lotus/models"
)

func serveWeb() *http.Server {
	port := config.GetString("/app/port", "80")
	log.Infoln("Listening at port:", port)

	srv := &http.Server{Addr: ":" + port}
	go func() {
		if e := srv.ListenAndServe(); e != nil {
			log.Errorln(e.Error())
		}
	}()

	return srv
}

func initLogger() error {
	log.Default.FileNamePrefix = "LOTUS_"
	log.Default.Period = 24 * time.Hour
	log.Default.WithFile = true
	log.Default.Folder = config.GetString("/log/folder", "")
	log.Default.ToStdErr = config.GetBool("/log/tostderr", true)
	log.Default.MinLevel = log.Level(config.GetInt("/log/minlevel", int(log.Info)))
	log.Default.ToFile = config.GetBool("/log/tofile", true)
	return log.Default.Start()
}

func main() {
	var mode = flag.String("mode", "", "application running mode, 'debug', 'release' or 'initdb'")
	flag.Parse()

	// load configuration file
	// use fmt.Fprintln in case of error because log not ready yet
	if e := config.ParseIniFile("config.ini"); e != nil {
		fmt.Fprintln(os.Stderr, "failed to load configuration:", e.Error())
		return
	}

	if e := initLogger(); e != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger:", e.Error())
		return
	}
	defer log.Default.Close()

	if len(*mode) == 0 {
		*mode = config.GetString("/app/mode", "release")
	}

	if e := models.Init(*mode == "debug"); e != nil {
		log.Fatalln("failed to connect to database:", e.Error())
	}

	if e := app.Init(*mode == "debug"); e != nil {
		log.Fatalln("failed to initialize application:", e.Error())
	}

	// start web server
	srv := serveWeb()

	// wait `Ctrl-C` or `Kill` to exit, `Kill` does not work on Windows
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, os.Kill)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	srv.Shutdown(ctx)
	cancel()

	app.Uninit()
	models.Uninit()
}
