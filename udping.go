package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var (
	msg  = []byte("Hello, World")
	addr = flag.String("a", "127.0.0.1:45643", "udp address")
)

func server(ctx context.Context) error {
	var lc net.ListenConfig
	pc, err := lc.ListenPacket(ctx, "udp", *addr)
	if err != nil {
		return err
	}
	context.AfterFunc(ctx, func() {
		pc.Close()
	})

	for {
		var buf [512]byte
		n, addr, err := pc.ReadFrom(buf[0:])
		if n != 0 {
			if n != len(msg) || !bytes.Equal(buf[:n], msg) {
				log.Printf("unexpected message %q from %v", buf[:n], addr)
				continue
			}
			if _, err := pc.WriteTo(buf[:n], addr); err != nil {
				log.Printf("error writing reply to %v: %v", addr, err)
				continue
			}
			log.Printf("successful ping-pong from %v", addr)
		}
		if err != nil {
			return err
		}
	}
}

func client(ctx context.Context) error {
	start := time.Now()
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "udp", *addr)
	if err != nil {
		return err
	}
	context.AfterFunc(ctx, func() {
		conn.Close()
	})

	if _, err := conn.Write(msg); err != nil {
		return err
	}
	var buf [512]byte
	n, err := conn.Read(buf[:])
	if err != nil {
		return err
	}
	if n != len(msg) || !bytes.Equal(buf[:n], msg) {
		return fmt.Errorf("unexpected reply %q from %v", buf[:n], *addr)
	}
	log.Printf("successful ping-pong to %v in %v", *addr, time.Since(start))
	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	flag.Parse()
	run := server
	if flag.Arg(0) == "client" {
		run = client
	}
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
