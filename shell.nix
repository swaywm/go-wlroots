with import <nixpkgs> {};

mkShell {
  buildInputs = [
    gcc
    go
    libxkbcommon
    wlroots
    pkgconfig
    wayland
    udev
    libGL
    pixman
    xorg.libX11
  ];
}

