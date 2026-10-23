#!/bin/sh
# Builds the Alpine package inside an Alpine container, then installs it to
# check its shared libraries resolve. The release workflow runs:
#
#   docker run --rm -v "$PWD:/src" -w /src alpine:3.23 sh packaging/alpine/build.sh <version>
#
# Expects free-mind-<version>.tar.gz (a `git archive` of the tree) in the
# current directory and writes free-mind-<version>-r0-<arch>.apk to dist/.
set -eu

version=$1

# abuild resolves makedepends from the local index, so keep it (no --no-cache).
apk update
apk add alpine-sdk
adduser -D builder
addgroup builder abuild

work=/home/builder/free-mind
mkdir -p "$work"
sed "s/{{VERSION}}/$version/" packaging/alpine/APKBUILD >"$work/APKBUILD"
cp "free-mind-$version.tar.gz" "$work/"
chown -R builder "$work"
# abuild refuses to run as root. The signing key is throwaway, so users
# install the package with --allow-untrusted; it's only trusted here so
# abuild can index the package it just built.
su builder -c "abuild-keygen -a -n"
cp /home/builder/.abuild/*.rsa.pub /etc/apk/keys/
su builder -c "cd $work && abuild checksum && REPODEST=$work/repo abuild -r"

arch=$(apk --print-arch)
mkdir -p dist
apk_file=$(find "$work/repo" -name "free-mind-$version-r0.apk")
cp "$apk_file" "dist/free-mind-$version-r0-$arch.apk"

apk add --allow-untrusted "dist/free-mind-$version-r0-$arch.apk"
missing=$(ldd /usr/bin/free-mind | grep 'not found' || true)
if [ -n "$missing" ]; then
	echo "Unresolved libraries:"
	echo "$missing"
	exit 1
fi
