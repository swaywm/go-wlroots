{
  description = "Nix flake for go-wlroots";
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        tinywl = pkgs.buildGoModule {
          pname = "tinywl";
          version = "0.0.0";
          src = self;
          vendorHash = "sha256-0vsi8q/+H7I5qvyr9sIZFLD30yZivySLWYT6KrJqIcc=";
          subPackages = [ "cmd/tinywl" ];
          nativeBuildInputs = [
            pkgs.pkg-config
            pkgs.wayland-scanner
          ];
          buildInputs = [
            pkgs.libxkbcommon
            pkgs.pixman
            pkgs.wlroots
            pkgs.wayland
            pkgs.libGL
            pkgs.udev
            pkgs.xorg.libX11
            pkgs.xorg.xcbutilwm
          ];
          preBuild = ''
            wayland-scanner private-code ${pkgs.wayland-protocols}/share/wayland-protocols/stable/xdg-shell/xdg-shell.xml wlroots/xdg-shell-protocol.c
            wayland-scanner server-header ${pkgs.wayland-protocols}/share/wayland-protocols/stable/xdg-shell/xdg-shell.xml wlroots/xdg-shell-protocol.h
          '';
        };
      in
      {
        packages = {
          inherit tinywl;
          default = tinywl;
        };
        devShells.default = pkgs.mkShell {
          inputsFrom = [ tinywl ];
          buildInputs = [
            pkgs.gcc
            pkgs.go
            pkgs.gopls
            pkgs.libxkbcommon
            pkgs.wlroots
            pkgs.pkg-config
            pkgs.wayland
            pkgs.udev
            pkgs.libGL
            pkgs.pixman
            pkgs.xorg.libX11
            pkgs.xorg.xcbutilwm
          ];
        };
      }
    );
}
