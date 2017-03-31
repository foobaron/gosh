package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func splitProcess(line string) [][]string {
	var cmds [][]string
	for _, v := range strings.Split(line, "|") {
		cmds = append(cmds, strings.Fields(v))
	}
	return cmds
}

func execPipe(line string) (output []byte, err error) {
	commands := splitProcess(line)
	cmds := make([]*exec.Cmd, len(commands))
	if len(commands) == 1 {
		switch commands[0][0] {
		case "exit":
			os.Exit(0)
		case "cd":
			var dir string
			if len(commands[0]) > 1 {
				dir = strings.Replace(commands[0][1], "~", os.Getenv("HOME"), 1)
			} else {
				dir = os.Getenv("HOME")
			}
			err = os.Chdir(dir)
			return nil, fmt.Errorf("Chdir() in execPipe(): %v", err)
		}
	}
	for i, c := range commands {
		cmds[i] = exec.Command(c[0], c[1:]...)
		if i > 0 {
			if cmds[i].Stdin, err = cmds[i-1].StdoutPipe(); err != nil {
				return nil, fmt.Errorf("StdoutPipe() in execPipe(): %v", err)
			}
		}
		cmds[i].Stderr = os.Stderr
	}
	var out bytes.Buffer
	cmds[len(cmds)-1].Stdout = &out
	for _, c := range cmds {
		if err = c.Start(); err != nil {
			return nil, fmt.Errorf("Start() in execPipe(): %v", err)
		}
	}
	for _, c := range cmds {
		if err = c.Wait(); err != nil {
			return nil, fmt.Errorf("Wait() in execPipe(): %v", err)
		}
	}
	return out.Bytes(), nil
}

func main() {
	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "$ ")
		if s.Scan() {
			if s.Text() != "" {
				out, err := execPipe(s.Text())
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
				}
				fmt.Print(string(out))
			}
		}
		if s.Err() != nil {
			log.Fatal(s.Err())
		}
	}
}
