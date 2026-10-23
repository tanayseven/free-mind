{
  lib,
  rustPlatform,
  cargo-tauri,
  fetchNpmDeps,
  glib-networking,
  gtk3,
  libayatana-appindicator,
  librsvg,
  nodejs,
  npmHooks,
  openssl,
  pkg-config,
  webkitgtk_4_1,
  wrapGAppsHook3,
  xdotool,
}:

let
  root = ../..;
in
rustPlatform.buildRustPackage (finalAttrs: {
  pname = "free-mind";
  version = (lib.importJSON (root + "/src-tauri/tauri.conf.json")).version;

  src = lib.fileset.toSource {
    inherit root;
    fileset = lib.fileset.unions [
      (lib.fileset.difference (root + "/frontend") (
        lib.fileset.maybeMissing (root + "/frontend/node_modules")
      ))
      (lib.fileset.difference (root + "/src-tauri") (
        lib.fileset.maybeMissing (root + "/src-tauri/target")
      ))
    ];
  };

  cargoRoot = "src-tauri";
  buildAndTestSubdir = "src-tauri";
  cargoLock.lockFile = root + "/src-tauri/Cargo.lock";

  npmRoot = "frontend";
  npmDeps = fetchNpmDeps {
    name = "${finalAttrs.pname}-${finalAttrs.version}-npm-deps";
    src = "${finalAttrs.src}/frontend";
    # Update whenever frontend/package-lock.json changes; the nix job in the
    # release workflow prints the new hash when it's stale.
    hash = "sha256-jOCQfmC0ZQjH52Q3dMdCNaIYOXogrSL4Y8NH/HMkyz0=";
  };

  nativeBuildInputs = [
    cargo-tauri.hook
    nodejs
    npmHooks.npmConfigHook
    pkg-config
    wrapGAppsHook3
  ];

  buildInputs = [
    glib-networking
    gtk3
    libayatana-appindicator
    librsvg
    openssl
    webkitgtk_4_1
    xdotool
  ];

  # The Rust tests run in the release workflow's build job.
  doCheck = false;

  # The tray loads libayatana-appindicator with dlopen.
  preFixup = ''
    gappsWrapperArgs+=(--prefix LD_LIBRARY_PATH : ${lib.makeLibraryPath [ libayatana-appindicator ]})
  '';

  meta = {
    description = "Block distracting websites and stay focused";
    homepage = "https://github.com/tanayseven/free-mind";
    # No license until the repository has one. lib.licenses.unfree would make
    # every user set allowUnfree.
    mainProgram = "free-mind";
    platforms = lib.platforms.linux;
  };
})
