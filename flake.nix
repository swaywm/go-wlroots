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
            libxkbcommon
            wlroots
            pkg-config
            wayland
            udev
            libGL
            pixman
            xorg.libX11
          ];
        };
      }
    );
}
