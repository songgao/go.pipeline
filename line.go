package pipeline

import (
	"bufio"
	"io"
)

type lineProcessorStation struct {
	in_stdout  <-chan string
	in_stderr  <-chan string
	out_stdout chan<- string
	out_stderr chan<- string

	stdoutProcessor LineProcessor
	stderrProcessor LineProcessor
}

func newLineProcessorStation(stdoutProcessor LineProcessor, stderrProcessor LineProcessor) *lineProcessorStation {
	var out, err LineProcessor
	if stdoutProcessor == nil {
		out = func(in string) string { return in }
	} else {
		out = stdoutProcessor
	}
	if stderrProcessor == nil {
		err = func(in string) string { return in }
	} else {
		err = stderrProcessor
	}
	return &lineProcessorStation{stdoutProcessor: out, stderrProcessor: err}
}

func (l *lineProcessorStation) SetStreams(inout <-chan string, inerr <-chan string, outout chan<- string, outerr chan<- string) {
	l.in_stdout = inout
	l.in_stderr = inerr
	l.out_stdout = outout
	l.out_stderr = outerr
}

func (l *lineProcessorStation) Start() {
	h := func(processor LineProcessor, in <-chan string, out chan<- string) {
		// tokens in "in" are arbitary strings. io.Pipe piles them together thus later can be read line by line
		reader, writer := io.Pipe()
		bufReader := bufio.NewReader(reader)

		// Read from in channel and write to the buffer
		go func() {
			ok := true
			var str string
			for ok {
				str, ok = <-in
				if ok {
					io.WriteString(writer, str)
				}
			}
			writer.CloseWithError(io.EOF)
		}()

		// Read from the buffer, process lines, and write to out channel
		go func() {
			var err error
			var line string
			for err == nil {
				line, err = bufReader.ReadString('\n')
				if err == nil {
					out <- processor(line)
				}
			}
			close(out)
		}()
	}

	h(l.stdoutProcessor, l.in_stdout, l.out_stdout)
	h(l.stderrProcessor, l.in_stderr, l.out_stderr)
}
