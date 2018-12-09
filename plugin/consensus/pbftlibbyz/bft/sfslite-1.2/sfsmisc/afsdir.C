/* $Id: afsdir.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1998-2000 David Mazieres (dm@uun.org)
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


#include "afsnode.h"

ihash<const u_int32_t, afsdirentry, &afsdirentry::cookie,
  &afsdirentry::clink> cookietab;

static void pentry (afsnode *, const char *);

static u_int32_t
gencookie ()
{
  static u_int32_t cctr;
  while (++cctr < 3 || cookietab[cctr])
    ;
  return cctr;
}

afsdirentry::afsdirentry (afsdir *dir, const str &name, afsnode *node)
  : dir (dir), name (name), node (mkref (node)), cookie (gencookie ())
{
  node->addlink ();
  cookietab.insert (this);
  dir->entries.insert (this);
}

afsdirentry::~afsdirentry ()
{
  node->remlink ();
  dir->entries.remove (this);
  cookietab.remove (this);
}

static bool
xdr_putentry (XDR *x, afsnode *n, filename name, u_int32_t cookie)
{
  return
    // entry * (non-null):
    xdr_putint (x, 1)
    // u_int fileid:
    && xdr_putint (x, n->ino)
    // filename name:
    && xdr_filename (x, &name)
    // nfscookie cookie:
    && xdr_putint (x, cookie);
}

static bool
xdr_putentry3 (XDR *x, afsnode *n, filename name, u_int32_t cookie)
{
  return
    // entry * (non-null):
    xdr_putint (x, 1)
    // uint64 fileid:
    && xdr_puthyper (x, n->ino)
    // filename name:
    && xdr_filename (x, &name)
    // uint64 cookie:
    && xdr_puthyper (x, cookie);
}

/* For correctness, this must *NOT* return attributes (since we are
 * trying to defeat all kernel attribute and name caching).  Note also
 * that the revocation/redirect mechanism also relies on readdirplus
 * not returning file handles or attributes. */
static bool
xdr_putentryplus3 (XDR *x, afsnode *n, filename name, u_int32_t cookie)
{
  return
    // entry * (non-null):
    xdr_putint (x, 1)
    // uint64 fileid:
    && xdr_puthyper (x, n->ino)
    // filename name:
    && xdr_filename (x, &name)
    // nfscookie cookie:
    && xdr_puthyper (x, cookie)
    // post_op_attr name_attributes;
    && xdr_putint (x, 0)
    // post_op_fh3 name_handle;
    && xdr_putint (x, 0);
}

BOOL
afsdir::xdr (XDR *x, void *_sbp)
{
  /* When encoding . and .. in an XDR, we want to make sure the string
   * memory doesn't get freed before the XDR uses it.  */
  static const filename dot (".");
  static const filename dotdot ("..");

  svccb *sbp = static_cast<svccb *> (_sbp);
  const sfs_aid aid = sbp2aid (sbp);
  assert (x->x_op == XDR_ENCODE);
  const bool v2 = sbp->vers () == 2;

  afsdir *d;
  u_int32_t cookie;
  u_int32_t count;
  bool (*putentry) (XDR *, afsnode *, filename, u_int32_t);

  if (v2) {
    const readdirargs *arg = sbp->Xtmpl getarg<readdirargs> ();
    d = static_cast<afsdir *> (afsnode::fh2node (&arg->dir));
    cookie = getint (arg->cookie.base ());
    count = arg->count;
    putentry = xdr_putentry;
  }
  else if (sbp->proc () == NFSPROC3_READDIR) {
    const readdir3args *arg = sbp->Xtmpl getarg<readdir3args> ();
    d = static_cast<afsdir *> (afsnode::fh3node (&arg->dir));
    cookie = arg->cookie;
    count = arg->count;
    putentry = xdr_putentry3;
  }
  else if (sbp->proc () == NFSPROC3_READDIRPLUS) {
    const readdirplus3args *arg = sbp->Xtmpl getarg<readdirplus3args> ();
    d = static_cast<afsdir *> (afsnode::fh3node (&arg->dir));
    cookie = arg->cookie;
    count = arg->dircount;
    putentry = xdr_putentryplus3;
  }
  else
    return xdr_putint (x, NFS3ERR_NOTSUPP);
  if (!d)
    return xdr_putint (x, NFSERR_STALE);

  afsdirentry *e = NULL;
  if (cookie >= 3 && (!(e = cookietab[cookie]) || !d->entryok (e, aid))) {
    warn ("afsdir::xdr: bad cookie 0x%x\n", cookie);
    return xdr_putint (x, v2 ? EINVAL : NFS3ERR_BAD_COOKIE);
  }

  if (!xdr_putint (x, NFS_OK))
    return false;
  if (!v2) {
    post_op_attr poa;
    d->mkpoattr (poa, aid);
    if (!xdr_post_op_attr (x, &poa) || !xdr_puthyper (x, 0))
      return false;
  }

  switch (cookie) {
  case 0:
    if (!putentry (x, d, dot, 1))
      return false;
  case 1:
    if (!putentry (x, d->parent, dotdot, 2))
      return false;
  case 2:
    e = d->firstentry (aid);
    break;
  default:
    e = d->nextentry (e, aid);
    break;
  }

  for (; e && XDR_GETPOS (x) + 24 + e->name.len () <= count;
       e = d->nextentry (e, aid))
    if (!putentry (x, e->node, e->name, e->cookie))
      return false;

  return xdr_putint (x, 0)	// NULL entry *
    && xdr_putint (x, !e);	// bool eof
}

afsdir::afsdir (afsdir *p)
  : afsnode (NF3DIR), parent (p ? p : this)
{
  mtime.seconds = mtime.nseconds = 0;
  nlinks += 2;
}

afsdir::~afsdir ()
{
  entries.traverse (wrap (&afsdirentry::del));
}

afsnode *
afsdir::lookup (const str &name, sfs_aid)
{
  if (name == ".")
    return this;
  if (name == "..")
    return parent;
  if (afsdirentry *e = entries[name])
    return e->node;
  return NULL;
}

bool
afsdir::entryok (afsdirentry *e, sfs_aid)
{
  return e->dir == this;
}

afsdirentry *
afsdir::firstentry (sfs_aid)
{
  return entries.first ();
}

afsdirentry *
afsdir::nextentry (afsdirentry *e, sfs_aid)
{
  return entries.next (e);
}

bool
afsdir::link (afsnode *node, const str &name)
{
  if (entries[name])
    return false;
  vNew afsdirentry (this, name, node);
  bumpmtime ();
  return true;
}

bool
afsdir::unlink (const str &name)
{
  afsdirentry *e = entries[name];
  if (e) {
    bumpmtime ();
    delete e;
  }
  return e;
}

ptr<afsdir>
afsdir::mkdir (const str &name)
{
  ref<afsdir> d = allocsubdir (this);
  if (!link (d, name))
    return NULL;
  addlink ();
  return d;
}

ptr<afslink>
afsdir::symlink (const str &contents, const str &name)
{
  ref<afslink> lnk = afslink::alloc (contents);
  if (!link (lnk, name))
    return NULL;
  return lnk;
}

#if 0
/* Returns true for "." and ".." */
static bool
isddot (const char *name)
{
  if (name[0] != '.')
    return false;
  if (name[1] == '\0')
    return true;
  if (name[1] != '.')
    return false;
  if (name[2] != '\0')
    return false;
  return true;
}
#endif

void
afsdir::nfs_lookup (svccb *sbp, str name)
{
  lookup_reply (sbp, lookup (name, sbp2aid (sbp)));
}

static char
tc (ftype3 t)
{
  switch (t) {
  case NF3REG:
    return '-';
  case NF3DIR:
    return 'd';
  case NF3LNK:
    return 'l';
  default:
    return '?';
  }
}
static char *
mc (u_int mode, char s[10])
{
  char *p = s;
  for (int i = 0; i < 3; i++) {
    *p++ = mode & 4 ? 'r' : '-';
    *p++ = mode & 2 ? 'w' : '-';
    *p++ = mode & 1 ? 'x' : '-';
    mode >>= 3;
  }
  *p = '\0';
  return s;
}
static void
pentry (afsnode *n, const char *name)
{
  fattr f;
  n->mkfattr (&f, NULL);

  time_t mtime = time (NULL);
  int year = gmtime (&mtime)->tm_year;
  mtime = f.mtime.seconds;
  tm *tmp = gmtime (&mtime);
  char t[24];
  if (!strftime (t, sizeof (t),
		year <= tmp->tm_year ? "%b %e %H:%M" : "%b %e  %Y",
		tmp))
    panic ("strftime overflow\n");

  char m[10];
  warnx ("%6d %c%s %2d %5d %5d %7d %s %s\n", (u_int32_t) n->ino,
	 tc (n->type), mc (f.mode, m),
	 n->getnlinks (), f.uid, f.gid, f.size, t, name);
}

void
afsdir::ls ()
{
  warnx ("files in directory %d:\n", (u_int32_t) ino);
  pentry (this, ".");
  pentry (parent, "..");
  for (afsdirentry *e = firstentry (NULL); e; e = nextentry (e, NULL))
    pentry (e->node, e->name);
}
