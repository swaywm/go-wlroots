{
  description = "Nix flake for go-wlroots";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { self, flake-utils, nixpkgs }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in {
        devShell = with pkgs; mkShell {
          buildInputs = [
            gcc
            go
            (libxkbcommon.overrideAttrs(finalAttrs: prevAttrs: rec {
              version = "1.6.0";
              src = fetchurl {
                url = "https://xkbcommon.org/download/${prevAttrs.pname}-${version}.tar.xz";
                hash = "sha256-DtwU7M3TkVFEWLxfWkuZhj7S1lHk3XYakKv09G75nCs=";
              };
            }))
            wlroots
            pkg-config
            wayland
            udev
            libGL
            pixman
            xorg.libX11
            xorg.xcbutilwm
          ];
        };
      }
    );
}
