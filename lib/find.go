/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : find.go

* Purpose :

* Creation Date : 03-19-2014

* Last Modified : Fri 18 Apr 2014 06:44:18 PM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package gfind

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/vaughan0/go-ini"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type FindConf struct {
	Dir       string
	Name      string
	Stat      *syscall.Stat_t
	Maxdepth  int
	Ftype     string
	Ext       string
	Rootdir   string
	Size      int64
	Smethod   string
	Ctime     int64
	Cmin      int64
	Mtime     int64
	Mmin      int64
	Atime     int64
	Amin      int64
	FlatSize  string
	RsyncTemp int // ignore rsync temp file by defalut, set 1 to noignore. filename like .in.FILENAME.EXT.
}

type MyFile struct {
	Path    string
	Name    string
	Ext     string
	Size    int64
	IsLink  bool
	IsDir   bool
	IsFile  bool
	Relpath string
	Stat    *syscall.Stat_t
}

var (
	rootdir     string
	reRsyncTemp = regexp.MustCompile(`^\.in\..*\.$`)
)

func parseSize(str string) (string, string) {
	if len(str) == 0 {
		return "0", "+"
	}
	var method string
	_, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		return str, "+"
	}
	if str[0:1] == "-" || str[0:1] == "+" {
		method = str[0:1]
		str = str[1:len(str)]
	} else {
		method = "+"
	}

	return str, method
}

func size2H(size int64) string {
	return humanize.IBytes(uint64(size))
}

func sizeFromH(str string) int64 {
	_, err := strconv.Atoi(str[len(str)-1:])
	if err == nil {
		str += "b"
	}
	n, err := strconv.ParseInt(str[:len(str)-1], 10, 64)
	if err != nil {
		fmt.Println("size not able to parse")
		os.Exit(1)
	}
	c := str[len(str)-1:]
	switch c {
	case "K", "k":
		return n * 1024
	case "M", "m":
		return n * 1024 * 1024
	case "G", "g":
		return n * 1024 * 1024 * 1024
	case "T", "t":
		return n * 1024 * 1024 * 1024 * 1024
	case "P", "p":
		return n * 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return n
	}
}

func getIniConfInt(f ini.File, key string) int64 {
	v, ok := f.Get("gfind", key)
	if !ok {
		return 0
	} else {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(key + "should be int")
		}
		return i
	}
}

func (conf *FindConf) ParseSize() {
	if conf.FlatSize == "" {
		conf.Size = 0
		conf.Smethod = "+"
		return
	}
	s, m := parseSize(conf.FlatSize)
	conf.Size = sizeFromH(s)
	conf.Smethod = m
}

func (conf *FindConf) ParseCMTime() {
	now := time.Now().Unix()

	var ct, mt, at syscall.Timespec
	ct.Sec = now - int64(conf.Cmin*60) - int64(conf.Ctime*24*3600)
	mt.Sec = now - int64(conf.Mmin*60) - int64(conf.Mtime*24*3600)
	at.Sec = now - int64(conf.Amin*60) - int64(conf.Atime*24*3600)

	conf.Stat.Ctim = ct
	conf.Stat.Mtim = mt
	conf.Stat.Atim = at
}

func InitFindConfByIni(confloc string) FindConf {
	var conf FindConf
	conf.Stat = new(syscall.Stat_t)
	var ok bool
	f, err := ini.LoadFile(confloc)
	if err != nil {
		panic(confloc + " file not found")
	}

	conf.Dir, ok = f.Get("gfind", "dir")
	if !ok {
		panic("'location' variable missing from 'gfind' section")
	}

	conf.Ftype, ok = f.Get("gfind", "type")
	if !ok {
		conf.Ftype = "f"
	} else {
		if conf.Ftype != "f" && conf.Ftype != "d" && conf.Ftype != "l" {
			fmt.Println("file type not suppoet")
			os.Exit(1)
		}
	}

	conf.Name, ok = f.Get("gfind", "name")
	if !ok {
		conf.Name = ""
	}
	conf.Ext, ok = f.Get("gfind", "ext")
	if !ok {
		conf.Name = ""
	}

	conf.FlatSize, ok = f.Get("gfind", "size")
	if !ok {
		conf.FlatSize = "0"
	} else {
		conf.ParseSize()
	}

	conf.Maxdepth = int(getIniConfInt(f, "maxdepth"))
	conf.Ctime = getIniConfInt(f, "ctime")
	conf.Cmin = getIniConfInt(f, "cmin")
	conf.Mtime = getIniConfInt(f, "mtime")
	conf.Mmin = getIniConfInt(f, "mmin")
	conf.Atime = getIniConfInt(f, "atime")
	conf.Amin = getIniConfInt(f, "amin")
	conf.ParseCMTime()

	conf.RsyncTemp = int(getIniConfInt(f, "rsynctemp"))

	rootdir, ok = f.Get("gfind", "rootdir")
	if !ok {
		conf.Rootdir = ""
	} else {
		if len(rootdir) > 0 {
			if rootdir[len(rootdir)-1:] == "/" {
				rootdir = rootdir[:len(rootdir)-1]
			}
		}
	}

	return conf
}

func Output(fs []MyFile, b bool) {
	var count int
	var size int64
	var str string

	for _, v := range fs {
		if b {
			str = fmt.Sprint(v.Relpath, " ", size2H(v.Size))
		} else {
			str = fmt.Sprint(v.Relpath)
		}
		fmt.Println(str)
		count++
		size += v.Size
	}
	if b {
		fmt.Println("total:", count, "size:", size2H(size))
	}
}

func OutputCh(ch chan MyFile, b bool) {
	var v MyFile
	var count int
	var size int64
	var str string
	ok := true
	for ok {
		if v, ok = <-ch; ok {
			if b {
				str = fmt.Sprint(v.Relpath, " ", size2H(v.Size))
			} else {
				str = fmt.Sprint(v.Relpath)
			}
			fmt.Println(str)
			count++
			size += v.Size
		}
	}
	if b {
		fmt.Println("total:", count, "size:", size2H(size))
	}
}

func chkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (f *MyFile) getInfo(path string) {
	var fstat os.FileInfo
	var err error
	f.Path = path
	f.getName()
	f.Relpath = path[len(rootdir):]
	fstat, err = os.Stat(path)
	if err != nil {
		f.IsLink = true
		fstat, err = os.Lstat(path)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	f.Stat = fstat.Sys().(*syscall.Stat_t)
	f.Size = fstat.Size()
	f.IsDir = fstat.IsDir()

	if !f.IsDir && !f.IsLink {
		f.IsFile = true
	}
}

func (f *MyFile) getName() {
	token := strings.Split(f.Path, "/")
	f.Name = token[len(token)-1:][0]
	f.getExt()
}

func (f *MyFile) getExt() {
	token := strings.Split(f.Name, ".")
	if len(token) > 1 {
		f.Ext = token[len(token)-1:][0]
	}
}

func Find(conf FindConf) []MyFile {
	var fs []MyFile
	err := filepath.Walk(conf.Dir, func(path string, _ os.FileInfo, _ error) error {
		var f MyFile
		f.getInfo(path)

		// only if all true then append
		send := conf.checkMdepth(f) && conf.checkSize(f) && conf.checkCtime(f) && conf.checkMtime(f) && conf.checkFType(f) && conf.checkFName(f) && conf.checkFExt(f) && conf.checkRsyncTemp(f)

		if send {
			fs = append(fs, f)
		}
		return nil
	})
	chkErr(err)
	return fs
}

func FindCh(ch chan MyFile, conf FindConf) {
	err := filepath.Walk(conf.Dir, func(path string, _ os.FileInfo, _ error) error {
		var f MyFile
		f.getInfo(path)

		// only if all true then append
		send := conf.checkMdepth(f) && conf.checkSize(f) && conf.checkCtime(f) && conf.checkMtime(f) && conf.checkAtime(f) && conf.checkFType(f) && conf.checkFName(f) && conf.checkFExt(f) && conf.checkRsyncTemp(f)

		if send {
			ch <- f
		}
		return nil
	})
	chkErr(err)
	close(ch)
}

func (conf *FindConf) checkMdepth(f MyFile) bool {
	if conf.Maxdepth == 0 {
		return true
	} else {
		locationToken := strings.Split(conf.Dir, "/")
		pathToken := strings.Split(f.Path, "/")
		if len(locationToken)+conf.Maxdepth >= len(pathToken) {
			return true
		}
	}
	return false
}

func (conf *FindConf) checkCtime(f MyFile) bool {
	// if not define in conf then return true
	if conf.Ctime == 0 && conf.Cmin == 0 {
		return true
	} else {
		// if file's info create time is later then set conf return true
		if f.Stat.Ctim.Sec > conf.Stat.Ctim.Sec {
			return true
		}
	}
	return false
}

func (conf *FindConf) checkMtime(f MyFile) bool {
	// if not define in conf then return true
	if conf.Mtime == 0 && conf.Mmin == 0 {
		return true
	} else {
		// if file's info modified time is later then set conf return true
		if f.Stat.Mtim.Sec > conf.Stat.Mtim.Sec {
			return true
		}
	}
	return false
}

func (conf *FindConf) checkAtime(f MyFile) bool {
	// if not define in conf then return true
	if conf.Atime == 0 && conf.Amin == 0 {
		return true
	} else {
		// if file's info access time is later then set conf return true
		if f.Stat.Atim.Sec > conf.Stat.Atim.Sec {
			return true
		}
	}
	return false
}

func (conf *FindConf) checkFType(f MyFile) bool {
	if f.IsFile && conf.Ftype == "f" {
		return true
	} else if f.IsDir && conf.Ftype == "d" {
		return true
	} else if f.IsLink && conf.Ftype == "l" {
		return true
	}
	return false
}

func (conf *FindConf) checkFName(f MyFile) bool {
	if conf.Name == "" {
		return true
	}
	re, err := regexp.Compile(conf.Name)
	if err != nil {
		fmt.Println("name regex not able to compile", err.Error())
		os.Exit(1)
	}
	if re.MatchString(f.Name) {
		return true
	}
	return false
}

func (conf *FindConf) checkFExt(f MyFile) bool {
	if conf.Ext == "" {
		return true
	}
	if conf.Ext == f.Ext {
		return true
	}
	return false
}

func (conf *FindConf) checkSize(f MyFile) bool {
	switch conf.Smethod {
	case "-":
		if f.Size < conf.Size {
			return true
		}
	default:
		if f.Size > conf.Size {
			return true
		}
	}
	return false
}

func (conf *FindConf) checkRsyncTemp(f MyFile) bool {
	if conf.RsyncTemp == 1 {
		return true
	}
	if reRsyncTemp.MatchString(f.Name) {
		return false
	}
	return true
}
