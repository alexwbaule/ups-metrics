package graylog

import (
	"github.com/alexwbaule/ups-metrics/internal/domain/port"
)

// Compile-time check to ensure Gelf implements LogWriter interface
var _ port.LogWriter = (*Gelf)(nil)
