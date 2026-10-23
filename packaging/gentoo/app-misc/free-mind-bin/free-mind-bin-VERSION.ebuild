# Copyright 2026 Tanay PrabhuDesai
# Distributed under the terms of the GNU General Public License v2

# Ebuild repackaging the release .deb, for a personal overlay. The release
# workflow renames this file to free-mind-bin-<version>.ebuild and generates
# its Manifest.

EAPI=8

inherit unpacker xdg

DESCRIPTION="Block distracting websites and stay focused"
HOMEPAGE="https://github.com/tanayseven/free-mind"
_RELEASE="https://github.com/tanayseven/free-mind/releases/download/v${PV}"
SRC_URI="
	amd64? ( ${_RELEASE}/Free.Mind_${PV}_amd64.deb -> ${P}-amd64.deb )
	arm64? ( ${_RELEASE}/Free.Mind_${PV}_arm64.deb -> ${P}-arm64.deb )
"
S="${WORKDIR}"

LICENSE="all-rights-reserved"
SLOT="0"
KEYWORDS="-* ~amd64 ~arm64"
RESTRICT="bindist mirror strip"

# polkit provides pkexec, which the one-time daemon install uses.
RDEPEND="
	dev-libs/libayatana-appindicator
	net-libs/webkit-gtk:4.1
	sys-auth/polkit
	x11-libs/gtk+:3
"

QA_PREBUILT="usr/bin/free-mind"

src_install() {
	cp -a usr "${ED}"/ || die
}
