/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : gfind.go

* Purpose :

* Creation Date : 03-24-2014

* Last Modified : Wed 02 Apr 2014 12:43:05 AM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package main

import (
	"flag"
	"fmt"
	"github.com/kiyor/gfind/lib"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	dir       *string = flag.String("dir", ".", "chk dir")
	root      *string = flag.String("rootdir", "", "web server root dir")
	fhost     *string = flag.String("host", "http://server.com", "http hostname")
	fvhost    *string = flag.String("vhost", "client.com", "vhost hostname")
	fctime    *int64  = flag.Int64("ctime", 0, "File's status was last changed n*24 hours ago")
	fcmin     *int64  = flag.Int64("cmin", 0, "File's status was last changed n mins ago")
	fmtime    *int64  = flag.Int64("mtime", 0, "File's data was last changed n*24 hours ago")
	fmmin     *int64  = flag.Int64("mmin", 0, "File's data was last changed n mins ago")
	fatime    *int64  = flag.Int64("atime", 0, "File's data was last access n*24 hours ago")
	famin     *int64  = flag.Int64("amin", 0, "File's data was last access n mins ago")
	fmaxdepth *int    = flag.Int("maxdepth", 0, "Descend at most levels (a non-negative integer) levels of directories below the command line arguments.")
	fftype    *string = flag.String("type", "f", "file type [f|d|l]")
	fsize     *string = flag.String("size", "+0", "file size [-|+]%d[k|m|g]")

	ex      *string = flag.String("exec", "", "exec, use {} as file input")
	verbose *bool   = flag.Bool("v", false, "output analysis")
)

func init() {
	flag.Parse()
}

func InitFindConfByFlag() gfind.FindConf {
	var conf gfind.FindConf
	if flag.Arg(0) != "" {
		fmt.Println(flag.Arg(0))
		conf.Dir = flag.Arg(0)
	} else {
		conf.Dir = *dir
	}
	conf.Stat = new(syscall.Stat_t)
	conf.Maxdepth = *fmaxdepth

	conf.Ctime = *fctime
	conf.Cmin = *fcmin
	conf.Mtime = *fmtime
	conf.Mmin = *fmmin
	conf.Atime = *fatime
	conf.Amin = *famin
	conf.ParseCMTime()

	conf.Ftype = *fftype

	conf.FlatSize = *fsize
	conf.ParseSize()

	return conf
}

func main() {
	conf := InitFindConfByFlag()
	// 	fs := gfind.Find(conf)
	// 	gfind.Output(fs, *verbose)

	ch := make(chan gfind.MyFile)
	go gfind.FindCh(ch, conf)
	if *ex == "" {
		gfind.OutputCh(ch, *verbose)
	} else {
		Exec(ch, *ex)
	}
}

func strip(v []byte) string {
	return strings.TrimSpace(strings.Trim(string(v), "\n"))
}

func Exec(ch chan gfind.MyFile, e string) {
	var v gfind.MyFile
	ok := true
	for ok {
		if v, ok = <-ch; ok {
			r := strings.NewReplacer("{}", v.Path)
			c := r.Replace(e)
			cmd := exec.Command("/bin/bash", "-c", c)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}
}
