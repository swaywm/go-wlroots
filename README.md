# go-wlroots [![Documentation](https://godoc.org/github.com/swaywm/go-wlroots/wlroots?status.svg)](https://godoc.org/github.com/swaywm/go-wlroots/wlroots)

__go-wlroots__ is a Go binding for [wlroots](https://github.com/swaywm/wlroots).

It is incomplete and supports just enough to run
[tinywl](https://gitlab.freedesktop.org/wlroots/wlroots/-/tree/master/tinywl):

![](https://alexbakker.me/u/ys7ucs0dcw.png)

The source of the Go version of tinywl can be found in [cmd/tinywl](cmd/tinywl).

> [!NOTE]
> There are currently no plans to continue development of go-wlroots, other than
merging the occasional pull request with changes required for compatibility with
new wlroots versions.

## Compiling

Go 1.8 or newer is required.

Make sure [wlroots](https://github.com/swaywm/wlroots) and its dependencies are
installed.

Run ``make all`` to build everything. Binaries can be found in the 'build'
folder.

## License

The source code of this project is licensed under the [MIT license](LICENSE).
>