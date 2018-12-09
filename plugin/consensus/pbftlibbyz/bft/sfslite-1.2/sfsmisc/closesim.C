/* $Id: closesim.C 2679 2007-04-04 20:53:20Z max $ */

/*
 *
 * Copyright (C) 2000 David Mazieres (dm@uun.org)
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

#include "nfsserv.h"
#include "itree.h"
#include "nfstrans.h"

enum {
  closesim_minlife = 60,	// Lifetime of unopen file handles
  closesim_openlife = 300,	// Lifetime of opened file handles
};

struct closesim;

struct fhtimer {
  const nfs_fh3 fh;
  closesim *const serv;
  time_t expire;
  bool opened;
  itree_entry<fhtimer> timelink;
  ihash_entry<fhtimer> fhlink;
  fhtimer (closesim *s, const nfs_fh3 &f) : fh (f), serv (s), opened (false) {}
};

struct fileq {
  int minopen;
  bool cleanlock;
  timecb_t *tmo;
  itree<time_t, fhtimer, &fhtimer::expire, &fhtimer::timelink> timeq;
  explicit fileq (int mo) : minopen (mo), cleanlock (false), tmo (NULL) {}
  ~fileq () { timecb_remove (tmo); }
  void fhclean (bool istmo = false);
};

struct closesim : public nfsserv {
  fileq *const fq;
  ihash<const nfs_fh3, fhtimer, &fhtimer::fh, &fhtimer::fhlink> fhtab;

  fhtimer *fhalloc (const nfs_fh3 &fh);
  void fhfree (fhtimer *fht);
  void fhopen (fhtimer *fht);
  void fhtouch (fhtimer *fht);
  void dofh (const nfs_fh3 *fhp, bool openit = false);

  closesim (fileq *f, ref<nfsserv> n);
  ~closesim ();

  void getcall (nfscall *nc);
  void getreply (nfscall *nc);
};

//static fileq defaultq (1024);
static fileq defaultq (0);

DUMBTRAVERSE (closesim)
inline bool
rpc_traverse (closesim &serv, nfs_fh3 &fh)
{
  serv.dofh (&fh);
  return true;
}

static void
fhcleancb (nfs_fh3 *deleteme, nfsstat3 *)
{
  delete deleteme;
}

void
fileq::fhclean (bool istmo)
{
  if (istmo)
    tmo = NULL;
  if (cleanlock)
    return;

  cleanlock = true;
  fhtimer *fht = NULL;
  if (minopen < 0 && (fht = timeq.first ()) && fht->expire < sfs_get_timenow())
    do {
      fhtimer *nfht = timeq.next (fht);
      nfs_fh3 *fhp = New nfs_fh3 (fht->fh);
      closesim *serv = fht->serv;
      serv->fhfree (fht);
      vNew nfscall_cb<NFSPROC_CLOSE> (NULL, fhp,
				      wrap (fhcleancb, fhp), serv);
      fht = nfht;
    } while (minopen < 0 && fht->expire < sfs_get_timenow());
  cleanlock = false;

  if (!tmo && minopen < 0 && fht)
    tmo = timecb (min<time_t> (sfs_get_timenow() + 10, fht->expire),
		  wrap (this, &fileq::fhclean, true));
}


fhtimer *
closesim::fhalloc (const nfs_fh3 &fh)
{
  fhtimer *fht = New fhtimer (this, fh);
  fht->expire = sfs_get_timenow() + closesim_minlife;
  fq->minopen--;
  fq->timeq.insert (fht);
  fhtab.insert (fht);
  fq->fhclean ();
  return fht;
}

void
closesim::fhfree (fhtimer *fht)
{
  assert (fht->serv == this);
  fq->timeq.remove (fht);
  fhtab.remove (fht);
  fq->minopen++;
  delete fht;
}

void
closesim::fhopen (fhtimer *fht)
{
  fht->opened = true;
  fhtouch (fht);
}

void
closesim::fhtouch (fhtimer *fht)
{
  fq->timeq.remove (fht);
  fht->expire = sfs_get_timenow() + 
    (fht->opened ? closesim_openlife : closesim_minlife);
  fq->timeq.insert (fht);
}

void
closesim::dofh (const nfs_fh3 *fhp, bool openit)
{
  if (fhtimer *fht = fhtab [*fhp]) {
    if (openit)
      fhopen (fht);
    else
      fhtouch (fht);
  }
  else {
    fht = fhalloc (*fhp);
    if (openit)
      fhopen (fht);
  }
}

closesim::closesim (fileq *f, ref<nfsserv> n)
  : nfsserv (n), fq (f)
{
}

closesim::~closesim ()
{
  fhtab.traverse (wrap (this, &closesim::fhfree));
}

void
closesim::getcall (nfscall *nc)
{
  switch (nc->proc ()) {
  case NFSPROC3_NULL:
    break;
  case NFSPROC3_ACCESS:
    dofh (nc->getfh3arg (), true);
    break;
  case NFSPROC3_RENAME:
    dofh (nc->getfh3arg ());
    dofh (&nc->Xtmpl getarg<rename3args> ()->to.dir);
    break;
  case NFSPROC3_LINK:
    dofh (nc->getfh3arg ());
    dofh (&nc->Xtmpl getarg<link3args> ()->link.dir);
    break;
  default:
    dofh (nc->getfh3arg ());
    break;
  }

  mkcb (nc);
}

void
closesim::getreply (nfscall *nc)
{
#define doproc(proc)							      \
 case proc:								      \
   rpc_traverse (*this, *static_cast<nfs3proc<proc>::res_type *> (nc->resp)); \
   break

  if (!nc->xdr_res)
    switch (nc->proc ()) {
      doproc (NFSPROC3_LOOKUP);
      doproc (NFSPROC3_CREATE);
      doproc (NFSPROC3_MKDIR);
      doproc (NFSPROC3_SYMLINK);
      doproc (NFSPROC3_MKNOD);
      doproc (NFSPROC3_READDIRPLUS);
    }

#undef doproc

  nc->sendreply ();
}

ref<nfsserv>
close_simulate (ref<nfsserv> ns)
{
  return New refcounted<closesim> (&defaultq, ns);
}
