`registrars` is a pomf clone/rewrite, written in Golang, powering
[sub.god.jp](https://sub.god.jp). It implements the [pomf
standard](https://github.com/pomf/pomf-standard), including all output
types. No database is required to run registrars, nor are there any
special account features. Files can be uploaded, which will delete
themselves after a specified ammount of time.

Build and run
=============

After having had made sure that Go is installed, and having downloaded
to code run "go build". This should produce a binary named
"registrars".

Make sure to execute the program within the directory with the other
source code, since it needs the templates and other static files to
run properly. Two directories will be created, if not already
existing, to store uploads (`hdir` and `udir`). `udir` contains all
the files with the file names, as they should be downloaded from the
server, while `hdir` contains the files named after the hash value of
it's content. Every file in `udir` links to a file in `hdir`.

Now the server should be running on port 8080, accessible by all
addresses. To change this behaviour, specify an environmental variable
"$HOST", eg.:

```sh
HOST="192.168.1.107:9090" ./registrars https://u.fileserver.com/f/
```

The first argument specifies the base url, onto which all uploaded
filenames are appended. So for example if the filename `Hrke417i.png`
were to be generated, registrars would create the link
`https://u.fileserver.com/f/Hrke417i.png`.

Auto-deleting files
===================

When starting the server, `registrars` tries to read in file records
fron the stdin. These specify when which file will have to delete
itself, and look like this:

```
Mon Jan  8 17:06:17 UTC 2018	Hu8eJ17I.png
```

The two fileds are seperated by a tab, with the first field containing
the date formatted with `UnixDate` (Golang date string: "Mon Jan _2
15:04:05 MST 2006"), and the second one the file found in `udir`.

Unless the environmental variable `KEEPHF` has any value, after all
references to a hash file have been deleted from `udir`, the
respective file in `hdir` will be deleted too.

Credits
=======

`registrars` was entirely written from scratch, and is in the public
domain (See LICENSE).
