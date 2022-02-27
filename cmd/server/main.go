package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/0990/chinadns"
	"github.com/0990/chinadns/logconfig"
	"github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

var confFile = flag.String("c", "chinadns.json", "config file")
var workingDir = flag.String("w", "", "working dir")

func main() {
	flag.Parse()

	cfgFile := *confFile
	if *workingDir != "" {
		cfgFile = *workingDir + "/" + cfgFile
	}

	file, err := os.Open(cfgFile)
	if err != nil {
		logrus.Fatalln(err)
	}

	var cfg chinadns.Config
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		logrus.Fatalln(err)
	}

	var logName = "chinadns"

	if *workingDir != "" {
		cfg.ChnDomain = filepath.Join(*workingDir, cfg.ChnDomain)
		cfg.GfwDomain = filepath.Join(*workingDir, cfg.GfwDomain)
		for i, v := range cfg.ChnIP {
			cfg.ChnIP[i] = filepath.Join(*workingDir, v)
		}
		logName = filepath.Join(*workingDir, logName)
	}

	logrus.Info("config:", cfg)

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.Fatalln(err)
	}

	go func() {
		if cfg.PProfPort > 0 {
			err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.PProfPort), nil)
			if err != nil {
				logrus.Fatalln(err)
			}
		}
	}()

	logconfig.InitLogrus(logName, 10, level)

	copts := []chinadns.ClientOption{
		chinadns.WithUDPMaxBytes(cfg.UDPMaxBytes),
		chinadns.WithTimeout(time.Duration(cfg.Timeout) * time.Second),
	}

	sopts := []chinadns.ServerOption{
		chinadns.WithListenAddr(cfg.Listen),
		chinadns.WithCacheExpireSec(cfg.CacheExpireSec),
		chinadns.WithDNS(cfg.DNSChina, cfg.DNSAbroad),
		chinadns.WithDomain2IP(cfg.Domain2IP),
		chinadns.WithCHNFile(cfg.ChnIP),
		chinadns.WithChnDomain(cfg.ChnDomain),
		chinadns.WithGfwDomain(cfg.GfwDomain),
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
