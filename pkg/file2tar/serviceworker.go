package file2tar

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Consumer struct {
	IngestChan chan *ControlFile
	JobsChan   chan *ControlFile
}

// CallbackFunc is invoked each time the external lib passes an event to us.
func (c Consumer) CallbackFunc(cf *ControlFile) {
	c.IngestChan <- cf
}

// WorkerFunc starts a single worker function that will range on the jobsChan until that channel closes.
func (c Consumer) WorkerFunc(wg *sync.WaitGroup, workerId int) {
	defer wg.Done()

	fmt.Printf("Novo Trabalhador! %d\n", workerId)
	for eventIndex := range c.JobsChan {

		eventIndex.Lock()
		fmt.Printf("executando serviço %v < index \n", eventIndex)
		eventIndex.Unlock()
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

func (c Consumer) ProxyMessages(ctx context.Context) {
	for {
		fmt.Printf("comeco do proxyMessages\n")
		select {
		case job := <-c.IngestChan:
			fmt.Printf("proxying job...\n")
			c.JobsChan <- job
			fmt.Printf("job was proxyed!\n")

		case <-ctx.Done():
			fmt.Println("Sinal de cancelar recebido, fechando canal de trabalhos!")
			close(c.JobsChan)
			fmt.Println("canal de trabalhos encerrados")

			return

		}
		fmt.Printf("fim do proxyMessages\n")

	}
}

type Producer struct {
	CallbackFunc func(cf *ControlFile)
}

func (p Producer) Start(cm *ControlFileManager) {
	const WAIT_TIME = 5
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	for {
		now := time.Now().Unix()
		fmt.Printf("now is %d\n", now)

		log.Print("locking cm")
		cm.Lock()
		newJobs := make([]*ControlFile, 0)
		for _, v := range cm.jobs {

			v.Lock()
			print("v " + v.DebugToStr())
			if v.is_done {

				v.Unlock()
				continue
			} else if !v.is_working && now-v.timestamp > WAIT_TIME {
				v.is_working = true

				p.CallbackFunc(v)
			}
			v.Unlock()
			newJobs = append(newJobs, v)
		}
		cm.jobs = newJobs

		cm.Unlock()
		log.Print("unlocking cm")

		time.Sleep(time.Second)
	}
}
