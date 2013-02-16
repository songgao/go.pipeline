package pipeline

import (
	"bufio"
	"io"
	"os"
	"sync"
)

// A pipeline consisting of stations that takes two input streams and produce two output streams
type Pipeline struct {
	stdout    <-chan string
	stderr    <-chan string
	chainable bool
}

// Create a new pipeline that has not input streams
func NewPipeline() *Pipeline {
	ch1, ch2 := make(chan string), make(chan string)
	p := &Pipeline{stdout: ch1, stderr: ch2, chainable: true}
	close(ch1)
	close(ch2)
	return p
}

// Create a new pipeline with input streams stdout and stderr
func StartPipelineWithStreams(stdout <-chan string, stderr <-chan string) *Pipeline {
	return &Pipeline{stdout: stdout, stderr: stderr, chainable: true}
}

// Create a new pipeline whose input streams are from a command output
func StartPipelineWithCommand(name string, arg ...interface{}) *Pipeline {
	out, err := make(chan string, 4), make(chan string, 4)
	p := &Pipeline{stdout: out, stderr: err, chainable: true}
	ch1, ch2 := make(chan string), make(chan string)
	var station Station
	if len(arg) == 0 {
		station = newCommand(name)
	} else {
		station = newCommand(name, arg...)
	}
	station.SetStreams(ch1, ch2, out, err)
	close(ch1)
	close(ch2)
	station.Start()
	return p
}

// Create a new pipeline whose input streams are /dev/stdin and /dev/stdout
func StartPipelineWithStdin() *Pipeline {
	out, err := make(chan string, 4), make(chan string, 4)
	p := &Pipeline{stdout: out, stderr: err, chainable: true}
	close(err)
	go func() {
		bufin := bufio.NewReader(os.Stdin)
		var er error
		str := make([]byte, _UNITSIZE)
		var nread int
		for er == nil {
			nread, er = bufin.Read(str)
			if er == nil || er == io.EOF {
				out <- string(str[:nread])
			}
		}
		close(out)
	}()
	return p
}

// Add a command station at the end of the pipeline, i.e., start a command, chain the output streams of current pipeline to the input streams of a the command, and let the output stremas of the command be the new output streams of the pipeline
func (p *Pipeline) ChainCommand(name string, arg ...interface{}) *Pipeline {
	if !p.chainable {
		panic("pipeline is not chainable")
	}
	out, err := make(chan string, 4), make(chan string, 4)
	r := &Pipeline{stdout: out, stderr: err, chainable: true}
	var station Station
	if len(arg) == 0 {
		station = newCommand(name)
	} else {
		station = newCommand(name, arg...)
	}
	station.SetStreams(p.stdout, p.stderr, out, err)
	station.Start()
	return r
}

// Add a line processing station at the end of the pipeline. The line processing station process the input streams line by line. The output stremas of the line processing station will be the new output streams of the pipeline.
func (p *Pipeline) ChainLineProcessor(stdoutProcessor LineProcessor, stderrProcessor LineProcessor) *Pipeline {
	if !p.chainable {
		panic("pipeline is not chainable")
	}
	out, err := make(chan string, 4), make(chan string, 4)
	r := &Pipeline{stdout: out, stderr: err, chainable: true}
	station := newLineProcessorStation(stdoutProcessor, stderrProcessor)
	station.SetStreams(p.stdout, p.stderr, out, err)
	station.Start()
	return r
}

// Seal the pipeline (prevent from chaining new stations into the pipeline) and print current output streams of the pipeline to /dev/stdout and /dev/stderr respectively
func (p *Pipeline) PrintAll() {
	if !p.chainable {
		panic("pipline is not chainable")
	}
	p.chainable = false
	var wg sync.WaitGroup
	h := func(ch_in <-chan string, out *os.File) {
		ok := true
		var str string
		for ok {
			str, ok = <-ch_in
			if ok {
				out.WriteString(str)
			}
		}
		out.Close()
		wg.Done()
	}
	wg.Add(2)
	go h(p.stdout, os.Stdout)
	go h(p.stderr, os.Stderr)
	wg.Wait()
}

// Seal the pipeline (prevent from chaining new stations into the pipeline) and return the two output streams
func (p *Pipeline) GetOutputStreams() (stdout <-chan string, stderr <-chan string) {
	stdout = p.stdout
	stderr = p.stderr
	p.chainable = false
	return
}

// Short for ChainCommand()
func (p *Pipeline) C(name string, arg ...interface{}) *Pipeline {
	var r *Pipeline
	if len(arg) == 0 {
		r = p.ChainCommand(name)
	} else {
		r = p.ChainCommand(name, arg...)
	}
	return r
}

// Short for ChainLineProcessor()
func (p *Pipeline) L(stdoutProcessor LineProcessor, stderrProcessor LineProcessor) *Pipeline {
	return p.ChainLineProcessor(stdoutProcessor, stderrProcessor)
}

// Short for PrintAll()
func (p *Pipeline) P() {
	p.PrintAll()
}
