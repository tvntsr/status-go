# This file defines custom shells as well as shortcuts
# for accessing more nested shells.
{ config ? {}
, pkgs ? import ./pkgs.nix { inherit config; } }:

let
  inherit (pkgs) lib callPackage;

  default = callPackage ./shell.nix { };

  shells = {
    inherit default;

  # Other shell for makefile or CI
  };
  mergeDefaultShell = (_: val: lib.mergeSh default [ val ]);

# values here can be selected using `nix-shell --attr shells.$TARGET default.nix`
# the nix/scripts/shell.sh wrapper does this for us and expects TARGET to be set
in lib.mapAttrs mergeDefaultShell shells
