/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : find.go

* Purpose :

* Creation Date : 03-19-2014

* Last Modified : Wed 28 May 2014 12:21:52 AM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package gfind

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/vaughan0/go-ini"
	"log"
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
	RevTime   bool
	FlatSize  string
	RsyncTemp int // ignore rsync temp file by defalut, set 1 to noignore. filename like .in.FILENAME.EXT.
}

type File struct {
	os.FileInfo
	Path    string
	Ext     string
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
		log.Fatalln("size not able to parse")
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
			log.Fatalln("file type not suppoet")
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

	var revt string
	revt, ok = f.Get("gfind", "revtime")
	if ok {
		if revt == "true" {
			conf.RevTime = true
		}
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
	}
	conf.SetRootdir()

	return conf
}

// set root dir by conf.Rootdir
func (c *FindConf) SetRootdir() {
	if len(c.Rootdir) > 0 {
		if c.Rootdir[len(c.Rootdir)-1:] == "/" {
			rootdir = c.Rootdir[:len(c.Rootdir)-1]
		} else {
			rootdir = c.Rootdir[:len(c.Rootdir)-1] + "/"
		}
	}
}

func Output(fs []File, b bool) {
	var count int
	var size int64
	var str string

	for _, v := range fs {
		if b {
			str = fmt.Sprint(v.Relpath, " ", size2H(v.Size()))
		} else {
			str = fmt.Sprint(v.Relpath)
		}
		fmt.Println(str)
		count++
		size += v.Size()
	}
	if b {
		fmt.Println("total:", count, "size:", size2H(size))
	}
}

func OutputCh(ch chan File, b bool) {
	var v File
	var count int
	var size int64
	var str string
	ok := true
	for ok {
		if v, ok = <-ch; ok {
			if b {
				str = fmt.Sprint(v.Relpath, " ", size2H(v.Size()))
			} else {
				str = fmt.Sprint(v.Relpath)
			}
			fmt.Println(str)
			count++
			size += v.Size()
		}
	}
	if b {
		fmt.Println("total:", count, "size:", size2H(size))
	}
}

func chkErr(err error) {
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func (f *File) IsLink() bool {
	realPath, _ := os.Readlink(f.Path)
	if realPath != "" {
		return true
	}
	return false
}

func (f *File) getInfo(path string) error {
	var fstat os.FileInfo
	var err error
	f.Path = path
	fstat, err = os.Stat(path)
	if err != nil {
		fstat, err = os.Lstat(path)
		if err != nil {
			return err
		}
	}
	f.FileInfo = fstat
	f.Relpath = path[len(rootdir):]
	// 	f.Relpath, err = filepath.Rel(rootdir, f.Path)
	// 	if err != nil {
	// 		log.Fatalln(err.Error())
	// 		return
	// 	}
	f.Stat = fstat.Sys().(*syscall.Stat_t)
	f.getExt()

	if !f.IsDir() && !f.IsLink() {
		f.IsFile = true
	}
	return nil
}

func (f *File) getExt() {
	token := strings.Split(f.Name(), ".")
	if len(token) > 1 {
		f.Ext = token[len(token)-1:][0]
	}
}

func Find(conf FindConf) []File {
	var fs []File
	err := filepath.Walk(conf.Dir, func(path string, _ os.FileInfo, _ error) error {
		var f File
		err := f.getInfo(path)
		if err != nil {
			return nil
		}

		// only if all true then append
		send := conf.checkAll(f)

		if send {
			fs = append(fs, f)
		}
		return nil
	})
	chkErr(err)
	return fs
}

func FindCh(ch chan File, conf FindConf) {
	err := filepath.Walk(conf.Dir, func(path string, _ os.FileInfo, _ error) error {
		var f File
		err := f.getInfo(path)
		if err != nil {
			return nil
		}

		// only if all true then append
		send := conf.checkAll(f)

		if send {
			ch <- f
		}
		return nil
	})
	chkErr(err)
	close(ch)
}

func (conf *FindConf) checkAll(f File) bool {
	return conf.checkMdepth(f) && conf.checkSize(f) && conf.checkCtime(f) && conf.checkMtime(f) && conf.checkAtime(f) && conf.checkFType(f) && conf.checkFName(f) && conf.checkFExt(f) && conf.checkRsyncTemp(f)
}

func (conf *FindConf) checkMdepth(f File) bool {
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

func (conf *FindConf) checkCtime(f File) bool {
	// if not define in conf then return true
	if conf.Ctime == 0 && conf.Cmin == 0 {
		return true
	} else {
		// if file's info create time is later then set conf return true
		if !conf.RevTime {
			if f.Stat.Ctim.Sec > conf.Stat.Ctim.Sec {
				return true
			}
		} else {
			if f.Stat.Ctim.Sec < conf.Stat.Ctim.Sec {
				return true
			}
		}
	}
	return false
}

func (conf *FindConf) checkMtime(f File) bool {
	// if not define in conf then return true
	if conf.Mtime == 0 && conf.Mmin == 0 {
		return true
	} else {
		// if file's info modified time is later then set conf return true
		if !conf.RevTime {
			if f.Stat.Mtim.Sec > conf.Stat.Mtim.Sec {
				return true
			}
		} else {
			if f.Stat.Mtim.Sec < conf.Stat.Mtim.Sec {
				return true
			}
		}
	}
	return false
}

func (conf *FindConf) checkAtime(f File) bool {
	// if not define in conf then return true
	if conf.Atime == 0 && conf.Amin == 0 {
		return true
	} else {
		// if file's info access time is later then set conf return true
		if !conf.RevTime {
			if f.Stat.Atim.Sec > conf.Stat.Atim.Sec {
				return true
			}
		} else {
			if f.Stat.Atim.Sec < conf.Stat.Atim.Sec {
				return true
			}
		}
	}
	return false
}

func (conf *FindConf) checkFType(f File) bool {
	if f.IsFile && conf.Ftype == "f" {
		return true
	} else if f.IsDir() && conf.Ftype == "d" {
		return true
	} else if f.IsLink() && conf.Ftype == "l" {
		return true
	}
	return false
}

func (conf *FindConf) checkFName(f File) bool {
	if conf.Name == "" {
		return true
	}
	re, err := regexp.Compile(conf.Name)
	if err != nil {
		log.Fatalln(conf.Name, "name regex not able to compile", err.Error())
		return false
		// 		os.Exit(1)
	}
	if re.MatchString(f.Name()) {
		return true
	}
	return false
}

func (conf *FindConf) checkFExt(f File) bool {
	if conf.Ext == "" {
		return true
	}
	if conf.Ext == f.Ext {
		return true
	}
	return false
}

func (conf *FindConf) checkSize(f File) bool {
	// 	defer func() {
	// 		if r := recover(); r != nil {
	// 			log.Fatalln("here")
	// 		}
	// 	}()
	switch conf.Smethod {
	case "-":
		if f.Size() < conf.Size {
			return true
		}
	default:
		if f.Size() >= conf.Size {
			return true
		}
	}
	return false
}

func (conf *FindConf) checkRsyncTemp(f File) bool {
	if conf.RsyncTemp == 1 {
		return true
	}
	if reRsyncTemp.MatchString(f.Name()) {
		return false
	}
	return true
}
