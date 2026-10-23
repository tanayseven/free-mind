#!/usr/bin/env bash
# Fills the package-manager templates in packaging/ with the release version,
# checksums and sizes, ready to submit by hand (Homebrew, winget, Scoop,
# Chocolatey, MacPorts, Solus, AUR, Gentoo).
#
# Usage: scripts/generate-manifests.sh <version> <artifacts-dir> <out-dir>
#
# <artifacts-dir> holds the release workflow's build artifacts, in any layout.
# Set SOURCE_TARBALL to GitHub's archive of the tag to fill in the source
# checksum for the source-built recipes; without it they get SKIP.
set -euo pipefail

version=$1
artifacts=$2
out=$3
packaging=$(cd "$(dirname "$0")/../packaging" && pwd)

# Prints the one artifact matching a file name pattern.
find_one() {
	local matches
	matches=$(find "$artifacts" -type f -name "$1")
	if [ "$(printf '%s' "$matches" | grep -c '')" -ne 1 ]; then
		echo "expected one artifact matching '$1', found: ${matches:-none}" >&2
		exit 1
	fi
	printf '%s\n' "$matches"
}

sha256() {
	sha256sum "$1" | cut -d' ' -f1
}

dmg=$(find_one "*_${version}_universal.dmg")
app_zip=$(find_one "free-mind_${version}_universal.app.zip")
deb_amd64=$(find_one "*_${version}_amd64.deb")
deb_arm64=$(find_one "*_${version}_arm64.deb")
msi_x64=$(find_one "*_${version}_x64_en-US.msi")
msi_arm64=$(find_one "*_${version}_arm64_en-US.msi")
portable_x64=$(find_one "free-mind_${version}_x64_portable.zip")
portable_arm64=$(find_one "free-mind_${version}_arm64_portable.zip")

source_sha=SKIP
if [ -n "${SOURCE_TARBALL:-}" ]; then
	source_sha=$(sha256 "$SOURCE_TARBALL")
fi

substitutions=(
	-e "s/{{VERSION}}/$version/g"
	-e "s/{{SHA256_DMG_UNIVERSAL}}/$(sha256 "$dmg")/g"
	-e "s/{{SHA256_APP_ZIP}}/$(sha256 "$app_zip")/g"
	-e "s/{{SIZE_APP_ZIP}}/$(stat -c %s "$app_zip")/g"
	-e "s/{{SHA256_DEB_AMD64}}/$(sha256 "$deb_amd64")/g"
	-e "s/{{SHA256_DEB_ARM64}}/$(sha256 "$deb_arm64")/g"
	-e "s/{{SHA256_MSI_X64}}/$(sha256 "$msi_x64")/g"
	-e "s/{{SHA256_MSI_ARM64}}/$(sha256 "$msi_arm64")/g"
	-e "s/{{SHA256_PORTABLE_X64}}/$(sha256 "$portable_x64")/g"
	-e "s/{{SHA256_PORTABLE_ARM64}}/$(sha256 "$portable_arm64")/g"
	-e "s/{{SHA256_SOURCE}}/$source_sha/g"
)

# fill <template, relative to packaging/> <output path>
fill() {
	mkdir -p "$(dirname "$2")"
	sed "${substitutions[@]}" "$packaging/$1" >"$2"
}

for template in \
	homebrew/free-mind.rb \
	winget/TanayPrabhuDesai.FreeMind.yaml \
	winget/TanayPrabhuDesai.FreeMind.installer.yaml \
	winget/TanayPrabhuDesai.FreeMind.locale.en-US.yaml \
	scoop/free-mind.json \
	chocolatey/free-mind.nuspec \
	chocolatey/tools/chocolateyinstall.ps1 \
	chocolatey/tools/chocolateyuninstall.ps1 \
	macports/Portfile \
	solus/package.yml \
	aur/free-mind-bin/PKGBUILD \
	aur/free-mind/PKGBUILD; do
	fill "$template" "$out/$template"
done

# Gentoo: a minimal overlay with the ebuild and its Manifest, which lists the
# debs under the names SRC_URI renames them to.
overlay=$out/gentoo
ebuild_dir=$overlay/app-misc/free-mind-bin
fill gentoo/app-misc/free-mind-bin/free-mind-bin-VERSION.ebuild "$ebuild_dir/free-mind-bin-$version.ebuild"
mkdir -p "$overlay/profiles" "$overlay/metadata"
echo free-mind >"$overlay/profiles/repo_name"
printf 'masters = gentoo\nthin-manifests = true\n' >"$overlay/metadata/layout.conf"
dist_entry() {
	echo "DIST $2 $(stat -c %s "$1") BLAKE2B $(b2sum "$1" | cut -d' ' -f1) SHA512 $(sha512sum "$1" | cut -d' ' -f1)"
}
{
	dist_entry "$deb_amd64" "free-mind-bin-$version-amd64.deb"
	dist_entry "$deb_arm64" "free-mind-bin-$version-arm64.deb"
} >"$ebuild_dir/Manifest"

if grep -rl '{{' "$out"; then
	echo "unfilled placeholders in the files above" >&2
	exit 1
fi
