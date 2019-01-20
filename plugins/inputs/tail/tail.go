// +build !solaris

package tail

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/influxdata/tail"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/internal/globpath"
	"github.com/lavaorg/telex/plugins/inputs"
	"github.com/lavaorg/telex/plugins/parsers"
)

const (
	defaultWatchMethod = "inotify"
)

type Tail struct {
	Files         []string
	FromBeginning bool
	Pipe          bool
	WatchMethod   string

	tailers    map[string]*tail.Tail
	parserFunc parsers.ParserFunc
	wg         sync.WaitGroup
	acc        telex.Accumulator

	sync.Mutex
}

func NewTail() *Tail {
	return &Tail{
		FromBeginning: false,
	}
}

func (t *Tail) Gather(acc telex.Accumulator) error {
	t.Lock()
	defer t.Unlock()

	return t.tailNewFiles(true)
}

func (t *Tail) Start(acc telex.Accumulator) error {
	t.Lock()
	defer t.Unlock()

	t.acc = acc
	t.tailers = make(map[string]*tail.Tail)

	return t.tailNewFiles(t.FromBeginning)
}

func (t *Tail) tailNewFiles(fromBeginning bool) error {
	var seek *tail.SeekInfo
	if !t.Pipe && !fromBeginning {
		seek = &tail.SeekInfo{
			Whence: 2,
			Offset: 0,
		}
	}

	var poll bool
	if t.WatchMethod == "poll" {
		poll = true
	}

	// Create a "tailer" for each file
	for _, filepath := range t.Files {
		g, err := globpath.Compile(filepath)
		if err != nil {
			t.acc.AddError(fmt.Errorf("E! Error Glob %s failed to compile, %s", filepath, err))
		}
		for _, file := range g.Match() {
			if _, ok := t.tailers[file]; ok {
				// we're already tailing this file
				continue
			}

			tailer, err := tail.TailFile(file,
				tail.Config{
					ReOpen:    true,
					Follow:    true,
					Location:  seek,
					MustExist: true,
					Poll:      poll,
					Pipe:      t.Pipe,
					Logger:    tail.DiscardingLogger,
				})
			if err != nil {
				t.acc.AddError(err)
				continue
			}

			log.Printf("D! [inputs.tail] tail added for file: %v", file)

			parser, err := t.parserFunc()
			if err != nil {
				t.acc.AddError(fmt.Errorf("error creating parser: %v", err))
			}

			// create a goroutine for each "tailer"
			t.wg.Add(1)
			go t.receiver(parser, tailer)
			t.tailers[tailer.Filename] = tailer
		}
	}
	return nil
}

// this is launched as a goroutine to continuously watch a tailed logfile
// for changes, parse any incoming msgs, and add to the accumulator.
func (t *Tail) receiver(parser parsers.Parser, tailer *tail.Tail) {
	defer t.wg.Done()

	var firstLine = true
	var metrics []telex.Metric
	var m telex.Metric
	var err error
	var line *tail.Line
	for line = range tailer.Lines {
		if line.Err != nil {
			t.acc.AddError(fmt.Errorf("E! Error tailing file %s, Error: %s\n",
				tailer.Filename, err))
			continue
		}
		// Fix up files with Windows line endings.
		text := strings.TrimRight(line.Text, "\r")

		if firstLine {
			metrics, err = parser.Parse([]byte(text))
			if err == nil {
				if len(metrics) == 0 {
					firstLine = false
					continue
				} else {
					m = metrics[0]
				}
			}
			firstLine = false
		} else {
			m, err = parser.ParseLine(text)
		}

		if err == nil {
			if m != nil {
				tags := m.Tags()
				tags["path"] = tailer.Filename
				t.acc.AddFields(m.Name(), m.Fields(), tags, m.Time())
			}
		} else {
			t.acc.AddError(fmt.Errorf("E! Malformed log line in %s: [%s], Error: %s\n",
				tailer.Filename, line.Text, err))
		}
	}

	log.Printf("D! [inputs.tail] tail removed for file: %v", tailer.Filename)

	if err := tailer.Err(); err != nil {
		t.acc.AddError(fmt.Errorf("E! Error tailing file %s, Error: %s\n",
			tailer.Filename, err))
	}
}

func (t *Tail) Stop() {
	t.Lock()
	defer t.Unlock()

	for _, tailer := range t.tailers {
		err := tailer.Stop()
		if err != nil {
			t.acc.AddError(fmt.Errorf("E! Error stopping tail on file %s\n", tailer.Filename))
		}
	}

	for _, tailer := range t.tailers {
		tailer.Cleanup()
	}
	t.wg.Wait()
}

func (t *Tail) SetParserFunc(fn parsers.ParserFunc) {
	t.parserFunc = fn
}

func init() {
	inputs.Add("tail", func() telex.Input {
		return NewTail()
	})
}
