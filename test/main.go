package main

import (
	"context"
	"geerpc"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int

type Args struct {
	Num1, Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatalf("network error: %v", err)
	}

	var foo Foo
	if err := geerpc.Register(&foo); err != nil {
		log.Fatalf("register error: %v", err)
	}

	geerpc.HandleHTTP()

	log.Printf("start rpc server on %v", l.Addr())
	addr <- l.Addr().String()
	_ = http.Serve(l, nil)
}

func call(addrCh chan string) {
	client, _ := geerpc.DialHTTP("tcp", <-addrCh)
	defer func() {
		_ = client.Close()
	}()

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatalf("call Foo.Sum error: %v", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()

}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go call(addr)
	startServer(addr)
}
