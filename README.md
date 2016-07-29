# Marble
Marble is a simple program which deletes old files under a list of given directories if the total size of that directory exceeds a given one.
It is designed to be run as a periodic cronjob in a non-interactive fashion.

## Command line
Usage:  

```./marble [flags] <root directories>```

Flags are:

```
-minquota=512 (Min disk quota in MB)
-maxquota=1024 (Max disk quota in MB)
-logfname="/var/log/marble.log" (Location of log file)
```

When launched like `./marble /foo`, marble will scan the directory `/foo` and:  
* if the total size of the files under `/foo` exceeds `maxquota`, it'll remove files until the total size is under `minquota`;  
* else, return without doing anything.

The files will be deleted starting from the ones with the oldest access time.

## Compatibility
Marble currently compiles and works under Linux and BSD.

## LICENSE
Marble is under the MIT license (see LICENSE).
