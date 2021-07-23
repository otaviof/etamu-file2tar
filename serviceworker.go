package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Consumer struct {
	ingestChan chan *ControlFile
	jobsChan   chan *ControlFile
}

// callbackFunc is invoked each time the external lib passes an event to us.
func (c Consumer) callbackFunc(cf *ControlFile) {
	c.ingestChan <- cf
}

// workerFunc starts a single worker function that will range on the jobsChan until that channel closes.
func (c Consumer) workerFunc(wg *sync.WaitGroup, workerId int) {
	defer wg.Done()

	fmt.Printf("Novo Trabalhador! %d\n", workerId)
	for eventIndex := range c.jobsChan {

		fmt.Printf("executando serviço %v < index \n", eventIndex)
		//time.Sleep(time.Second / 3)
		time.Sleep(time.Second * 20)

		eventIndex.Lock()
		eventIndex.is_done = true
		eventIndex.is_working = false
		eventIndex.Unlock()
		//fmt.Printf("still doing work\n")

		fmt.Printf("serviço finalizado %v!\n\n", eventIndex)

	}
	fmt.Printf("Trabalhador %d desligado!\n", workerId)
}

func (c Consumer) proxyMessages(ctx context.Context) {
	for {
		fmt.Printf("comeco do proxyMessages\n")
		select {
		case job := <-c.ingestChan:
			fmt.Printf("proxying job...\n")
			c.jobsChan <- job
			fmt.Printf("job was proxyed!\n")

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
	callbackFunc func(cf *ControlFile)
}

func (p Producer) start(cm *ControlFileManager) {
	const WAIT_TIME = 5
	for {
		now := time.Now().Unix()
		fmt.Printf("now is %d\n", now)

		cm.Lock()

		newJobs := make([]*ControlFile, 0)

		for _, v := range cm.jobs {

			v.Lock()
			println("v " + v.DebugToStr())
			if v.is_done {

				v.Unlock()
				continue
			} else if !v.is_working && now-v.timestamp > WAIT_TIME {
				v.is_working = true

				p.callbackFunc(v)
			}
			v.Unlock()
			newJobs = append(newJobs, v)
		}
		cm.jobs = newJobs

		cm.Unlock()

		time.Sleep(time.Second)
	}
}
