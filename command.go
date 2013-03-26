package pipeline

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type command struct {
	in_stdout  <-chan string
	in_stderr  <-chan string
	out_stdout chan<- string
	out_stderr chan<- string
	err        *error

	cmd *exec.Cmd
}

func newCommand(name string, err *error, args ...interface{}) Station {
	c := new(command)
	c.err = err
	if len(args) == 0 {
		c.cmd = exec.Command(name)
	} else {
		arg := make([]string, len(args))
		for i := 0; i < len(args); i++ {
			arg[i] = fmt.Sprint(args[i])
		}
		c.cmd = exec.Command(name, arg...)
	}
	return c
}

func (c *command) SetStreams(inout <-chan string, inerr <-chan string, outout chan<- string, outerr chan<- string) {
	c.in_stdout = inout
	c.in_stderr = inerr
	c.out_stdout = outout
	c.out_stderr = outerr
}

func (c *command) Start() {
	stdin, err := c.cmd.StdinPipe()
	stdout, err := c.cmd.StdoutPipe()
	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		panic("Failed to open one of pipes")
	}
	bufout := bufio.NewReader(stdout)
	buferr := bufio.NewReader(stderr)

	var wg sync.WaitGroup
	wg.Add(3) // two routines writting to out_stderr; and exit code watcher

	// Read from stdout in input, and write to the stdin of cmd
	go func() {
		ok := true
		var str string
		for ok {
			str, ok = <-c.in_stdout
			if ok {
				io.WriteString(stdin, str)
			}
		}
		stdin.Close()
	}()

	// Read from stderr in input, and directly write to stderr in output
	go func() {
		ok := true
		var str string
		for ok {
			str, ok = <-c.in_stderr
			if ok {
				c.out_stderr <- str
			}
		}
		// not the only one writting to out_stderr; need to wait before close the channel
		wg.Done()
	}()

	// Read from stdout from cmd, and write to stdout in output
	go func() {
		var err error
		str := make([]byte, _UNITSIZE)
		var nread int
		for err == nil {
			nread, err = bufout.Read(str)
			if err == nil || err == io.EOF {
				c.out_stdout <- string(str[:nread])
			}
		}
		close(c.out_stdout)
	}()

	// Read from stderr from cmd, and write to stderr in output
	go func() {
		var err error
		str := make([]byte, _UNITSIZE)
		var nread int
		for err == nil {
			nread, err = buferr.Read(str)
			if err == nil || err == io.EOF {
				c.out_stderr <- string(str[:nread])
			}
		}
		// not the only one writting to out_stderr; need to wait before close the channel
		wg.Done()
	}()

	// Wait for goroutines that are writting to out_stderr to finish, and close the channel
	go func() {
		wg.Wait()
		close(c.out_stderr)
	}()

	c.cmd.Start()

	go func() {
		err := c.cmd.Wait()
		if err != nil {
			(*c.err) = err
		}
		wg.Done()
	}()
}
