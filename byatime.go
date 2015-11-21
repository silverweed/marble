package main

import (
	"log"
	"syscall"
)

// Utility to sort by atime
type ByAtime []fileInfo

func (f ByAtime) Len() int {
	return len(f)
}

func (f ByAtime) Less(i, j int) bool {
	var st1, st2 syscall.Stat_t
	if err := syscall.Stat(f[i].AbsPath, &st1); err != nil {
		log.Fatal("Error on stat(" + f[i].AbsPath + "): " + err.Error())
	}
	if err := syscall.Stat(f[j].AbsPath, &st2); err != nil {
		log.Fatal("Error on stat(" + f[j].AbsPath + "): " + err.Error())
	}
	return st1.Atim.Sec < st2.Atim.Sec
}
func (f ByAtime) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
