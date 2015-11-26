// Prunes old files under a directory.
// Activates whenever quota > maxquota
// and deletes file until quota < minquota.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
)

var (
	root     = flag.String("root", "/cache/", "Cache root directory")
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

	logfile, err := os.OpenFile(*logfname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Couldn't open log file: %s\n", err.Error())
	} else {
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	allfiles, totsize, err := traverse(*root)
	if totsize < *maxquota {
		log.Printf("Quota is below max allowed (%d / %d), exiting.\n", totsize, *maxquota)
		return
	}

	log.Printf("Quota above max allowed (%d / %d): pruning files...\n", totsize, *maxquota)

	initByAtime(len(allfiles))
	sort.Sort(ByAtime(allfiles))
	var (
		pruned     int
		bytespared int64
	)
	for _, f := range allfiles {
		if err := os.Remove(f.AbsPath); err != nil {
			log.Printf("Error deleting %s: %s\n", f.Name(), err.Error())
			continue
		}
		log.Printf("Deleted %s ... (%d bytes)\n", f.Name(), f.Size())
		pruned++
		bytespared += f.Size()
		removeIfEmpty(path.Dir(f.AbsPath))
		if totsize-int(bytespared/1024) <= *minquota {
			break
		}
	}
	log.Printf("Deleted %d files (total: %d kB)\n", pruned, bytespared)
}

// traverse recursively lists all files and directories under `dir`
// and returns a flattened list of all files, their total size in MB
// and an error (if occurred anywhere during the traversal)
func traverse(dir string) (files []fileInfo, size int, err error) {
	all, err := ioutil.ReadDir(dir)
	if err != nil {
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
			size += int(file.Size() / 1024 / 1024)
		}
	}
	return
}

// removeIfEmpty checks if directory is empty, and deletes it if so;
// it also calls itself recursively on the parent directory if the
// inmost dir was deleted.
func removeIfEmpty(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading dir %s: %s\n", dir, err.Error())
		return
	}
	if len(files) == 0 {
		os.Remove(dir)
		log.Printf("Removed empty directory %s.\n", dir)
		removeIfEmpty(path.Dir(dir))
	}
}
