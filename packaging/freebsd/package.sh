#!/bin/sh
# Builds Free Mind from source on FreeBSD and packages it with `pkg create`.
# Run from the repository root inside a FreeBSD VM.
#
# Usage: packaging/freebsd/package.sh <version> <outdir>
set -eu

version=$1
out=$2
# Runtime packages; polkit provides pkexec for the one-time daemon install.
runtime="webkit2-gtk_41 gtk3 libayatana-appindicator librsvg2-rust polkit"

npm --prefix frontend ci
npm --prefix frontend run build
cargo build --release --locked \
	--manifest-path src-tauri/Cargo.toml \
	--features tauri/custom-protocol

stage=$(mktemp -d)
sh packaging/install-tree.sh src-tauri/target/release/free-mind "$stage" /usr/local

meta=$(mktemp -d)
sed "s/{{VERSION}}/$version/" packaging/freebsd/+MANIFEST >"$meta/+MANIFEST"
# pkg needs each dependency's origin and version, so read them from the
# packages installed for the build.
{
	echo "deps: {"
	# shellcheck disable=SC2086 # one argument per package
	pkg query '  %n: { origin: "%o", version: "%v" }' $runtime
	echo "}"
} >>"$meta/+MANIFEST"
(cd "$stage" && find . -type f | sed 's|^\./|/|') >"$meta/plist"

mkdir -p "$out"
pkg create -M "$meta/+MANIFEST" -p "$meta/plist" -r "$stage" -o "$out"
rm -rf "$stage" "$meta"
