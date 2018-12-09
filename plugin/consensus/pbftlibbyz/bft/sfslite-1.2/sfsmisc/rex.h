// -*-c++-*-
/* $Id: rex.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2001-2003 Michael Kaminsky (kaminsky@lcs.mit.edu)
 * Copyright (C) 2000-2001 Eric Peterson (ericp@lcs.mit.edu)
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

#ifndef _SFSMISC_REX_H_
#define _SFSMISC_REX_H_ 1

#include "sfsmisc.h"
#include "sfsconnect.h"
#include "qhash.h"
#include "crypt.h"
#include "aios.h"
#include "rex_prot.h"
#include "agentconn.h"

TYPE2STRUCT(, sfsstat);

class rexchannel;

class rexfd : public virtual refcount {
protected:
  rexchannel *pch;
  ptr<aclnt> proxy;
  u_int32_t channo;
  int fd;

public:
  // these implement null fd behavior, so you'll probably want to override them
  static bool garbage_bool;
  virtual void abort ();
  virtual void data (svccb *sbp);
  virtual void newfd (svccb *sbp) { sbp->replyref (false); }
  virtual void exited () {};	/* called when remote module exits */
  rexfd (rexchannel *pch, int fd);
  virtual ~rexfd ();
};

class unixfd : public rexfd {
protected:
  int localfd_in;
  int localfd_out;
  int rsize;
  ptr<aios> paios_out;

  const bool unixsock;
  bool weof;
  bool reof;
  bool shutrdonexit;
  cbv closecb;

  void update_connstate (int how, int error = 0);
public:
  void datacb (int nbytes, ptr<bool> okp, clnt_stat) {
    assert (nbytes <= rsize);
    bool stalled = *okp && !reof && rsize >= hiwat;
    rsize -= nbytes;
    if (stalled && rsize < hiwat)
      fdcb (localfd_in, selread, wrap (this, &unixfd::rcb));
    if (!*okp)
      update_connstate (SHUT_RDWR);
  }
protected:
  void newfdcb (int fdrecved, ptr<rex_newfd_res> resp, clnt_stat err);

public:
  enum { hiwat = 0x20000 };

  virtual void readeof ();
  virtual void rcb ();
  virtual void newfd (svccb *sbp);
  virtual void data (svccb *sbp);
  virtual void abort () {
    update_connstate (SHUT_RDWR);
  }
  void exited () {
    if (shutrdonexit)
      update_connstate (SHUT_RD);
  }
  
  unixfd (rexchannel *pch, int fd, int localfd_in, int localfd_out = -1,
	  bool noclose = false, bool shutrdonexit = false,
	  cbv closecb = cbv_null);

  virtual ~unixfd () {
    if (localfd_in >= 0) {
      fdcb (localfd_in, selread, NULL);
      if (!paios_out || (localfd_out >= 0 && localfd_in != localfd_out))
	close (localfd_in);
    }
    if (paios_out)
      paios_out->flush ();
    closecb ();
  }
};

class rexsession;

class rexchannel {
  vec <ptr<rexfd> > vfds;
  int fdc;
  void deref_vfds ();

protected:
  rexsession *sess;
  ptr<aclnt> proxy;
  u_int32_t channo;
  bool got_exit_cb;

  int initnfds;
  vec <str> command;
  
  friend class rexsession;

  virtual void quit ();
  virtual void abort ();
  virtual void madechannel (int error) {};

  void channelinit (u_int32_t chnumber, ptr<aclnt> proxyaclnt, int error);

  virtual void data (svccb *sbp);
  virtual void newfd (svccb *sbp);
  virtual void exited (int status);

public:
  void insert_fd (int fdn, ptr<rexfd> rfd);
  void remove_fd (int fdn);  

  int get_initnfds () { return initnfds; }
  const vec<str> &get_cmd () { return command; }
  u_int32_t get_channo () { return channo; }
  ptr<aclnt> get_proxy () { return proxy; }
      
  rexchannel (rexsession *sess, int initialfdcount, vec <str> command)
    : fdc (0), sess (sess), got_exit_cb (false), initnfds (initialfdcount),
      command(command) {
    // warn << "--reached rexchannel: fdc = " << fdc << "\n";
  }
  virtual ~rexchannel () {
    // warn << "--reached ~rexchannel\n";
    deref_vfds ();  // allow fd destructors to run, possibly calling remove_fd
  }
};

class rexsession {
  bool verbose;

  ptr<axprt_crypt> proxyxprt;
  sfs_seqno seqno;
  vec<char> secretid;
  bool resumable;
  bool suspended;

  //todo : make this non-refcounted pointer
  qhash<u_int32_t, ref<rexchannel> > channels;
  qhash<u_int32_t, ref<rexchannel> > channels_pending_exit;
  int cchan;
  int channelspending;

  callback<void>::ptr endcb;
  callback<bool>::ptr failcb;
  callback<bool>::ptr timeoutcb;

  str schost;

  ifchgcb_t *ifchg;
  bool silence_tmo_enabled;
  time_t silence_tmo_min;
  timecb_t *silence_check_cb;
  callbase *probe_call;

  ref<asrv_resumable> rexserv;
public:
  ref<aclnt_resumable> proxy;
  time_t last_heard;

private:
  void rexcb_dispatch (svccb *sbp);
  bool fail ();
  void ifchg_cb_set ();
  void ifchg_cb_clear ();
  void rpc_call_hook ();
  void rpc_recv_hook ();
  void silence_tmo_init ();
  void silence_check ();
  void probed (clnt_stat err);

  void resumed (ptr<axprt_crypt> xprt, ref<bool> resp, ptr<aclnt> proxytmp,
                callback<void, bool>::ref cb, clnt_stat err);
  void madechannel (ptr<rex_mkchannel_res> resp, 
		    ptr<rexchannel> newchan, clnt_stat err);
  void quitcaller (const u_int32_t &chno, ptr<rexchannel> pchan) {
    pchan->quit ();
  }
  void abortcaller (const u_int32_t &chno, ptr<rexchannel> pchan) {
    pchan->abort ();
  }

public:
  // gets called when all channels close or we get EOF from proxy
  void setendcb (cbv endcb) { rexsession::endcb = endcb; }
  void set_verbose (bool status) { verbose = status; }
  void makechannel (ptr<rexchannel> newchan, rex_env env = rex_env ());
  void remove_chan (int channo);
  void remove_chan_pending_exit (int channo);

  // informs all channels that client wants to quit.  default
  // behavior is to kill remote module.
  void quit () { channels.traverse (wrap (this, &rexsession::quitcaller)); }

  // calls the abort member function of every channel, which should blow
  // all the channels away
  void abort () {
    endcb = NULL;
    channels.traverse (wrap (this, &rexsession::abortcaller));
  }

  rexsession (str schostname, ptr<axprt_crypt> proxyxprt, vec<char> &secretid,
              callback<bool>::ptr failcb, callback<bool>::ptr timeoutcb = NULL,
      	      bool verbose = false, bool resumable_mode = false);
  ~rexsession ();

  void suspend ();
  callbase *resume (ptr<axprt_crypt> xprt, sfs_seqno seqno,
                    callback<void, bool>::ref cb);
  void setresumable (bool mode);
  bool getresumable () const { return resumable; }

  void set_call_hook (cbv::ptr cb) {
    proxy->set_send_hook (cb);
  }
  void set_recv_hook (cbv::ptr cb) {
    proxy->set_recv_hook (cb);
    rexserv->set_recv_hook (cb);
  }

  void silence_tmo_reset ();
  void silence_tmo_enable ();
  void silence_tmo_disable ();
  callbase *ping (callback<void, clnt_stat>::ref, time_t timeout = 0);
};

#endif /* _SFSMISC_REX_H_ */
