package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {

	if WorkDir == "" {
		panic("Missing ENV WORK_DIR")
	}
	if _, err := os.Stat(WorkDir); os.IsNotExist(err) {
		panic(fmt.Sprintf("Missing WORK_DIR does not exists: %v", err))
	}
	if !strings.HasSuffix(WorkDir, "/") {
		WorkDir = WorkDir + "/"
	}

	if BaseDir == "" {
		panic("Missing ENV BASE_DIR")
	}
	if _, err := os.Stat(BaseDir); os.IsNotExist(err) {
		panic(fmt.Sprintf("Missing BASE_DIR does not exists: %v", err))
	}
	if !strings.HasSuffix(BaseDir, "/") {
		BaseDir = BaseDir + "/"
	}

}

func main() {
	app := echo.New()

	// Middlewares
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	cm := NewControlFileManager()
	err := cm.AddControlFromDir(WorkDir)
	if err != nil {
		panic(fmt.Sprintf("Cannot start, AddControlFromDir failed with %v", err))
	}

	app.POST("/add", func(c echo.Context) error {

		return adding_post(c, func(frl *FileResponseList) error {
			cm.AddControlFile(frl)
			println(frl.timestamp)
			return nil
		})
	})

	app.GET("/debug", func(c echo.Context) error {

		return c.String(http.StatusOK, cm.DebugToStr())
	})

	// Start server
	//go func() {
	if err := app.Start(":1323"); err != nil && err != http.ErrServerClosed {
		app.Logger.Fatal("Desligando http")
	}
	//}()
	/*
		consumer := Consumer{
			ingestChan: make(chan int),
			jobsChan:   make(chan int),
		}

		// Set up cancellation context and waitgroup
		ctx, cancelFunc := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		go consumer.proxyMessages(ctx)
		wg.Add(1)
		go consumer.workerFunc(wg)

		producer := Producer{callbackFunc: consumer.callbackFunc}
		go producer.start()

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
	*/
}

/*

type Consumer struct {
	ingestChan chan int
	jobsChan   chan int
}

// callbackFunc is invoked each time the external lib passes an event to us.
func (c Consumer) callbackFunc(event int) {
	c.ingestChan <- event
}

// workerFunc starts a single worker function that will range on the jobsChan until that channel closes.
func (c Consumer) workerFunc(wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Novo Trabalhador!\n")
	for eventIndex := range c.jobsChan {

		fmt.Printf("executando serviço %d < index \n", eventIndex)
		//time.Sleep(time.Second / 3)
		//fmt.Printf("still doing work\n")
		time.Sleep(time.Second * 4)
		fmt.Printf("serviço finalizado %d!\n\n", eventIndex)

	}
	fmt.Printf("Trabalhador desligado!\n")
}

func (c Consumer) proxyMessages(ctx context.Context) {
	for {
		fmt.Printf("comeco do proxyMessages\n")
		select {
		case job := <-c.ingestChan:
			fmt.Printf("publicando novo serviço...\n")
			c.jobsChan <- job
			fmt.Printf("servico publicado no canal!\n")

		case <-ctx.Done():
			fmt.Println("Sinal de cancelar recebido, fechando canal de trabalhos!")
			close(c.jobsChan)
			fmt.Println("canal de trabalhos encerrados")

			return

		}
		fmt.Printf("fim do proxyMessages\n")

	}
}

type Producer struct {
	callbackFunc func(event int)
}

func (p Producer) start() {
	eventIndex := 1
	for {
		fmt.Printf("preparando novo servico %d\n", eventIndex)
		p.callbackFunc(eventIndex)
		eventIndex++

		//time.Sleep(time.Second)
	}
}

*/
