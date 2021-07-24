package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/0990/chinadns"
	"github.com/0990/chinadns/logconfig"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"time"
)

var confFile = flag.String("c", "chinadns.json", "config file")

func main() {
	flag.Parse()

	file, err := os.Open(*confFile)
	if err != nil {
		logrus.Fatalln(err)
	}

	var cfg chinadns.Config
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		logrus.Fatalln(err)
	}

	logrus.Info("config:", cfg)

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.Fatalln(err)
	}

	logconfig.InitLogrus("chinadns", 10, level)

	copts := []chinadns.ClientOption{
		chinadns.WithUDPMaxBytes(cfg.UDPMaxBytes),
		chinadns.WithTimeout(time.Duration(cfg.Timeout) * time.Second),
	}

	sopts := []chinadns.ServerOption{
		chinadns.WithListenAddr(cfg.Listen),
		chinadns.WithDNS(cfg.DNSChina, cfg.DNSAbroad),
		chinadns.WithGFWFile(cfg.GFWPath),
	}

	client := chinadns.NewClient(copts...)
	server, err := chinadns.NewServer(client, sopts...)
	if err != nil {
		panic(err)
	}

	go func() {
		err := server.Run()
		if err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("quit,Got signal:", s)
}
