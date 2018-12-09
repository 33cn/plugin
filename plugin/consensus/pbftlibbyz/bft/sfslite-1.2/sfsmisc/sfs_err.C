/* $Id: sfs_err.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */

#include "sfs_prot.h"
#include "sha1.h"
#include "amisc.h"
#include "serial.h"
#include "sfsmisc.h"

const strbuf &
strbuf_cat (const strbuf &sb, sfsstat err)
{
  switch (err) {
  case SFS_REDIRECT:
    return strbuf_cat (sb, "hostname or public key has changed", false);
  default:
    return strbuf_cat (sb, nfsstat3 (err));
  }
}

const strbuf &
strbuf_cat (const strbuf &sb, sfsauth_stat status)
{
  switch (status) {
  case SFSAUTH_OK:
    return strbuf_cat (sb, "Authorization OK");
  case SFSAUTH_LOGINMORE:
    return strbuf_cat (sb, "Authorization login not sufficient yet");
  case SFSAUTH_FAILED:
    return strbuf_cat (sb, "Authorization failed");
  case SFSAUTH_LOGINALLBAD:
    return strbuf_cat (sb, "Authorization invalid login; don't try again");
  case SFSAUTH_NOTSOCK:
    return strbuf_cat (sb, "Authorization failed--can't conect via network");
  case SFSAUTH_BADUSERNAME:
    return strbuf_cat (sb, "Authorization failed--username not in password"
		       " file");
  case SFSAUTH_WRONGUID:
    return strbuf_cat (sb, "Authorization failed--uid doesn't match"
		       " username");
  case SFSAUTH_DENYROOT:
    return strbuf_cat (sb, "Authorization failed--Can't register root"
		       " account");
  case SFSAUTH_BADSHELL:
    return strbuf_cat (sb, "Authorization failed--shell is not in"
		       " /etc/shells");
  case SFSAUTH_DENYFILE:
    return strbuf_cat (sb, "Authorization failed--user explicitly denied");
  case SFSAUTH_BADPASSWORD:
    return strbuf_cat (sb, "Authorization failed--incorrect password");
  case SFSAUTH_USEREXISTS:
    return strbuf_cat (sb, "Authorization failed--user exists in auth"
		       " database");
  case SFSAUTH_NOCHANGES:
    return strbuf_cat (sb, "Authorization failed--no changes");
  case SFSAUTH_BADSIGNATURE:
    return strbuf_cat (sb, "Authorization failed--bad signature");
  case SFSAUTH_PROTOERR:
    return strbuf_cat (sb, "Authorization failed--protocol error");
  case SFSAUTH_NOTTHERE:
    return strbuf_cat (sb, "Authorization failed--not there");
  case SFSAUTH_BADAUTHID:
    return strbuf_cat (sb, "Authorization failed--bad auth ID");
  case SFSAUTH_KEYEXISTS:
    return strbuf_cat (sb, "Authorization failed--key exists");
  case SFSAUTH_BADKEYNAME:
    return strbuf_cat (sb, "Authorization failed--bad key name");
  default:
    return strbuf_cat (sb, "Unknown response from authserv");
  }
}
