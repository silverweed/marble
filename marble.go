// Prunes old files under a directory.
// Activates whenever quota > maxquota
// and deletes file until quota < minquota.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"log"
	"os"
	"path"
	"sort"
)

var (
	minquota = flag.Int("minquota", 512, "Min disk quota in MB")
	maxquota = flag.Int("maxquota", 1024, "Max disk quota in MB")
	logfname = flag.String("logfile", "/var/log/marble.log", "Location of log file")
)

// A FileInfo + full path
type fileInfo struct {
	os.FileInfo
	AbsPath string
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Printf("Usage: %s <root directories>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	logfile, err := os.OpenFile(*logfname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Couldn't open log file: %s\n", err.Error())
	} else {
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	for _, root := range flag.Args() {
		pruneDirectory(root)
	}
}

func pruneDirectory(root string) {
	allfiles, byteTotsize, err := traverse(root)
	totsize := int(float64(byteTotsize) / 1024.0 / 1024.0)
	if err != nil {
		log.Printf("Error while traversing %s: %s\n", root, err.Error())
	}

	if totsize < *maxquota {
		log.Printf("[%s] Quota is below max allowed (%d / %d), exiting.\n", root, totsize, *maxquota)
		return
	}

	log.Printf("[%s] Quota above max allowed (%d / %d): pruning files...\n", root, totsize, *maxquota)

	initByAtime(len(allfiles))
	sort.Sort(ByAtime(allfiles))

	var (
		pruned     int
		bytespared int64
	)
	for _, f := range allfiles {
		if err := os.Remove(f.AbsPath); err != nil {
			log.Printf("[%s] Error deleting %s: %s\n", root, f.Name(), err.Error())
			continue
		}
		log.Printf("[%s] Deleted %s ... (%d bytes)\n", root, f.Name(), f.Size())
		pruned++
		bytespared += f.Size()
		if totsize-int(float64(bytespared)/1048576) <= *minquota {
			break
		}
	}
	log.Printf("[%s] Deleted %d files (total: %d kB)\n", root, pruned, int(float64(bytespared)/1024.0))
}

// traverse recursively lists all files and directories under `dir`
// and returns a flattened list of all files, their total size in bytes
// and an error (if occurred anywhere during the traversal)
func traverse(dir string) (files []fileInfo, size int64, err error) {
	all, e := ioutil.ReadDir(dir)
	if e != nil {
		err = e
		return
	}

	for _, file := range all {
		if file.IsDir() {
			f, s, e := traverse(path.Join(dir, file.Name()))
			if e != nil {
				err = e
				return
			}
			files = append(files, f...)
			size += s
		} else {
			files = append(files, fileInfo{file, path.Join(dir, file.Name())})
			size += file.Size()
		}
	}

	// Delete empty directories
	if len(files) == 0 {
		if err := os.Remove(dir); err == nil {
			log.Printf("Removed empty directory %s.\n", dir)
		} else {
			log.Printf("Error removing %s: %s\n", dir, err.Error())
		}
	}

	return
}
