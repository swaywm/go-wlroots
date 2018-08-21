# go-wlroots

__go-wlroots__ is a __WIP__ Go binding for
[wlroots](https://github.com/swaywm/wlroots). Note: The API is incomplete and
subject to change.

I'm currently working on porting [SirCmpwn's Tiny Wayland
compositor](https://gist.github.com/SirCmpwn/ae4d1cdcca97ffeb2c35f0878d75dc17):

![](https://u.alexbakker.me/dl?l=HW4XBE3Wc7B_mfQw.png)

## Compiling

Go 1.8 or newer is required.

Make sure [wlroots](https://github.com/swaywm/wlroots) and its dependencies are
installed.

Run ``make all`` to build everything. Binaries can be found in the 'build'
folder.

## License

The source code of this project is licensed under the [MIT license](LICENSE).
