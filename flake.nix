{
  description = "Nix flake for go-wlroots";
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, utils, gomod2nix }:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };
        libxkbcommon =
          pkgs.libxkbcommon.overrideAttrs (finalAttrs: prevAttrs: rec {
            version = "1.6.0";
            src = pkgs.fetchurl {
              url = "https://xkbcommon.org/download/${prevAttrs.pname}-${version}.tar.xz";
              hash = "sha256-DtwU7M3TkVFEWLxfWkuZhj7S1lHk3XYakKv09G75nCs=";
            };
          });
      in
      {
        packages = rec {
          default = tinywl;
          tinywl = with pkgs; buildGoApplication {
            pname = "tinywl";
            version = "v0.0.0";
            src = ./.;
            subPackages = "cmd/tinywl";
            modules = ./gomod2nix.toml;
            nativeBuildInputs = [ pkg-config wayland ];
            buildInputs = [
              libxkbcommon
              pixman
              wlroots
              wayland
              libGL
              udev
              xorg.libX11
              xorg.xcbutilwm
            ];
            preBuild = ''
              wayland-scanner private-code ${wayland-protocols}/share/wayland-protocols/stable/xdg-shell/xdg-shell.xml wlroots/xdg-shell-protocol.c
              wayland-scanner server-header ${wayland-protocols}/share/wayland-protocols/stable/xdg-shell/xdg-shell.xml wlroots/xdg-shell-protocol.h
            '';
          };
        };
        devShells.default = with pkgs; mkShell {
          buildInputs = [
            gcc
            go
            gopls
            gomod2nix.packages.${system}.default
            libxkbcommon
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
