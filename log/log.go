// Copyright 2014 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"bytes"

	"bitbucket.org/tshannon/freehold/data"
	"bitbucket.org/tshannon/freehold/fail"
	"bitbucket.org/tshannon/freehold/setting"
)

const (
	// DS is the location of the log dS file
	DS = "core" + string(os.PathSeparator) + "log.ds"
	// AuthType is the log type for authentication log entries
	AuthType = "authentication"
	// Four04Type is the log type for 404 errors
	Four04Type = "404"
)

// Log is the storage stucture for a log entry
type Log struct {
	When string `json:"when"`
	Type string `json:"type"`
	Log  string `json:"log"`
}

// Iter is an iterator for log entries
type Iter struct {
	Type  string           `json:"type,omitempty"`
	From  *json.RawMessage `json:"from,omitempty"`
	To    *json.RawMessage `json:"to,omitempty"`
	Skip  int              `json:"skip,omitempty"`
	Limit int              `json:"limit,omitempty"`
	Order string           `json:"order,omitempty"`
}

// NewEntry inserts a new log entry into the datastore
func NewEntry(Type string, entry string) {
	ds, err := data.OpenCoreDS(DS)
	if err != nil {
		SyslogError(errors.New("Error can't log entry to freehold instance. Entry: " +
			entry + " error: " + err.Error()))
		return

	}
	when := time.Now().Format(time.RFC3339)

	log := &Log{
		When: when,
		Type: Type,
		Log:  entry,
	}

	err = ds.Put(when+"_"+Type, log)

	if err != nil {
		SyslogError(errors.New("Error can't log entry to freehold instance. Entry: " +
			entry + " error: " + err.Error()))
		return
	}

}

// Get retrieves a list of logs from the datastore
func Get(iter *Iter) ([]*Log, error) {
	ds, err := data.OpenCoreDS(DS)
	if err != nil {
		return nil, err
	}

	//TODO: I don't like that this iter is an exact copy of the data.Iter

	if iter.Order != "asc" && iter.Order != "dsc" && iter.Order != "" {
		return nil, fail.New("Invalid Order specified", iter)
	}

	var from []byte
	var to []byte

	if iter.From != nil {
		from = []byte(*iter.From)
	} else {
		from, err = ds.Min()
		if err != nil {
			return nil, err
		}
	}
	if iter.To != nil {
		to = []byte(*iter.To)
	} else {
		to, err = ds.Max()
		if err != nil {
			return nil, err
		}

	}

	if iter.Order != "" {
		if iter.Order != "asc" && iter.Order != "dsc" {
			return nil, fail.New("Invalid Order specified", iter)
		}
		//Flip from and to if order is specified.  If order isn't specified, then iteration direction
		// is implied
		if (iter.Order == "dsc" && bytes.Compare(from, to) != 1) ||
			(iter.Order == "asc" && bytes.Compare(to, from) == -1) {
			from, to = to, from
		}
	}

	i, err := ds.Iter(from, to)
	defer i.Close()
	if err != nil {
		return nil, err
	}

	skip := 0
	limit := 0
	result := make([]*Log, 0, iter.Limit)

	for i.Next() {
		if iter.Limit > 0 && iter.Limit <= limit {
			break
		}
		if i.Err() != nil {
			return nil, i.Err()
		}
		value := i.Value()
		if value == nil {
			//shouldn't happen
			panic("Nil value returned from datastore iterator")
		}

		entry := &Log{}
		err = json.Unmarshal(value, entry)
		if err != nil {
			return nil, err
		}

		if iter.Type != "" && entry.Type != iter.Type {
			continue
		}

		skip++
		if iter.Skip > 0 && iter.Skip >= skip {
			continue
		}

		result = append(result, entry)

		limit++
	}

	return result, nil
}

// Error logs and error to the core log datastore.  For core code the rule for error logging
// is to not log it until it's "bubbled up to the top".  Meaning only the http handler
// should log the error.  This is to prevent the same error from being logged a bunch of times.
func Error(err error) {
	if !setting.Bool("LogErrors") {
		return
	}
	if setting.Bool("LogErrorsToSyslog") {
		SyslogError(err)
	}
	NewEntry("error", err.Error())
}

// Fail logs a failure in the datastore
func Fail(err *fail.Fail, who string) {
	if !setting.Bool("LogFailures") {
		return
	}

	data, jerr := json.Marshal(err.Data)
	if jerr != nil {
		Error(jerr)
		return
	}

	entry := ""
	if who != "" {
		entry += "Who: " + who + ", "
	}
	entry += fmt.Sprintf("Message: %s  Data: %s", err.Message, data)

	NewEntry("failure", entry)
}

type fhLogWriter struct{}

func (w *fhLogWriter) Write(p []byte) (n int, err error) {
	if setting.Bool("LogServerErrors") {
		NewEntry("server error", string(p))
	}
	return len(p), nil
}

// FHLogger returns a logger for use with the http server
func FHLogger() *log.Logger {
	return log.New(&fhLogWriter{}, "", 0)
}
