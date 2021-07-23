package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/renatocron/etamu-file2tar/pkg/file2tar"
)

func init() {

	if file2tar.WorkDir == "" {
		panic("Missing ENV WORK_DIR")
	}
	if _, err := os.Stat(file2tar.WorkDir); os.IsNotExist(err) {
		panic(fmt.Sprintf("Missing WORK_DIR does not exists: %v", err))
	}
	if !strings.HasSuffix(file2tar.WorkDir, "/") {
		file2tar.WorkDir = file2tar.WorkDir + "/"
	}

	if file2tar.BaseDir == "" {
		panic("Missing ENV BASE_DIR")
	}
	if _, err := os.Stat(file2tar.BaseDir); os.IsNotExist(err) {
		panic(fmt.Sprintf("Missing BASE_DIR does not exists: %v", err))
	}
	if !strings.HasSuffix(file2tar.BaseDir, "/") {
		file2tar.BaseDir = file2tar.BaseDir + "/"
	}

}

func main() {
	const BG_WORKERS = 4
	app := echo.New()

	// Middlewares
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	cm := file2tar.NewControlFileManager()
	err := cm.AddControlFromDir(file2tar.WorkDir)
	if err != nil {
		panic(fmt.Sprintf("Cannot start, AddControlFromDir failed with %v", err))
	}

	app.POST("/add", func(c echo.Context) error {

		return file2tar.Adding_post(c, func(frl *file2tar.FileResponseList) error {
			cm.AddControlFile(frl)
			println(frl.Timestamp)
			return nil
		})
	})

	app.GET("/debug", func(c echo.Context) error {

		return c.String(http.StatusOK, cm.DebugToStr())
	})

	// Start server
	go func() {
		if err := app.Start(":1323"); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("Desligando http")
		}
	}()

	consumer := file2tar.Consumer{
		IngestChan: make(chan *file2tar.ControlFile, 1),
		JobsChan:   make(chan *file2tar.ControlFile, 1000),
	}

	// Set up cancellation context and waitgroup
	ctx, cancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	go consumer.ProxyMessages(ctx)

	for i := 0; i < BG_WORKERS; i++ {
		wg.Add(1)
		go consumer.WorkerFunc(wg, i)
	}

	producer := file2tar.Producer{CallbackFunc: consumer.CallbackFunc}
	go producer.Start(cm)

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("Sinal para desligar recebido!")
	cancelFunc() // Signal cancellation to context.Context
	wg.Wait()    // Block here until are workers are done

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		app.Logger.Fatal(err)
	}

}
