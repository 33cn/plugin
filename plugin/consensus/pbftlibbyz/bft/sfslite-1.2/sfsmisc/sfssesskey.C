/* $Id: sfssesskey.C 1754 2006-05-19 20:59:19Z max $ */

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

/*
 * A Note on Protocol Versions:  This code can currently support versions
 * 1 and 2 of the SFSPROC_ENCRYPT RPC call.  Eventually, version 1 will
 * be phased out, and version 2 should be able to accommodate future
 * changes to the session key encryption protocol -- including a change
 * in cryptosystem.  All version-dependent code involves the variable
 * "pvers".  Future maintainers who wish to wipe support for version 1
 * should look for this variable in the code. -MK 08.01.2002
 */

#include "sfsmisc.h"
#include "crypt.h"
#include "hashcash.h"
#include "sfscrypt.h"
#include "sfssesscrypt.h"

#ifdef MAINTAINER
bool sfs_nocrypt = safegetenv ("SFS_NOCRYPT");
#endif /* MAINTAINER */

void
sfs_get_sesskey (sfs_hash *ksc, sfs_hash *kcs,
		 const sfs_servinfo &si, const sfs_kmsg *smsg, 
		 const sfs_connectinfo &ci, 
		 const sfs_kmsg *cmsg)
{
  sfs_sesskeydat kdat;

  kdat.type = SFS_KSC;
  kdat.si = si;
  kdat.sshare = smsg->ksc_share;
  kdat.ci = ci;
  kdat.cshare = cmsg->ksc_share;
  sha1_hashxdr (ksc->base (), kdat, true);

  kdat.type = SFS_KCS;
  kdat.sshare = smsg->kcs_share;
  kdat.cshare = cmsg->kcs_share;
  sha1_hashxdr (kcs->base (), kdat, true);

  bzero (&kdat.sshare, sizeof (kdat.sshare));
  bzero (&kdat.cshare, sizeof (kdat.cshare));
}

void
sfs_get_sessid (sfs_hash *sessid, const sfs_hash *ksc, const sfs_hash *kcs)
{
  sfs_sessinfo si;
  si.type = SFS_SESSINFO;
  si.ksc = *ksc;
  si.kcs = *kcs;
  sha1_hashxdr (sessid->base (), si, true);

  bzero (si.ksc.base (), si.ksc.size ());
  bzero (si.kcs.base (), si.kcs.size ());
}

void
sfs_get_authid (sfs_hash *authid, sfs_service service, sfs_hostname name,
		const sfs_hash *hostid, const sfs_hash *sessid,
		sfs_authinfo *authinfo)
{
  sfs_authinfo aui;
  aui.type = SFS_AUTHINFO;
  aui.service = service;
  aui.name = name;
  aui.hostid = *hostid;
  aui.sessid = *sessid;
  if (authinfo) *authinfo = aui;
  sha1_hashxdr (authid->base (), aui);
}

static void
set_random_key (axprt_crypt *cx, sfs_hash *sessid)
{
  sfs_hash ksc, kcs;
  rnd.getbytes (ksc.base (), ksc.size ());
  rnd.getbytes (kcs.base (), kcs.size ());

  cx->encrypt (ksc.base (), ksc.size (), kcs.base (), kcs.size ());

  if (sessid)
    sfs_get_sessid (sessid, &ksc, &kcs);
  bzero (ksc.base (), ksc.size ());
  bzero (kcs.base (), kcs.size ());
}

static inline axprt_crypt *
xprt2crypt (axprt *x)
{
  // XXX - dynamic_cast is busted in egcs
  axprt_crypt *cx = static_cast<axprt_crypt *> (&*x);
  // assert (typeid (*cx) == typeid (refcounted<axprt_crypt>));
  return cx;
}


void
sfs_server_crypt (svccb *sbp, sfspriv *sk,
		  const sfs_connectinfo &ci, 
		  ref<const sfs_servinfo_w> si,
		  sfs_hash *sessid, const sfs_hashcharge &charge,
		  axprt_crypt *cx, int pvers)
{
  assert (sbp->prog () == SFS_PROGRAM);
  assert ((pvers == 1 && sbp->proc () == SFSPROC_ENCRYPT) ||
	  (pvers == 2 && sbp->proc () == SFSPROC_ENCRYPT2));
  if (!cx)
    cx = xprt2crypt (sbp->getsrv ()->xprt ());
  if (!ci.civers) {
    warn << "sfs_server_crypt: client called encrypt before connect?\n";
    sbp->reject (PROC_UNAVAIL);
    set_random_key (cx, sessid);
    return;
  }

  ptr<sfspub> clntpub;
  sfs_encryptarg *arg = NULL;
  sfs_encryptarg2 *arg2 = NULL;
    
  if (pvers == 1) {
    arg = sbp->Xtmpl getarg<sfs_encryptarg> ();
    clntpub = sfscrypt.alloc (arg->pubkey, SFS_ENCRYPT);
  } else {
    arg2 = sbp->Xtmpl getarg<sfs_encryptarg2> ();
    clntpub = sfscrypt.alloc (arg2->pubkey, SFS_ENCRYPT);
  }
  
  sfs_kmsg smsg;
  sfs_hash hostid;
  si->mkhostid (&hostid);
  const sfs_hashpay &pay = (pvers == 1 ? arg->payment : arg2->payment);
  
  if (!hashcash_check (pay.base (), hostid.base (), charge.target.base (), 
		       charge.bitcost)) {
    warn << "payment doesn't match charge\n";
    sbp->reject (GARBAGE_ARGS);
    set_random_key (cx, sessid);
    return;
  }

  str s;
  if (!clntpub->check_keysize (&s)) {
    warn << s << "\n";
    sbp->reject (GARBAGE_ARGS);
    set_random_key (cx, sessid);
    return;
  }

  rnd.getbytes (&smsg, sizeof (smsg));
  sfs_encryptres2 res2;
  sfs_encryptres res;
  sfs_ctext ct;
  sfs_ctext2 ct2;
  if (!(pvers == 1 ? clntpub->encrypt (&res, wstr (&smsg, sizeof (smsg))) :
	clntpub->encrypt (&res2, wstr (&smsg, sizeof (smsg))))) {
    warn << "could not encrypt with client's public key\n";
    sbp->reject (GARBAGE_ARGS);
    set_random_key (cx, sessid);
    return;
  }

  // Can't refer to arg or arg2 after call to sbp->reply, 
  // since sbp is destroyed therein.  
  if (pvers == 1) {
    ct = arg->kmsg;
    sbp->reply (&res);
  } else {
    ct2 = arg2->kmsg;
    sbp->reply (&res2);
  }
  
  sfs_hash ksc, kcs;
  str cmsgptxt;

  bool valid = false;
  if (pvers == 1) valid = sk->decrypt (ct,  &cmsgptxt);
  else if (pvers == 2) valid = sk->decrypt (ct2, &cmsgptxt, sizeof (sfs_kmsg));
  if (!valid) {
    set_random_key (cx, sessid);
    return;
  }

  sfs_get_sesskey (&ksc, &kcs, si->get_xdr (), &smsg, ci, 
		   sfs_get_kmsg (cmsgptxt));
  if (sessid)
    sfs_get_sessid (sessid, &ksc, &kcs);

#ifdef MAINTAINER
  if (!sfs_nocrypt)
#endif /* MAINTAINER */
    cx->encrypt (ksc.base (), ksc.size (), kcs.base (), kcs.size ());

  bzero (ksc.base (), ksc.size ());
  bzero (kcs.base (), kcs.size ());
  bzero (&smsg, sizeof (smsg));
}

struct sfs_client_crypt_state {
  typedef callback<void, const sfs_hash *>::ref cb_t;
  int vers;
  sfs_kmsg cmsg;
  sfs_encryptres res;
  sfs_encryptres2 res2;
  ptr<axprt> x;
  ref<sfspriv> csk;  /* client secret key */
  ptr<sfspub>  spk;  /* server public key */
  ref<const sfs_servinfo_w> si;
  sfs_connectinfo ci;
  sfs_hash hostid;
  sfs_hashcharge charge;
  str hostname;
  cb_t cb;
  sfs_client_crypt_state (ref<sfspriv> cs, ref<const sfs_servinfo_w> ssi,
			  const sfs_connectinfo &cci, cb_t c)
    : csk (cs) , si (ssi), ci (cci), cb (c) {}
};

static void
sfs_client_crypt_cb (ptr<aclnt> c, ref<sfs_client_crypt_state> st, 
		     int pvers, clnt_stat err)
{
  if (err) {
    warnx << st->si->get_hostname () << ": negotiating session key: "
	  << err << "\n";
    (*st->cb) (NULL);
    return;
  }

  sfs_hash ksc, kcs;
  str smsgptxt;

  if (!(pvers == 1 ? st->csk->decrypt (st->res, &smsgptxt) :
	st->csk->decrypt (st->res2, &smsgptxt, sizeof (sfs_kmsg)))) {
    /* Bad ciphertext -- just start encrypting with a random key */
    rnd.getbytes (ksc.base (), ksc.size ());
    rnd.getbytes (kcs.base (), kcs.size ());
  }
  else
    sfs_get_sesskey (&ksc, &kcs, st->si->get_xdr (), sfs_get_kmsg (smsgptxt),
		     st->ci, &st->cmsg);

  sfs_hash sessid;
  sfs_get_sessid (&sessid, &ksc, &kcs);

#ifdef MAINTAINER
  if (!sfs_nocrypt) 
#endif /* MAINTAINER */
    xprt2crypt (st->x)->encrypt (&kcs, sizeof (kcs), &ksc, sizeof (ksc));

  bzero (&ksc, sizeof (ksc));
  bzero (&kcs, sizeof (kcs));
  bzero (&st->cmsg, sizeof (st->cmsg));

  (*st->cb) (&sessid);
}

static callbase *
sfs_client_crypt_call (ptr<aclnt> c, ref<sfs_client_crypt_state> st, int pvers)
{
  sfs_encryptarg arg;
  sfs_encryptarg2 arg2;

  if (pvers == 1) {
    hashcash_pay (arg.payment.base (), st->hostid.base (), 
		  st->charge.target.base (), st->charge.bitcost);
    if (!st->spk->encrypt (&arg.kmsg, wstr (&st->cmsg, sizeof (st->cmsg))) 
	|| !st->csk->export_pubkey (&arg.pubkey)) {
      warn << st->hostname << ": For SFSPROC_ENCRYPT, must have Rabin "
	   << "cryptosystem on both ends\n";
      goto fail;
    }
    /* The timed call is for forward secrecy.  clntkey gets used for
     * other session keys, so we don't want to sit on it forever. */
    return c->timedcall (600, SFSPROC_ENCRYPT, &arg, &(st->res),
			 wrap (sfs_client_crypt_cb, c, st, pvers));
  } else {
    hashcash_pay (arg2.payment.base (), st->hostid.base (),
		  st->charge.target.base (), st->charge.bitcost);
    if (!st->spk->encrypt (&arg2.kmsg, wstr (&st->cmsg, sizeof (st->cmsg)))) {
      warn << st->hostname << ": Encrypt with server public key failed\n";
      goto fail;
    }
    if (!st->csk->export_pubkey (&arg2.pubkey)) {
      warn << st->hostname << ": cannot retrieve client public key\n";
      goto fail;
    }
    return c->timedcall (600, SFSPROC_ENCRYPT2, &arg2, &(st->res2),
			 wrap (sfs_client_crypt_cb, c, st, pvers));
  }

 fail:
  set_random_key (xprt2crypt (c->xprt ()), NULL);
  (*(st->cb)) (NULL);
  return NULL;
}

callbase *
sfs_client_crypt (ptr<aclnt> c, ptr<sfspriv> clntpriv,
		  const sfs_connectinfo &ci, 
		  const sfs_connectok &cres,
		  ref<const sfs_servinfo_w> si,
		  callback<void, const sfs_hash *>::ref cb,
		  ptr<axprt_crypt> cx)
{
  assert (c->rp.progno == SFS_PROGRAM && c->rp.versno == SFS_VERSION);
  int pvers = si->get_vers ();
  str s;
  ref<sfs_client_crypt_state> st
    = New refcounted<sfs_client_crypt_state> (clntpriv, si, ci, cb);

  st->hostname = si->get_hostname (); 
  if (!(st->spk = sfscrypt.alloc (si->get_pubkey (), SFS_ENCRYPT))) {
    warn << st->hostname << ": bad public key\n";
    goto fail;
  }

  if (!si->mkhostid_client (&(st->hostid))) { 
    warn << st->hostname << ": bad hostinfo\n";
    goto fail;
  }
  if (!st->spk->check_keysize (&s)) {
    warn << st->hostname << ": " << s << "\n";
    goto fail;
  }
  if (cres.charge.bitcost > sfs_maxhashcost) {
    warn << st->hostname << ": hashcash charge too great\n";
    goto fail;
  }
  st->charge = cres.charge;

  if (cx)
    st->x = cx;
  else
    st->x = c->xprt ();
  rnd.getbytes (&st->cmsg, sizeof (st->cmsg));

  return sfs_client_crypt_call (c, st, pvers);

 fail:
  set_random_key (xprt2crypt (c->xprt ()), NULL);
  (*cb) (NULL);
  return NULL;
}

