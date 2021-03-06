package log

import (
	"io"

	"github.com/op/go-logging"
)

// Variabili relative ai formati default di log
var (
	DefaultLogFormatter         = logging.MustStringFormatter("%{color}%{time:2006-01-02 15:04} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}")
	LowVerboseLogFormatter      = logging.MustStringFormatter("%{time:2006-01-02 15:04} ▶ %{level:.4s} %{message}")
	VerboseLogFilePathFormatter = logging.MustStringFormatter("%{color}%{time:2006-01-02 15:04} %{longpkg} %{longfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{longfile} %{message}")
)

var backendList []logging.Backend

var logger *logging.Logger

// setBackendFormat - Imposta il formato dell'output di logging
func setBackendFormat(backend *logging.LogBackend, format logging.Formatter) logging.Backend {

	var b logging.Backend

	if format != nil {
		b = logging.NewBackendFormatter(backend, format)
	} else {
		b = logging.NewBackendFormatter(backend, DefaultLogFormatter)
	}

	return b
}

// GetLogger - Restituisce il logger
func GetLogger() *logging.Logger {
	return logger
}

// Init - Si occupa di inizializzare il logging
func Init(appName string) {

	logger = logging.MustGetLogger(appName)

	logging.SetBackend(backendList...)
}

// NewLogBackend - Richiama init per un nuovo backend di logging
func NewLogBackend(out io.Writer, prefix string, flag int, level logging.Level, format logging.Formatter) {

	backend := logging.NewLogBackend(out, prefix, flag)

	b := setBackendFormat(backend, format)

	backendLevel := logging.AddModuleLevel(b)
	backendLevel.SetLevel(level, "")

	backendList = append(backendList, backendLevel)
}
