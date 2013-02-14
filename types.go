package pipeline

// A station(node) in the pipeline that takes two streams of input, and produces two streams of output.
type Station interface {
	// Set 4 streams to the station: 1) stdout stream into the station; 2) stderr stream into the station; 3) stdout stream out of the station; 4) stderr stream out of the station.
	SetStreams(<-chan string, <-chan string, chan<- string, chan<- string)

	// Start the station and return
	Start()
}

// A string->string function to process the input line by line
type LineProcessor func(string) string

const (
	_UNITSIZE = 256
)
