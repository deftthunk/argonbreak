package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/matthewhartstonge/argon2"
)

var pf = fmt.Printf
var pl = fmt.Println

func argon_config() argon2.Config {
	argon := argon2.DefaultConfig()

	argon.HashLength = 64
	argon.SaltLength = 16
	argon.TimeCost = 5
	argon.MemoryCost = 10 * 1024
	argon.Parallelism = 2
	argon.Mode = argon2.ModeArgon2id
	argon.Version = argon2.Version13

	return argon
}

func main() {
	var threads int = 8
	var ch chan []byte = make(chan []byte, threads)
	var counter int = 0
	var block_size int = 50
	var word string = ""
	var win string = "FOUNDIT!"

	file, err := os.Open("passwords_10m.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	argon := argon_config()

	status := func() {
		buf_size := int64(30)
		hist_buffer := queue.New(buf_size)
		delay := 1 * time.Second
		rate, average := float64(0), float64(0)

		for {
			hist_buffer.Put(counter)
			time.Sleep(delay)

			pf("\r                                            ")
			pf("\rCur: %s\t  Rate: %.2f/sec\tTotal: %d        ", word, rate, counter)

			if hist_buffer.Len() >= buf_size {
				buf_item, _ := hist_buffer.Get(1)
				popped_count := float64(buf_item[0].(int))
				buf_item2, _ := hist_buffer.Peek()
				next_count := float64(buf_item2.(int))

				average = (average + popped_count + next_count) / 3
				rate = (float64(counter) - average) / float64(buf_size)
			} else {
				last, _ := hist_buffer.Peek()
				average = (average + float64(last.(int))) / 2
				rate = (float64(counter) - average) / float64(hist_buffer.Len())
			}
		}
	}
	go status()

	thread_count := 0
	for {
		work_block := make([]string, block_size)
		for i := 0; i < block_size; i++ {
			scanner.Scan()
			work_block[i] = scanner.Text()
		}

		word = work_block[0]
		// spinup set number of threads
		if thread_count < threads {
			go crack(argon, work_block, win, ch)
			counter += block_size
			thread_count++
			time.Sleep(800 * time.Millisecond)
			continue
		}

		retn := <-ch
		if string(retn) == win {
			pl("Successfully cracked password")
		}
		go crack(argon, work_block, win, ch)
		counter += block_size
	}
}

func crack(argon argon2.Config, work_block []string, win string, ch chan []byte) {
	var hash []byte
	for _, password := range work_block {
		hash, _ = argon.HashEncoded([]byte(password))
	}

	ch <- hash
}
