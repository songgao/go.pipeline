package pipeline

import (
	"sync"
)

type redirectStation struct {
	in_stdout  <-chan string
	in_stderr  <-chan string
	out_stdout chan<- string
	out_stderr chan<- string

	to int
}

func newRedirectStation(to int) *redirectStation {
	if to != STDOUT && to != STDERR {
		panic("Can only redirect to STDOUT(1) or STDERR(2)")
	}

	return &redirectStation{to: to}
}

func (s *redirectStation) SetStreams(inout <-chan string, inerr <-chan string, outout chan<- string, outerr chan<- string) {
	s.in_stdout = inout
	s.in_stderr = inerr
	s.out_stdout = outout
	s.out_stderr = outerr
}

func (s *redirectStation) Start() {
	var wg sync.WaitGroup

	h := func(ch_in <-chan string, ch_out chan<- string) {
		ok := true
		var str string
		for ok {
			str, ok = <-ch_in
			if ok {
				ch_out <- str
			}
		}
		wg.Done()
	}
	c := func(ch chan<- string) {
		wg.Wait()
		close(ch)
	}

	wg.Add(2)
	if s.to == STDOUT {
		close(s.out_stderr)
		go h(s.in_stdout, s.out_stdout)
		go h(s.in_stderr, s.out_stdout)
		go c(s.out_stdout)
	} else {
		close(s.out_stdout)
		go h(s.in_stdout, s.out_stderr)
		go h(s.in_stderr, s.out_stderr)
		go c(s.out_stderr)
	}
}
