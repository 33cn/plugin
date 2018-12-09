/* $Id: sfspath.C 2679 2007-04-04 20:53:20Z max $ */

/*
 *
 * Copyright (C) 1999 David Mazieres (dm@uun.org)
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

#include "sfscrypt.h"
#include "sfsmisc.h"
#include "crypt.h"
#include "parseopt.h"

#if 0
#include "rxx.h"
static rxx ppv2 ("@([\\w\\.\\-]+)(%([1-9]\\d*))?,([a-z0-9]+)");
#endif


bool
sfs_ascii2hostid (sfs_hash *hid, const char *p)
{
  const u_char *const up = reinterpret_cast<const u_char *> (p);
  if (armor32len (up) != ascii_hostid_len || p[ascii_hostid_len])
    return false;
  if (hid) {
    str raw = dearmor32 (p, ascii_hostid_len);
    memcpy (hid->base (), raw, hid->size ());
  }
  return true;
}

bool
sfsgethost_label (const char *&p)
{
  const char *s = p;
  while ((*s >= 'a' && *s <= 'z')
	 || (*s >= '0' && *s <= '9') || *s == '-')
    s++;
  if (s - p == 0 || s - p > 63)
    return false;
  p = s;
  return true;
}
bool
sfsgethost_dotlabel (const char *&p)
{
  const char *s = p;
  if (*s++ != '.')
    return false;
  if (!sfsgethost_label (s))
    return false;
  p = s;
  return true;
}
str
sfsgethost (const char *&p, bool qualified)
{
  const char *s = p;
  if (!sfsgethost_label (s) || (qualified && !sfsgethost_dotlabel (s)))
    return NULL;
  while (sfsgethost_dotlabel (s))
    ;
  if (s - p > 255)
    return NULL;

  str host (p, s - p);
  p = s;
  return host;
}

bool
sfs_parsepath (str path, str *host, sfs_hash *hostid, u_int16_t *portp,
	       int *versp)
{
  int vers;
  bool ret;
  const char *p = path;
  if (strchr (p, ':')) {
    vers = 1;
    ret =  sfs_parsepath_v1 (path, host, hostid, portp);
  } else {
    vers = 2;
    ret = sfs_parsepath_v2 (path, host, hostid, portp);
  }
  if (versp) *versp = vers;
  return ret;
}

bool
sfsgetlocation (const char *&pp, str *hostp, u_int16_t *portp,
		bool qualified)
{
  const char *p = pp;

  str host = sfsgethost (p, qualified);
  if (!host)
    return false;

  int64_t port = 0;
  if (*p == '%') {
    p++;
    port = strtoi64 (p, const_cast<char **> (&p), 10);
    if (port <= 0 || port >= 0xffff || isdigit (*p))
      return false;
  }

  pp = p;
  if (hostp)
    *hostp = host;
  if (portp)
    *portp = port;
  return true;
}

bool
sfsgetatlocation (const char *&pp, str *hostp, u_int16_t *portp,
		  bool qualified)
{
  const char *p = pp;
  if (*p++ != '@')
    return false;
  if (!sfsgetlocation (p, hostp, portp, qualified))
    return false;
  pp = p;
  return true;
}

bool
sfs_parsepath_v2 (str path, str *hostp, sfs_hash *hostidp, u_int16_t *portp)
{
#if 1
  str host;
  u_int16_t port;
  sfs_hash dummy;
  if (!hostidp) 
    hostidp = &dummy;

  const char *p = path;
  if (!sfsgetatlocation (p, &host, &port))
    return false;
  if (*p++ != ',')
    return false;
  if (!sfs_ascii2hostid (hostidp, p))
    return false;

  if (hostp)
    *hostp = host;
  if (portp)
    *portp = port;
  return true;

#else
  u_int16_t port = 0;
  const char *cp;
  sfs_hash dummy;
  if (!hostidp) 
    hostidp = &dummy;
  
  if (!ppv2.match (path))
    return false;
  if (!ppv2[4] || !ppv2[4].len () || !sfs_ascii2hostid (hostidp, ppv2[4]))
    return false;
  if (!ppv2[1] || !ppv2[1].len () || !(cp = ppv2[1]) || !sfsgethost (cp))
    return false;
  if (hostp) 
    *hostp = ppv2[1];
  if (str portstr = ppv2[3]) {
    if (!portstr.len () || portstr[0] == '\0')
      return false;
    char *endp = NULL;
    int64_t atno = strtoi64 (portstr, &endp, 10);
    if (atno <= 0 || atno >= 0x10000 || !endp || *endp)
      return false;
    port = atno;
  }
  if (portp)
    *portp = port;
  return true;
#endif
}

bool
sfs_parsepath_v1 (str path, str *host, sfs_hash *hostid, u_int16_t *portp)
{
  const char *p = path;
  u_int16_t port = 0;

  if (isdigit (*p)) {
    int64_t atno = strtoi64 (path, const_cast<char **> (&p), 10);
    if (p && *p == '@' && atno > 0 && atno < 0x10000) {
      port = atno;
      path = substr (path, p + 1 - path.cstr ());
    }
    p = path;
  }

  if (portp)
    *portp = port;
  if (!sfsgethost (p))
    return false;
  if (host)
    host->setbuf (path, p - path);
  return *p++ == ':' && sfs_ascii2hostid (hostid, p);
}

ptr<sfs_servinfo_w>
sfs_servinfo_w::alloc (const sfs_servinfo &s)
{
  if (s.sivers < 7) {
    return New refcounted<sfs_servinfo_w_v1> (s);
  } else if (s.sivers == 7) {
    return New refcounted<sfs_servinfo_w_v2> (s);
  } else {
    warn << "Cannot parse sfs_servinfo for release " << s.sivers << "\n";
    return NULL;
  }
}

ref<sfs_servinfo_w>
sfs_servinfo_w::alloc (const sfs_hostinfo2 &h)
{
  sfs_servinfo si;
  si.set_sivers (7);
  si.cr7->host = h;
  return New refcounted<sfs_servinfo_w_v2> (si);
}

bool
sfs_servinfo_w::mkhostid (sfs_hash *id, int vers) const
{
  switch (vers) {
  case 2:
    return mkhostid_v2 (id);
  case 1:
    return mkhostid_v1 (id);
  default:
    warn << "sfs_mkhostid: unrecognized version number\n";
    return false;
  }
}

bool
sfs_servinfo_w::mkhostid_v2 (sfs_hash *id) const
{
  ptr<sfspub> pk = sfscrypt.alloc (get_pubkey ());
  return pk->get_pubkey_hash (id, 2);
}

bool 
sfs_servinfo_w::mkhostid_v1 (sfs_hash *id) const
{
  sfs_hostinfo info;
  info.type = SFS_HOSTINFO;
  info.hostname = get_hostname ();
  if (!(info.pubkey = get_rabin_pubkey ()))
    return false;
  const char *p = info.hostname;
  if (info.type != SFS_HOSTINFO || !sfsgethost (p) || *p) {
    bzero (id->base (), id->size ());
    return false;
  }
  
  xdrsuio x;
  if (!xdr_sfs_hostinfo (&x, const_cast<sfs_hostinfo *> (&info))
      || !xdr_sfs_hostinfo (&x, const_cast<sfs_hostinfo *> (&info))) {
    warn << "sfs_mkhostid (" << info.hostname << ", "
	 << hexdump (id->base (), id->size ()) << "): XDR failed!\n";
    bzero (id->base (), id->size ());
    return false;
  }
  sha1_hashv (id->base (), x.iov (), x.iovcnt ());
  return true;
}

bigint
sfs_servinfo_w_v2::get_rabin_pubkey () const 
{
  sfs_pubkey2 pk2 = get_pubkey ();
  if (pk2.type == SFS_RABIN) 
    return *pk2.rabin;
  else
    return 0;
}

str
sfs_servinfo_w::mkpath (int vers, int port) const
{
  strbuf b;
  sfs_hash hostid;
  if (!mkhostid (&hostid, vers))
    return ":ERROR: Could not make hostid"; 
  if (vers == 1) {
    if (port)
      b << port << "@";
    b << get_hostname () << ":" << armor32 (&hostid, sizeof (hostid)) ;
  } else {
    b << "@" << get_hostname ();
    if (port == -1)
      port = get_port ();
    if (port)
      b << "%" << port;
    b << "," << armor32 (&hostid, sizeof (hostid));
  }
  return b;
}


bool
sfs_servinfo_w::ckhostid (const sfs_hash *id, int vers) const
{
  sfs_hash rid;
  return mkhostid (&rid, vers) && !memcmp (&rid, id, sizeof (rid));
}

bool
sfs_servinfo_w::ckpath (const str &sname) const
{
  str location;
  sfs_hash hid;
  int vers;
  if (!sfs_parsepath (sname, &location, &hid, NULL, &vers))
    return false;
  str hn = get_hostname ();
  if (location != hn)
    return false;
  if (!ckhostid (&hid, vers))
    return false;
  return true;
}

bool
sfs_servinfo_w::operator== (const sfs_servinfo_w &s) const 
{
  ptr<sfspub> pk = sfscrypt.alloc (get_pubkey ());
  str h1 = get_hostname ();
  str h2 = s.get_hostname ();
  return (pk && *pk == s.get_pubkey () && 
	  ((!h1 && !h2) || (h1 && h2 && h1 == h2)));
}

sfs_pubkey2
sfs_servinfo_w_v1::get_pubkey () const 
{
  sfs_pubkey2 pk (SFS_RABIN);
  *(pk.rabin) = si.cr5->host.pubkey;
  return pk;
}

bool
sfs_servinfo_w::ckci (const sfs_connectinfo &ci) const
{
  if (ci.civers <= 4)
    return ckhostid (&ci.ci4->hostid, 1);
  return ckpath (ci.ci5->sname);
}

bool
sfs_pathrevoke_w::check (sfs_hash *hidp)
{
  sfs_hash hostid;
  if (!hidp) hidp = &hostid;
  if (!si || !si->mkhostid (hidp))
    return false;

  if (rev.msg.type != SFS_PATHREVOKE
      || rev.msg.path.type != SFS_HOSTINFO
      || (rev.msg.redirect && rev.msg.redirect->hostinfo.type != SFS_HOSTINFO))
    return false;

  if (!sfscrypt.verify (rev.msg.path.pubkey, rev.sig, xdr2str (rev.msg)))
    return false;

  if (rsi
      && (!rsi->mkhostid (&hostid)
	  || (rev.msg.redirect->expire >= 
	      implicit_cast<sfs_time> (sfs_get_timenow ()))))
    return false;
  
  return true;
}

sfs_pathrevoke_w::sfs_pathrevoke_w (const sfs_pathrevoke &r) 
  : rev (r)
{
  si = sfs_servinfo_w::alloc (rev.msg.path);
  if (rev.msg.redirect)
    rsi = sfs_servinfo_w::alloc (rev.msg.redirect->hostinfo);
}

sfs_pathrevoke_w::~sfs_pathrevoke_w () 
{
  // XXX - what the heck...
  si = NULL;
  rsi = NULL;
}
