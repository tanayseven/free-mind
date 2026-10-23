# Skeleton FreeBSD port (sysutils/free-mind) for submission to the ports tree.
# Before submitting, set DISTVERSION, then run `make makesum` and
# `make cargo-crates > Makefile.crates` to generate distinfo and the crate
# list, and vendor the npm dependencies (ports can't download at build time).

PORTNAME=	free-mind
DISTVERSIONPREFIX=	v
DISTVERSION=	0.1.0
CATEGORIES=	sysutils

MAINTAINER=	tanayseven@users.noreply.github.com
COMMENT=	Block distracting websites and stay focused
WWW=		https://github.com/tanayseven/free-mind

LIB_DEPENDS=	libayatana-appindicator3.so:devel/libayatana-appindicator \
		libwebkit2gtk-4.1.so:www/webkit2-gtk@41
BUILD_DEPENDS=	npm:www/npm
RUN_DEPENDS=	pkexec:sysutils/polkit

USES=		cargo gnome pkgconfig
USE_GITHUB=	yes
GH_ACCOUNT=	tanayseven
USE_GNOME=	cairo gdkpixbuf gtk30 librsvg2

CARGO_CARGOTOML=	${WRKSRC}/src-tauri/Cargo.toml
CARGO_FEATURES=	tauri/custom-protocol
CARGO_INSTALL=	no

pre-build:
	cd ${WRKSRC} && npm --prefix frontend ci && npm --prefix frontend run build

do-install:
	cd ${WRKSRC} && sh packaging/install-tree.sh \
		${CARGO_TARGET_DIR}/release/free-mind ${STAGEDIR} ${PREFIX}

.include <bsd.port.mk>
