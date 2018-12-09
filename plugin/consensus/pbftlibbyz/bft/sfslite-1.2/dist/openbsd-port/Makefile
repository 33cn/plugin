# New ports collection makefile for:	sfslite
# Date created:				10 September 2005
# Whom:					max
#
# $FreeBSD: ports/devel/sfslite/Makefile,v 1.6 2006/08/14 22:55:16 clsung Exp $
#

BASENAME=	sfslite
VERSION=	1.0.2
DISTNAME=       ${BASENAME}-${VERSION}
CATEGORIES=	devel

MASTER_SITES=	http://dist.okws.org/dist/

MAINTAINER=	ports@okws.org
COMMENT=	"Standard version of the sfslite library"

USE_GMAKE=	yes
AUTOCONF_VERSION = 2.59
AUTOMAKE_VERSION = 1.9
CONFIGURE_STYLE = gnu

CONFIGURE_ARGS+= --with-mode=shared \
		 --with-libsfs

BUILD_DEPENDS+=	bison-*::devel/bison
LIB_DEPENDS=	gmp::devel/gmp

REV=1.0
SUBST_VARS=REV

NO_REGRESS=	Yes

post-install:
	ln -s ${PREFIX}/lib/sfslite-${REV} ${PREFIX}/lib/sfslite
	ln -s ${PREFIX}/include/sfslite-${REV} ${PREFIX}/include/sfslite

.include <bsd.port.mk>
