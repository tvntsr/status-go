{ config ? {}
, pkgs ? import ./pkgs.nix { inherit config; } }:

let
  inherit (pkgs) lib stdenv;
  /* No Android SDK for Darwin aarch64. */
  isMacM1 = stdenv.isDarwin && stdenv.isAarch64;
  /* Lock requires Xcode verison. */
  xcodeWrapper = pkgs.xcodeenv.composeXcodeWrapper {
    version = "14.3";
    allowHigher = true;
  };
  /* Gomobile also needs the Xcode wrapper. */
  #FIXME:something is not working with this
  gomobileMod = pkgs.gomobile.override {
    inherit xcodeWrapper;
    withAndroidPkgs = !isMacM1;
  };
in pkgs.mkShell {
  name = "status-go-shell";

  buildInputs = with pkgs; [
    git jq which
    go_1_19 golangci-lint go-junit-report gopls go-bindata gomobileMod
    mockgen protobuf3_20 protoc-gen-go gotestsum
  ] ++ lib.optional stdenv.isDarwin xcodeWrapper;
   # For some reason `ANDROID_HOME=${pkgs.androidPkgs.androidsdk}/libexec/android-sdk` is not working
   shellHook = lib.optionalString (!isMacM1) ''
     ANDROID_NDK=$ANDROID_HOME/ndk-bundle
     ANDROID_SDK_ROOT=$ANDROID_HOME
     ANDROID_NDK_HOME=$ANDROID_NDK
  '';
  # Sandbox causes Xcode issues on MacOS. Requires sandbox=relaxed.
  # https://github.com/status-im/status-mobile/pull/13912
  __noChroot = stdenv.isDarwin;
}

