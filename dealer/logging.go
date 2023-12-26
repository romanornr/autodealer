package dealer

import (
	"github.com/romanornr/autodealer/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/thrasher-corp/gocryptotrader/common/convert"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
	"strings"
	"sync/atomic"
	"time"
)

// +----------+
// | LogState |
// +----------+

const (
	// Dormant mode: the AwakenLogger outputs trace logs as info
	// every once in a while.
	Dormant = iota
	// Awaken mode: everything is outputted as-is.
	Awaken = iota
)

// LogState keeps track of what the current state is.
type LogState struct {
	state int32
	// For how long state should be kept Awaken.
	duration time.Duration
}

// NewLogState creates a new LogState.
func NewLogState(duration time.Duration) LogState {
	return LogState{
		state:    Dormant,
		duration: duration,
	}
}

// WakeUp sets the state to Awaken.
func (s *LogState) WakeUp() {
	if atomic.CompareAndSwapInt32(&s.state, Dormant, Awaken) {
		time.AfterFunc(s.duration, func() {
			if !atomic.CompareAndSwapInt32(&s.state, Awaken, Dormant) {
				panic("illegal state")
			}
		})
	}
}

// Awaken returns true if the state is Awaken.
func (s *LogState) Awaken() bool {
	return atomic.LoadInt32(&s.state) == Awaken
}

// +--------------+
// | AwakenLogger |
// +--------------+

// AwakenLogger is a zerolog logger that outputs trace logs as info
type AwakenLogger struct {
	state LogState

	traceEvery time.Duration
	traceLast  time.Time
}

// NewAwakenLogger creates a new AwakenLogger.
func NewAwakenLogger(d time.Duration) AwakenLogger {
	// We use the same duration for both the time awaken and how
	// often trace logs should be allowed.
	return AwakenLogger{
		state:      NewLogState(d),
		traceEvery: d,
		traceLast:  time.Time{},
	}
}

// WakeUp gets the AwakenLogger out of its dormant state for a an
// amount of time.
func (t *AwakenLogger) WakeUp() {
	t.state.WakeUp()
}

// Trace logs a trace message.
func (t *AwakenLogger) Trace() *zerolog.Event {
	if t.state.Awaken() {
		return log.Info()
	}

	return t.dormantTrace()
}

// dormantTrace is left unlocked on purpose.  If there is a race
// condition and more than one thread set `traceLast` to Now(), we
// don't care, it's still Now().
func (t *AwakenLogger) dormantTrace() *zerolog.Event {
	if time.Since(t.traceLast) > t.traceEvery {
		t.traceLast = time.Now()

		return log.Info()
	}

	return log.Trace()
}

// +-------------------+
// | Stateless logging |
// +-------------------+

// Code is a helper function to log a code.
func Code(e *zerolog.Event, code string) {
	if code != "" {
		e = e.Str("code", code)
	}

	e.Msg(util.Location2())
}

// What is a helper function to log a what.
func What(e *zerolog.Event, what string) {
	if what != "" {
		e = e.Str("what", what)
	}

	e.Msg(util.Location2())
}

// Msg is a helper function to log a message.
func Msg(e *zerolog.Event) {
	e.Msg(util.Location2())
}

// +-------------------+
// | GCT logging |
// +-------------------+.

// setupGCTLogging sets up the GCT logger.
func (d *Dealer) setupGCTLogging() {
	d.Config.Logging.AdvancedSettings.ShowLogSystemName = convert.BoolPtr(false)
	d.Config.Logging.AdvancedSettings.Headers.Info = "i"
	d.Config.Logging.AdvancedSettings.Headers.Warn = "w"
	d.Config.Logging.AdvancedSettings.Headers.Debug = "d"
	d.Config.Logging.AdvancedSettings.Headers.Error = "e"
	d.Config.Logging.SubLoggerConfig.Output = "stdout"

	gctlog.SetGlobalLogConfig(&d.Config.Logging)
	gctlog.SetupGlobalLogger("autodealer", true)

	// TODO: formatting per or GCTConsoleWriter
	//var console GCTConsoleWriter
	//
	//// override all sublogger outputs with our own writer
	//for _, subLogger := range gctlog.SubLoggers {
	//	subLogger.SetOutput(console)
	//}
}

// GCTConsoleWriter is a zerolog writer that outputs to the console.
type GCTConsoleWriter struct{}

// Write implements the Writer interface.
func (c GCTConsoleWriter) Write(p []byte) (n int, err error) {
	var l *zerolog.Event

	// look at the first byte of the data, it will be the header
	// we defined at setupGCTLogging, use that to determine the
	// log level to apply
	switch p[0] {
	case 'i':
		l = log.Info()
	case 'w':
		l = log.Warn()
	case 'd':
		l = log.Debug()
	case 'e':
		l = log.Error()
	default:
		l = log.Debug()
	}

	l.Msg(strings.TrimSuffix(string(p[1:]), "\n"))

	return len(p), nil
}
