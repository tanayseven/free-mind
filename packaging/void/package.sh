#!/bin/sh
# Builds a Void Linux .xbps from the release .deb with xbps-create. Run inside
# a Void container.
#
# Usage: packaging/void/package.sh <deb> <version> <x86_64|aarch64> <outdir>
set -eu

deb=$1
version=$2
arch=$3
out=$4

stage=$(mktemp -d)
bsdtar -xOf "$deb" 'data.tar.*' | bsdtar -xf - -C "$stage"

# xbps-create doesn't scan for shared libraries like xbps-src does, so the
# runtime packages are listed by hand. polkit provides pkexec for the
# one-time daemon install.
mkdir -p "$out"
cd "$out"
xbps-create \
	--architecture "$arch" \
	--pkgver "free-mind-${version}_1" \
	--desc "Block distracting websites and stay focused" \
	--homepage "https://github.com/tanayseven/free-mind" \
	--license "custom:unknown" \
	--maintainer "Tanay PrabhuDesai <tanayseven@users.noreply.github.com>" \
	--dependencies "libwebkit2gtk41>=0 gtk+3>=0 libayatana-appindicator>=0 polkit>=0" \
	"$stage"
rm -rf "$stage"
