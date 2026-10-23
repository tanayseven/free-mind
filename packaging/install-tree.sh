#!/bin/sh
# Installs a source-built Free Mind binary with its desktop entry and icons.
# Used by packages that compile from source instead of repackaging the .deb
# (Alpine, AUR source, Solus, FreeBSD).
#
# Usage: packaging/install-tree.sh <binary> <destdir> <prefix>
#   e.g. packaging/install-tree.sh src-tauri/target/release/free-mind "$pkgdir" /usr
set -eu

bin=$1
dest=$2$3
root=$(cd "$(dirname "$0")/.." && pwd)
icons=$root/src-tauri/icons

# BSD install has no -D, so create parent directories first.
put() {
	mkdir -p "$(dirname "$3")"
	install -m "$1" "$2" "$3"
}

put 755 "$bin" "$dest/bin/free-mind"
put 644 "$root/packaging/linux/free-mind.desktop" "$dest/share/applications/free-mind.desktop"
put 644 "$icons/32x32.png" "$dest/share/icons/hicolor/32x32/apps/free-mind.png"
put 644 "$icons/64x64.png" "$dest/share/icons/hicolor/64x64/apps/free-mind.png"
put 644 "$icons/128x128.png" "$dest/share/icons/hicolor/128x128/apps/free-mind.png"
put 644 "$icons/128x128@2x.png" "$dest/share/icons/hicolor/128x128@2/apps/free-mind.png"
