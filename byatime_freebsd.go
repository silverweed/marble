// +build freebsd
package main

import (
	"log"
	"syscall"
)

var memo []syscall.Stat_t

func initByAtime(length int) {
	memo = make([]syscall.Stat_t, length)
}

// Utility to sort by atime
type ByAtime []fileInfo

func (f ByAtime) Len() int {
	return len(f)
}

// Less performs a memoized comparison between Atimes of files
func (f ByAtime) Less(i, j int) bool {
	st1 := memo[i]
	if st1.Atimespec.Sec == 0 {
		if err := syscall.Stat(f[i].AbsPath, &st1); err != nil {
			log.Fatal("Error on stat(" + f[i].AbsPath + "): " + err.Error())
		} else {
			memo[i] = st1
		}
	}
	st2 := memo[j]
	if st2.Atimespec.Sec == 0 {
		if err := syscall.Stat(f[j].AbsPath, &st2); err != nil {
			log.Fatal("Error on stat(" + f[j].AbsPath + "): " + err.Error())
		} else {
			memo[j] = st2
		}
	}
	return st1.Atimespec.Sec < st2.Atimespec.Sec
}
func (f ByAtime) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
