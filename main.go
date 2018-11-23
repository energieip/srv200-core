package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/energieip/srv200-coreservice-go/internal/service"
)

func main() {
	var confFile string

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.StringVar(&confFile, "config", "", "Specify an alternate configuration file.")
	flag.StringVar(&confFile, "c", "", "Specify an alternate configuration file.")
	flag.Parse()

	service := service.CoreService{}
	err := service.Initialize(confFile)
	if err != nil {
		log.Println("Error during service connexion " + err.Error())
		os.Exit(1)
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Received SIGTERM")
		service.Stop()
		os.Exit(0)
	}()

	err = service.Run()
	if err != nil {
		log.Println("Error during service execution " + err.Error())
		os.Exit(1)
	}
}
