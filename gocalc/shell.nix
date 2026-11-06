{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.wails
    pkgs.pkg-config
    pkgs.gtk3
    pkgs.webkitgtk_4_1
    pkgs.nodejs
  ];
}
