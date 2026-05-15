package main

import (
	"cache"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	addr := flag.String("addr", "0.0.0.0:3000", "server listen address")
	advertiseAddr := flag.String("advertise-addr", "", "server address registered to etcd")
	capacity := flag.Int("capacity", 1024*1024*64, "cache capacity in bytes")
	etcdEndpoints := flag.String("etcd-endpoints", "", "comma-separated etcd endpoints")
	serviceName := flag.String("service-name", "cache", "service name registered to etcd")

	flag.Parse()

	server := cache.NewServer(*capacity, *addr)
	var registry *cache.EtcdRegistry
	if strings.TrimSpace(*etcdEndpoints) != "" {
		registerAddr := strings.TrimSpace(*advertiseAddr)
		if registerAddr == "" {
			registerAddr = *addr
		}
		var err error
		registry, err = cache.RegisterEtcd(context.Background(), strings.Split(*etcdEndpoints, ","), *serviceName, registerAddr)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := registry.Close(context.Background()); err != nil {
				log.Printf("failed to close etcd registry: %v", err)
			}
		}()
		log.Printf("cache server registered to etcd, service=%s, addr=%s", *serviceName, registerAddr)
	}

	log.Printf("cache server listening on %s, capacity=%d bytes", *addr, *capacity)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	case <-signalCh:
		log.Printf("cache server shutting down")
	}
}
