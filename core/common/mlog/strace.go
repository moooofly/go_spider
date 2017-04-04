package mlog

import (
    "log"
    "os"
)

// Strace represents an active object that strace the processing of spider.
// The strace info is output to os.Stderr.
// The loginst is an pointer of logger in Std-Packages.
// The isopen is a label represents whether open strace or not.
type strace struct {
    plog

    loginst *log.Logger
}

// 未显式初始化，则初始化为对应类型对零值
var pstrace *strace

// 获取 strace 对象（单例模式）
func StraceInst() *strace {
    if pstrace == nil { // NOTE: nil comparison right?
        pstrace = newStrace()
    }
    return pstrace
}

// The newStrace returns initialized strace object.
func newStrace() *strace {
    pstrace := &strace{}
    pstrace.loginst = log.New(os.Stderr, "", log.LstdFlags)
    pstrace.isopen = true
    return pstrace
}

// Println output the str to os.Stderr.
func (this *strace) Println(str string) {
    if !this.isopen {
        return
    }
    this.loginst.Printf("%s\n", str)
}
