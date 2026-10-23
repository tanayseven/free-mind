#!/bin/sh
# Builds a Slackware .txz from the release .deb. Slackware packages are plain
# tarballs of the installed tree plus install/slack-desc.
#
# Usage: packaging/slackware/package.sh <deb> <version> <x86_64|aarch64> <outdir>
set -eu

deb=$1
version=$2
arch=$3
out=$4

stage=$(mktemp -d)
dpkg-deb -x "$deb" "$stage"
mkdir -p "$stage/install"
cp "$(dirname "$0")/slack-desc" "$stage/install/slack-desc"

mkdir -p "$out"
tar --owner=0 --group=0 -C "$stage" -cJf "$out/free-mind-$version-$arch-1.txz" .
rm -rf "$stage"
