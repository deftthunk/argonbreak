package main

import (
    "fmt"
    "time"
    "bufio"
    "log"
    "os"
    "github.com/matthewhartstonge/argon2"
    ds "github.com/golang-collections/go-datastructures/queue"
)

var pf = fmt.Printf

func argon_config() argon2.Config {
    argon := argon2.DefaultConfig()

    argon.HashLength = 32
    argon.SaltLength = 16
    argon.TimeCost = 2
    argon.MemoryCost = 10 * 1024
    argon.Parallelism = 2
    argon.Mode = argon2.ModeArgon2id
    argon.Version = argon2.Version13

    return argon
}

func main() {
    var threads int = 4
    var ch chan []byte = make(chan []byte, threads)
    var counter int = 0
    var block_size int = 100
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
        rb := ds.NewRingBuffer(10)
        begin := time.Now()
        delay := 1 * time.Second
        last, rate := 0, 0

        for {
            rb.Put(counter)
            time.Sleep(delay)

            pf("\r                                        ")
            pf("\rCur: %s\tRate: %d/sec\tTotal: %d        ", word, rate, counter)

            // FIX averaging not working yet
            if time.Since(begin) > (3 * time.Second) {
                begin = time.Now()
                rate = (counter - last) / 3
                placeholder, _ := rb.Get()
                last = placeholder.(int)
            }

            if rb.Len() > 9 {
                _, _ = rb.Get()
            }
        }
    }
    go status()

    // spinup
    for i:=0; i<threads; i++ {
        go crack(argon, []string{"spinup"}, win, ch)
    }

    for {
        work_block := make([]string, block_size)
        for i:=0; i<block_size; i++ {
            scanner.Scan()
            work_block[i] = scanner.Text()
        }

        word = work_block[0]

        select {
            case retn := <-ch:
                if string(retn) == win {
                    fmt.Println("Successfully cracked password")
                }
                go crack(argon, work_block, win, ch)
                counter += block_size
                continue
        }
    }
}

func crack(argon argon2.Config, work_block []string, win string, ch chan []byte) {
    var hash []byte
    for _, password := range work_block {
        hash, _ = argon.HashEncoded([]byte(password))
    }

    // PLACEHOLDER
    if win == string(hash) {
        fmt.Println("yay we found it")
    }
    ch<-hash
}



