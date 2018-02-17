`registrars` is a pomf clone/rewrite, written in [Golang][go],
powering [sub.god.jp][subgod]. It implements the [pomf
standard][pomf], including all output types.

No database is required to run registrars, nor are there any special
account features. Files can be uploaded, which will delete themselves
after a specified ammount of time.


Build and run
=============

Short overview (assuming Go and [bindata][bindata]):

	git clone https://github.com/xatasan/registrars
	go generate
	go build -v
	./registrars

Two directories will be created, if not already existing, to store
uploads (`hdir` and `udir`). `udir` contains all the files with the
file names, as they should be downloaded from the server, while `hdir`
contains the files named after the hash value of it's content. Every
file in `udir` links to a file in `hdir`.

To specify which port and on which address `registrars` is supposed to
listen, set the `HOST` environmental variable as follows: 

	HOST="192.168.0.110:9090" ./registrars

If one wishes, it is also possible to only specify the port
(eg. `:25330`) or the interface (eg. `localhost`, `43.211.2.150`, ...)

An additional web server will be required to host the files. 

Auto-deleting files
===================

When starting the server, `registrars` tries to read in file records
fron the stdin. These specify when which file will have to delete
itself, and look like this:

	1518849646	ucolui.png
	1518849720	rgocpy.png
	1518854731	blycoa.png

The two fields are separated by a tab, with the first field containing
the a Unix timestamp signifying when to delete the file and the second
one the to be deleted file, found in `udir`.

To prevent hashfiles from being deleted, set the envvar `KEEPHF` to a
non-zero string. 

Credits
=======

`registrars` was entirely written from scratch, and is in the public
domain (See [LICENSE][legal]).

[go]: https://golang.org/
[subgod]: https://sub.god.jp/
[pomf]: https://github.com/pomf/pomf-standard
[bindata]: https://github.com/jteeuwen/go-bindata/
[legal]: ./LICENSE
