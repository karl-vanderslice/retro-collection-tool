package app

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type commandSpinner struct {
	prefix string
	phase  string

	mu     sync.Mutex
	status string
	active bool

	frames []string
	idx    int
	done   chan struct{}
	silent bool
}

func newCommandSpinner(g globalFlags, prefix, phase, status string) *commandSpinner {
	s := &commandSpinner{
		prefix: prefix,
		phase:  phase,
		status: strings.TrimSpace(status),
		frames: []string{"-", "\\", "|", "/"},
		done:   make(chan struct{}),
		silent: g.isJSONOutput(),
	}
	if s.silent {
		return s
	}

	s.active = stdoutIsTerminal()
	if !s.active {
		fmt.Printf("[%s] [%s] %s\n", s.prefix, s.phase, s.status)
		return s
	}

	go s.run()
	return s
}

func (s *commandSpinner) run() {
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			frame := s.frames[s.idx%len(s.frames)]
			s.idx++
			status := truncateStatus(s.status, 96)
			s.mu.Unlock()
			fmt.Printf("\r\033[2K[%s] [%s] %s %s", s.prefix, s.phase, frame, status)
		}
	}
}

func (s *commandSpinner) Update(status string) {
	s.mu.Lock()
	s.status = strings.TrimSpace(status)
	s.mu.Unlock()
}

func (s *commandSpinner) Stop(ok bool, final string) {
	if s.silent {
		return
	}
	if s.active {
		close(s.done)
		label := "done"
		if !ok {
			label = "fail"
		}
		fmt.Printf("\r\033[2K[%s] [%s] %s %s\n", s.prefix, s.phase, label, truncateStatus(strings.TrimSpace(final), 120))
		return
	}

	if ok {
		fmt.Printf("[%s] [%s] done %s\n", s.prefix, s.phase, strings.TrimSpace(final))
		return
	}
	fmt.Printf("[%s] [%s] fail %s\n", s.prefix, s.phase, strings.TrimSpace(final))
}

func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func truncateStatus(status string, max int) string {
	status = strings.TrimSpace(status)
	if max <= 0 || len(status) <= max {
		return status
	}
	if max <= 3 {
		return status[:max]
	}
	return status[:max-3] + "..."
}
