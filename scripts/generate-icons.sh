#!/usr/bin/env bash
# Regenerates every app icon from the artwork in frontend/static/.
#
# The source SVGs share an 800x600 canvas with the artwork in the middle. Each
# is cropped to the same square viewBox (the union of their bounding boxes plus
# a little padding), so all icons share one scale and the tray icon doesn't
# jump around when it switches state. Every size is rasterised straight from
# the vector rather than downscaled from a large PNG.
#
# Requires rsvg-convert (librsvg) and the Tauri CLI.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
static="$root/frontend/static"
icons="$root/src-tauri/icons"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# Union of the artwork bounding boxes is x 141..660, y 48..547 (centre 400.5,
# 297.5); a 530 square around that leaves ~1% padding on the widest side.
viewbox="135.5 32.5 530 530"

# Rewrites the root <svg> size and viewBox to the square crop.
crop() {
	# The root <svg> is on line 2, after the XML declaration; `1,/re/` (unlike
	# GNU-only `0,/re/`) works with both GNU and BSD sed.
	sed -E '1,/<svg /s/width="800" height="600" viewBox="0 0 800 600"/width="530" height="530" viewBox="'"$viewbox"'"/' "$1"
}

render() { # svg size out
	rsvg-convert -w "$2" -h "$2" "$1" -o "$3"
}

# Tray / window status icons: name:source artwork. (A plain list rather than
# an associative array so it runs on macOS's bash 3.2.)
status=(
	blocked:frame
	unblocking:broken-chain-1
	unblocked:broken-chain-2
	inactive:frame-grayscale
)

mkdir -p "$icons/status"
for entry in "${status[@]}"; do
	name="${entry%%:*}"
	crop "$static/${entry#*:}.svg" >"$tmp/$name.svg"
	render "$tmp/$name.svg" 64 "$icons/status/$name.png"
done

# App bundle icons (all platforms) from the locked brain.
crop "$static/frame.svg" >"$tmp/app.svg"
render "$tmp/app.svg" 1024 "$tmp/app.png"
(cd "$root" && tauri icon "$tmp/app.png")

# Favicon: cropped vector plus PNG fallbacks.
cp "$tmp/app.svg" "$static/favicon.svg"
render "$tmp/app.svg" 32 "$static/favicon.png"
render "$tmp/app.svg" 180 "$static/apple-touch-icon.png"
