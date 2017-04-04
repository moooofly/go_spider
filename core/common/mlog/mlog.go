// Package mlog implements log operations.
package mlog

import (
    "runtime"
)

// The plog is a public function combination for other log objects.
type plog struct {
    // a label represents whether open or not.
    isopen bool
}

// GetCaller returns file name and line number at the third step of runtime.
func (*plog) getCaller() (string, int) {
    _, file, line, ok := runtime.Caller(3)  // NOTE: is 3 a magic number?
    if !ok {
        file = "???"
        line = 0
    }
    return file, line
}

// Open makes log open.
func (this *plog) Open() {
    this.isopen = true
}

// Close makes log close.
func (this *plog) Close() {
    this.isopen = false
}
