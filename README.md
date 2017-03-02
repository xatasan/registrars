`registrars` is a pomf clone/rewrite, written in [Golang](http://golang.org/), powering [`regist.ra.rs`](http://regist.ra.rs/).

It implements the [pomf standard](https://github.com/pomf/pomf-standard), including all output types.

No database is required to run `registrars`, nor are there any special account features.

## Build and run

After having had made sure that Go is installed, and having downloaded to code run

```sh
cd registrars
go build
./registrars
```

Make sure to execute the program within the directory with the other source code, since it needs the templates and other static files to run properly. Two directories will be created, if not already existing, to store uploads (`hdir` and `udir`).

Now the server should be running on port 8080, accessible by all addresses. To change this behaviour, specify an environmental variable `$HOST`, eg.:

```sh
HOST="192.168.2.107:9090" ./registrars
```

## Credits

`registrars` was entirely written from scratch, and is published under a modified BSD license (See LICENSE).

`registrars` was inspired and built after by [pomf.se](pomf.se), [pomf.cat](https://github.com/banksymate/Pomf), [pomfe.co](https://github.com/H3X-Dev/pomfe.co), [lainfile.pw](https://gitla.in/installgen2/flup), [aww.moe](https://github.com/maxpowa/nodepomf), [fuwa.se](https://github.com/luminarys/eientei).
