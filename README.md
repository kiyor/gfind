<!-----------------------------

- File Name : README.md

- Purpose :

- Creation Date : 03-26-2014

- Last Modified : Thu 10 Apr 2014 12:39:08 AM UTC

- Created By : Kiyor

------------------------------->

# golang find library

it include two way using find
is't super ugly but working good

-	golang slice, in this way, better to do multi process using single file list
-	golang channel, in this way, if you have huge mount of file list and single process

### For Example:

if you need get file list then channel is better. like after you install then run gfind ~
  
else if you need get file list and do a http request to precache something, which in my other repo, then use slice way is better.


### How to use:

``` go

package main

import (
  "github.com/kiyor/gfind/lib"
)

func main() {
  conf := gfind.InitFindConfByIni("config.ini")
  fs := gfind.Find(conf)
  gfind.Output(fs, false)
}

```

### Support flag


| flag     | desc                  |
|----------|-----------------------|
| dir      | location              |
| name     | file name support regex |
| ext      | file extension        |
| rootdir  | ignore prefix path    |
| maxdepth | like find             |
| ctime    | like find             |
| cmin     | like find             |
| mtime    | like find             |
| mmin     | like find             |
| type     | like find             |
| size     | use [-/+]num[k/m/g/t] |
| exec     | try this 'ls {};du {}'|


### sample config


```

[gfind]
dir = /home/user/folder
type = f
rootdir = /home/user
size = +20m
mtime = 1


```


### note
only support linux
