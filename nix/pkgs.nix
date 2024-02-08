# This file controls the pinned version of nixpkgs we use for our Nix environment
# as well as which versions of package we use, including their overrides.
{ config ? { } }:

let 
  # For testing local version of nixpkgs
  #nixpkgsSrc = (import <nixpkgs> { }).lib.cleanSource "/home/jakubgs/work/nixpkgs";

  # We follow the master branch of official nixpkgs.
  nixpkgsSrc = builtins.fetchTarball {
    url = "https://github.com/NixOS/nixpkgs/archive/e7603eba51f2c7820c0a182c6bbb351181caa8e7.tar.gz";
    sha256 = "sha256:0mwck8jyr74wh1b7g6nac1mxy6a0rkppz8n12andsffybsipz5jw";
  };

  # Status specific configuration defaults
  defaultConfig = {
      allowUnfree = true;
      android_sdk.accept_license = true;
  };
  # Override some packages and utilities
  # FIXME: doesn't work due to the use of nixpkgsSrc in androidenv. 
  # See use it in `nix/overlay.nix`
  #  pkgsOverlay = import ./overlay.nix { inherit nixpkgsSrc; };
  pkgsOverlay = [
    (self: super: {
      androidPkgs = nixpkgsSrc.androidenv.composeAndroidPackages {
        toolsVersion = "26.1.1";
        platformToolsVersion = "33.0.3";
        buildToolsVersions = [ "31.0.0" ];
        platformVersions = [ "31" ];
        cmakeVersions = [ "3.18.1" ];
        ndkVersion = "22.1.7171670";
        includeNDK = true;
        includeExtras = [
          "extras;android;m2repository"
          "extras;google;m2repository"
        ];
      };
    })
  ];

in
  # import nixpkgs with a config override
  (import nixpkgsSrc) {
    config = defaultConfig // config;
    overlays =  pkgsOverlay;
  }
