# Override some packages and utilities in 'pkgs'
# and make them available globally via callPackage.
#
# For more details see:
# - https://nixos.wiki/wiki/Overlays
# - https://nixos.org/nixos/nix-pills/callpackage-design-pattern.html

{ pkgs }:
let
  
in {
  /* FIXME not sure for the pkgs*/
  androidPkgs = pkgs.androidenv.composeAndroidPackages {
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
}
