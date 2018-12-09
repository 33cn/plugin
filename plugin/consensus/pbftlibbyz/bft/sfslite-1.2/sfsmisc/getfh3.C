/* $Id: getfh3.C 1754 2006-05-19 20:59:19Z max $ */

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

#include "arpc.h"
#include "getfh3.h"
#include "sfsmisc.h"
#include "rxx.h"

static AUTH *auth_root = authunixint_create (myname () ? myname ().cstr ()
					     : "localhost", 0, 0, 0, NULL);

const strbuf &
strbuf_cat (const strbuf &sb, mountstat3 stat)
{
  switch (stat) {
  case MNT3_OK:
    return strbuf_cat (sb, "no error", false);
  case MNT3ERR_PERM:
    return strbuf_cat (sb, "Not owner", false);
  case MNT3ERR_NOENT:
    return strbuf_cat (sb, "No such file or directory", false);
  case MNT3ERR_IO:
    return strbuf_cat (sb, "I/O error", false);
  case MNT3ERR_ACCES:
    return strbuf_cat (sb, "Permission denied", false);
  case MNT3ERR_NOTDIR:
    return strbuf_cat (sb, "Not a directory", false);
  case MNT3ERR_INVAL:
    return strbuf_cat (sb, "Invalid argument", false);
  case MNT3ERR_NAMETOOLONG:
    return strbuf_cat (sb, "Filename too long", false);
  case MNT3ERR_NOTSUPP:
    return strbuf_cat (sb, "Operation not supported", false);
  case MNT3ERR_SERVERFAULT:
    return strbuf_cat (sb, "Server failure", false);
  }
  return sb << "Unknown error " << int (stat);
}

template<class T> inline str
stat2str (T xstat, clnt_stat stat)
{
  if (stat)
    return strbuf () << stat;
  else if (xstat)
    return strbuf () << xstat;
  else
    return NULL;
}

static void
splitpath (vec<str> &out, str in)
{
  const char *p = in.cstr ();
  const char *e = p + in.len ();
  const char *n;

  for (;;) {
    while (*p == '/')
      p++;
    for (n = p; n < e && *n != '/'; n++)
      ;
    if (n == p)
      return;
    out.push_back (str (p, n - p));
    p = n;
  }
    
}

struct lookup3obj {
  typedef callback<void, const nfs_fh3 *, const fattr3exp *, str>::ref cb_t;

  ref<aclnt> c;
  vec<str> cns;
  cb_t cb;
  lookup3res res;
  getattr3res ares;
  AUTH *auth;

  void getattr (clnt_stat stat) {
    if (stat || ares.status)
      (*cb) (NULL, NULL, stat2str (ares.status, stat));
    else {
      stompcast (ares);
      (*cb) (&res.resok->object, ares.attributes.addr (), NULL);
    }
    delete this;
  }

  void nextcn (const nfs_fh3 &fh, const fattr3exp *attrp) {
    if (!cns.size ()) {
      if (attrp) {
	(*cb) (&fh, attrp, NULL);
	delete this;
      }
      else
	c->call (NFSPROC3_GETATTR, &fh, &ares,
		 wrap (this, &lookup3obj::getattr), auth);
      return;
    }
      
    diropargs3 arg;
    arg.dir = fh;
    arg.name = cns.pop_front ();
    c->call (NFSPROC3_LOOKUP, &arg, &res,
	     wrap (this, &lookup3obj::getres), auth);
  }

  void getres (clnt_stat stat) {
    if (stat || res.status) {
      (*cb) (NULL, NULL, stat2str (res.status, stat));
      delete this;
    }
    else {
      stompcast (res);
      nextcn (res.resok->object,
	      res.resok->obj_attributes.present
	      ? res.resok->obj_attributes.attributes :
	      static_cast<fattr3exp *> (NULL));
    }
  }

  lookup3obj (ref<aclnt> c, const nfs_fh3 &start,
	      str path, cb_t cb, const authunix_parms *aup)
    : c (c), cb (cb) {
    if (aup) {
      auth = authopaque_create ();
      authopaque_set (auth, aup);
    }
    else
      auth = auth_root;
    splitpath (cns, path);
    res.resok->object = start;
    nextcn (start, NULL);
  }
  ~lookup3obj () { if (auth != auth_root) AUTH_DESTROY (auth); }
};

void
lookupfh3 (ref<aclnt> c, const nfs_fh3 &start, str path, lookup3obj::cb_t cb,
	   const authunix_parms *aup)
{
  vNew lookup3obj (c, start, path, cb, aup);
}


static rxx stripslash ("^(.*[^/]+)/+$");
static rxx dirbaserx ("^((/+[^/]+)*)/+([^/]+)$");

struct getfh3obj {
  typedef callback<void, const nfs_fh3 *, str>::ref cb_t;
  cb_t cb;

  ptr<aclnt> c;
  str mountpath;
  str lookuppath;
  mountres3 res;
  mountstat3 firsterr;
  str host;

  void fail (clnt_stat stat) { (*cb) (NULL, strbuf () << stat); delete this; }
  void fail (mountstat3 stat) { (*cb) (NULL, strbuf () << stat); delete this; }

  void lookupcb (const nfs_fh3 *fhp, const fattr3exp *, str err) {
    if (err)
      /* Note, err may be more descriptive than firsterr, but it will
       * also be confusing.  For example, suppose / is exported, /usr
       * is an unexported file system, and we are trying to get the
       * handle for /usr/local.  firsterr will be permission denied,
       * while err will say no such file or directory. */
      fail (firsterr);
    else {
      (*cb) (fhp, NULL);
      delete this;
    }
  }

  void gotnfsc (ptr<aclnt> nfsc, clnt_stat stat) {
    if (stat)
      fail (stat);
    else {
      nfs_fh3 fh;
      fh.data = res.mountinfo->fhandle;
      lookupfh3 (nfsc, fh, lookuppath, wrap (this, &getfh3obj::lookupcb));
    }
  }

  void gotfh3 (clnt_stat stat) {
    if (stat)
      fail (stat);
    else if (res.fhs_status) {
      if (!firsterr)
	firsterr = res.fhs_status;
      if (stripslash.match (mountpath))
	mountpath = stripslash[1];
      else if (dirbaserx.match (mountpath) && host) {
	if (lookuppath)
	  lookuppath = dirbaserx[3] << "/" << lookuppath;
	else
	  lookuppath = dirbaserx[3];
	mountpath = dirbaserx[1];
	if (!mountpath.len ())
	  mountpath = "/";
      }
      else {
	fail (firsterr);
	return;
      }
      domount ();
    }
    else {
      nfs_fh3 fh;
      fh.data = res.mountinfo->fhandle;
      if (lookuppath)
	aclntudp_create (host, 0, nfs_program_3,
			 wrap (this, &getfh3obj::gotnfsc));
      else {
	(*cb) (&fh, NULL);
	delete this;
      }
    }
  }

  void domount () {
    c->call (MOUNTPROC3_MNT, &mountpath, &res,
	     wrap (this, &getfh3obj::gotfh3), auth_root);
  }

  void gotc (ptr<aclnt> cc, clnt_stat stat) {
    c = cc;
    if (stat)
      fail (stat);
    else
      domount ();
  }
  getfh3obj (str host, dirpath path, cb_t cb)
    : cb (cb), mountpath (path), firsterr (MNT3_OK), host (host) {
    aclntudp_create (host, 0, mount_program_3, 
		     wrap (this, &getfh3obj::gotc));
  }
  getfh3obj (ref<aclnt> cc, dirpath path, cb_t cb)
    : cb (cb), c (cc), mountpath (path), firsterr (MNT3_OK) {
    domount ();
  }
};

void
getfh3 (str host, const str path, getfh3obj::cb_t cb)
{
  vNew getfh3obj (host, path, cb);
}

void
getfh3 (ref<aclnt> c, const str path, getfh3obj::cb_t cb)
{
  vNew getfh3obj (c, path, cb);
}

