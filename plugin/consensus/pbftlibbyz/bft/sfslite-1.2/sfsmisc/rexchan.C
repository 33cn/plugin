/* $Id: rexchan.C 1754 2006-05-19 20:59:19Z max $ */

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

/* This code is shared by the agent (agent/agentrex.C) and rex
 * (rex/rex.C), both of which have occasion to connect to rexd.
 */

#include "rex.h"

bool rexfd::garbage_bool;

#ifndef O_ACCMODE
# define O_ACCMODE (O_RDONLY|O_WRONLY|O_RDWR)
#endif /* !O_ACCMODE */

bool
is_fd_wronly (int fd)
{
  int n;
  if ((n = fcntl (fd, F_GETFL)) < 0)
    fatal ("fcntl failed on fd = %d\n", fd);
  return (n & O_ACCMODE) == O_WRONLY;
}

bool
is_fd_rdonly (int fd)
{
  int n;
  if ((n = fcntl (fd, F_GETFL)) < 0)
    fatal ("fcntl failed on fd = %d\n", fd);
  return (n & O_ACCMODE) == O_RDONLY;
}

rexfd::rexfd (rexchannel *pch, int fd)
  : pch (pch), proxy (pch->get_proxy ()), channo (pch->get_channo ()),
    fd (fd)
{
/*   warn << "--reached rexfd\n"; */
  if (fd < 0)
    fatal ("attempt to create negative fd: %d\n", fd);
  pch->insert_fd (fd, mkref (this));
}

rexfd::~rexfd () { 
/*   warn << "--reached ~rexfd\n"; */
  rex_int_arg arg;
  arg.channel = channo;
  arg.val = fd;
  proxy->call (REX_CLOSE, &arg, &garbage_bool, aclnt_cb_null);

// NOTE: We don't call remove_fd() here but leave it up to the derived
// class.  Calling the remove_fd() function removes the fd from the
// channel's list which is the last reference to it which causes it to
// be deleted which causes the destructor to be called which causes
// this base class destructor to be called which would then call
// remove_fd() again which finally results in an error because we
// already removed it once */
}

void
rexfd::abort ()
{
  rex_payload payarg;
  payarg.channel = channo;
  payarg.fd = fd;
  proxy->call (REX_DATA, &payarg, &garbage_bool, aclnt_cb_null);
  
  pch->remove_fd (fd); 
}

void
rexfd::data (svccb *sbp)
{
  rex_payload *argp = sbp->Xtmpl getarg<rex_payload> ();
  if (!argp->data.size ()) {
    rex_payload payarg;
    payarg.channel = channo;
    payarg.fd = fd;
    payarg.data.set ((char *)NULL, 0);
    proxy->call (REX_DATA, &payarg, &garbage_bool, aclnt_cb_null);
    
    pch->remove_fd (fd); 
  }
  sbp->replyref (true);
}

void
unixfd::update_connstate (int how, int)
{
  if (localfd_in < 0)
    return;
  
  if (how == SHUT_WR)
    weof = true; 
  else if (how == SHUT_RD)
    reof = true;
  else
    weof = reof = true;

  /* XXX - what about SHUT_RDWR?  Shouldn't that sendeof, too? */
  if (how == SHUT_WR)
    paios_out->sendeof ();
  
  if (weof && reof) {
    localfd_in = -1;
    pch->remove_fd (fd);
  }
}

void
unixfd::readeof ()
{
  if (!reof) {
    rex_payload payarg;
    payarg.channel = channo;
    payarg.fd = fd;
    proxy->call (REX_DATA, &payarg, &garbage_bool, aclnt_cb_null);
    update_connstate (SHUT_RD);
  }
}

void 
unixfd::rcb ()
{
  if (reof) {
    fdcb (localfd_in, selread, NULL);
    return;
  }

  char buf[16*1024];
  int fdrecved = -1;
  ssize_t n;
  if (unixsock)
    n = readfd (localfd_in, buf, sizeof (buf), &fdrecved);
  else
    n = read (localfd_in, buf, sizeof (buf));

  if (n < 0) {
    if (errno != EAGAIN)
      abort ();
    return;
  }

  if (!n) {
    readeof ();
    return;
  }
  else {
    rex_payload arg;
    arg.channel = channo;
    arg.fd = fd;
    arg.data.set (buf, n);

    if (fdrecved >= 0) {
      close_on_exec (fdrecved);
      rex_newfd_arg arg;
      arg.channel = channo;
      arg.fd = fd;
      ref<rex_newfd_res> resp (New refcounted<rex_newfd_res> (false));
      proxy->call (REX_NEWFD, &arg, resp,
		   wrap (mkref (this), &unixfd::newfdcb, fdrecved, resp));
    }

    ref<bool> pres (New refcounted<bool> (false));
    rsize += n;
    proxy->call (REX_DATA, &arg, pres,
		 wrap (mkref (this), &unixfd::datacb, n, pres));
  }

  if (rsize >= hiwat)
    fdcb (localfd_in, selread, NULL);
}

void
unixfd::newfdcb (int fdrecved, ptr<rex_newfd_res> resp, clnt_stat)
{
  if (!resp->ok) {
    close (fdrecved);
    return;
  }
  vNew refcounted<unixfd> (pch, *resp->newfd, fdrecved);
}

void
unixfd::newfd (svccb *sbp)
{
  assert (paios_out);

  rexcb_newfd_arg *argp = sbp->Xtmpl getarg<rexcb_newfd_arg> ();
    
  int s[2];
    
  if(socketpair(AF_UNIX, SOCK_STREAM, 0, s)) {
    warn << "error creating socketpair";
    sbp->replyref (false);
    return;
  }
    
  make_async (s[1]);
  make_async (s[0]);
  close_on_exec (s[1]);
  close_on_exec (s[0]);
    
  paios_out->sendfd (s[1]);
    
  vNew refcounted<unixfd> (pch, argp->newfd, s[0]);
  sbp->replyref (true);
}

void
unixfd::data (svccb *sbp)
{
  assert (paios_out);
    
  rex_payload *argp = sbp->Xtmpl getarg<rex_payload> ();
    
  if (argp->data.size () > 0) {
    if (weof) {
      sbp->replyref (false);
      return;
    }
    else {
      str data (argp->data.base (), argp->data.size ());
      paios_out << data;
      sbp->replyref (true);
    }
  }
  else {
    sbp->replyref (true);
      
    //we don't shutdown immediately to give data a chance to
    //asynchronously flush
    int flag = SHUT_WR;
    paios_out->setwcb (wrap (this, &unixfd::update_connstate, flag));
  }
}

/* unixfd specific arguments:

   localfd_in:	local file descriptor for input (or everything for non-tty)

   localfd_out: local file descriptor for output (only used for tty support)
     Set localfd_out = -1 (or omit the argument entirely) if you're not
     working with a remote tty; then the localfd_in is connected directly
     to the remote FD for both reads and writes (except it it's RO or WO).

   noclose: Unixfd will not use close or shutdown calls on the local
     file descriptor (localfd_in); useful for terminal descriptors,
     which must hang around so that raw mode can be disabled, etc.

   shutrdonexit: When the remote module exits, we shutdown the read
     direction of the local file descriptor (_in).  This isn't always
     done since not all file descriptors managed on the REX channel
     are necessarily connected to the remote module.
*/
unixfd::unixfd (rexchannel *pch, int fd, int localfd_in, int localfd_out,
		bool noclose, bool shutrdonexit, cbv closecb)
  : rexfd::rexfd (pch, fd),
    localfd_in (localfd_in), localfd_out (localfd_out), rsize (0),
    unixsock (isunixsocket (localfd_in)), weof (false), reof (false),
    shutrdonexit (shutrdonexit), closecb (closecb)
{
  if (noclose) {
    int duplocalfd = dup (localfd_in);
    if (duplocalfd < 0)
      warn ("failed to duplicate fd for noclose behavior (%m)\n");
    else
      unixfd::localfd_in = duplocalfd;
  }

  make_async (this->localfd_in);
  if (!is_fd_wronly (this->localfd_in))
    fdcb (this->localfd_in, selread, wrap (this, &unixfd::rcb));
    
  /* for tty support we split the input/output to two local FDs */
  if (localfd_out >= 0)
    paios_out = aios::alloc (this->localfd_out);
  else
    paios_out = aios::alloc (this->localfd_in);
}


void
rexchannel::remove_fd (int fdn)
{
  // warn ("--[%p] remove_fd (%d), fdc = %d, size = %d\n",
  //       this, fdn, fdc, vfds.size ());
  assert (fdn >= 0);
  assert ((unsigned) fdn < vfds.size ());
  vfds[fdn] = NULL;
  if (!--fdc)
    sess->remove_chan (channo); 
}

void
rexchannel::deref_vfds ()
{
  size_t lvfds = vfds.size ();
  for (size_t f = 0; f < lvfds; f++)
    vfds[f] = NULL;
}

void
rexchannel::abort () {
  size_t lvfds = vfds.size ();
  for (size_t f = 0; f < lvfds; f++)
    if (vfds[f])
      vfds[f]->abort ();
}

void
rexchannel::quit ()
{
  // warn ("--[%p] rexchannel::quit\n", this);
  rex_int_arg arg;
  arg.channel = channo;
  arg.val = 15;
  proxy->call (REX_KILL, &arg, &rexfd::garbage_bool, aclnt_cb_null);
}

void
rexchannel::channelinit (u_int32_t chnumber, ptr<aclnt> proxyaclnt, int error)
{
  proxy = proxyaclnt;
  channo = chnumber;
  madechannel (error);
}

void
rexchannel::data (svccb *sbp)
{
  assert (sbp->prog () == REXCB_PROG && sbp->proc () == REXCB_DATA);
  rex_payload *dp = sbp->Xtmpl getarg<rex_payload> ();
  assert (dp->channel == channo);
  if (dp->fd < 0 ||
      implicit_cast<size_t> (dp->fd) >= vfds.size () ||
      !vfds[dp->fd]) {
    warn ("payload fd %d out of range\ndata:%s\n", dp->fd,
	  dp->data.base ());
    sbp->replyref (false);
    return;
  }

  vfds[dp->fd]->data (sbp);
}

void
rexchannel::newfd (svccb *sbp)
{
  assert (sbp->prog () == REXCB_PROG && sbp->proc () == REXCB_NEWFD);
  rexcb_newfd_arg *arg = sbp->Xtmpl getarg<rexcb_newfd_arg> ();

  int fd = arg->fd;

  if (fd < 0 || implicit_cast<size_t> (fd) >= vfds.size () || !vfds[fd]) {
    warn ("newfd received on invalid fd %d at rexchannel::newfd\n", fd);
    sbp->replyref (false);
    return;
  }
      
  vfds[fd]->newfd (sbp);
}

void
rexchannel::exited (int status)
{
  // warn ("--[%p] rexchannel::exited (%d); fdc = %d, vfds.size = %d\n",
  //       this, status, fdc, vfds.size ());
  for (size_t ix = 0; ix < vfds.size();  ix++) {
    if (!vfds[ix]) continue;
    vfds[ix]->exited ();
  } 
}

void
rexchannel::insert_fd (int fdn, ptr<rexfd> rfd)
{
  assert (fdn >= 0);

  // warn ("--[%p] insert_fd (%d), fdc = %d, size = %d\n",
  //       this, fdn, fdc, vfds.size ());
  size_t oldsize = vfds.size ();
  size_t neededsize = fdn + 1;
    
  if (neededsize > oldsize) {
    vfds.setsize (neededsize);
    for (int ix = oldsize; implicit_cast <size_t> (ix) < neededsize; ix++)
      vfds[ix] = NULL;
  }
    
  if (vfds[fdn]) {
    warn ("creating fd on busy fd %d at rexfd::rexfd, overwriting\n", fdn);
    assert (false);
  }
    
  vfds[fdn] = rfd;
  fdc++;
}


/* rexsession */

rexsession::rexsession (str schostname, ptr<axprt_crypt> proxyxprt,
                        vec<char> &secretid,
			callback<bool>::ptr failcb,
                        callback<bool>::ptr timeoutcb,
			bool verbose, bool resumable_init)
  : verbose (verbose), proxyxprt (proxyxprt), secretid (secretid),
    resumable (false), suspended (false),
    cchan (0), channelspending (0),
    endcb (NULL), failcb (failcb), timeoutcb (timeoutcb),
    schost (schostname), ifchg (NULL), silence_tmo_enabled (false),
    rexserv (asrv_resumable::alloc (proxyxprt, rexcb_prog_1,
                                    wrap (this, &rexsession::rexcb_dispatch))),
    proxy (aclnt_resumable::alloc (proxyxprt, rex_prog_1,
                                   wrap (this, &rexsession::fail)))
{
  silence_tmo_init ();
  setresumable (resumable_init);
}

rexsession::~rexsession ()
{
  ifchg_cb_clear ();
  silence_tmo_disable ();

  set_call_hook (NULL);
  set_recv_hook (NULL);
}

void
rexsession::ifchg_cb_set ()
{
  if (!ifchg)
    ifchg = ifchgcb (wrap (implicit_cast<axprt_stream *> (proxyxprt.get ()),
                           &axprt_stream::sockcheck));
}

void
rexsession::ifchg_cb_clear ()
{
  if (ifchg) {
    ifchgcb_remove (ifchg);
    ifchg = NULL;
  }
}

#define SILENCE_TMO ((time_t) 60)
#define PROBE_TMO ((time_t) 15)

void
rexsession::silence_tmo_init ()
{
  set_call_hook (wrap (this, &rexsession::rpc_call_hook));
  set_recv_hook (wrap (this, &rexsession::rpc_recv_hook));

  silence_check_cb = NULL;
  probe_call = NULL;

  silence_tmo_reset ();
}

void
rexsession::silence_tmo_reset ()
{
  silence_tmo_min = timenow + SILENCE_TMO;
  last_heard = timenow;
}

void
rexsession::silence_tmo_enable ()
{
  assert (!silence_tmo_enabled);
  if (timeoutcb) {
    silence_tmo_enabled = true;
    silence_check ();
  }
}

void
rexsession::silence_tmo_disable ()
{
  silence_tmo_enabled = false;
  if (silence_check_cb) {
    timecb_remove (silence_check_cb);
    silence_check_cb = NULL;
  }
  if (probe_call) {
    probe_call->cancel ();
    probe_call = NULL;
  }
}

void
rexsession::silence_check ()
{
  silence_check_cb = NULL;

  if (!proxy->calls_outstanding ())
    return;

  time_t tmo_time = max<time_t> (silence_tmo_min, last_heard + SILENCE_TMO);
  if (timenow >= tmo_time) {
    silence_tmo_disable ();
    probe_call = ping (wrap (this, &rexsession::probed), PROBE_TMO);
  }
  else
    silence_check_cb = timecb (tmo_time,
                               wrap (this, &rexsession::silence_check));
}

void
rexsession::probed (clnt_stat err)
{
  probe_call = NULL;

  if (err == RPC_TIMEDOUT)
    (*timeoutcb) ();
  else if (err)
    fail ();
}

inline void
rexsession::rpc_call_hook ()
{
  if (!proxy->calls_outstanding ())
    silence_tmo_min = timenow + SILENCE_TMO;

  if (!silence_check_cb && silence_tmo_enabled)
    silence_check ();
}

inline void
rexsession::rpc_recv_hook ()
{
  last_heard = timenow;
}

callbase *
rexsession::ping (callback<void, clnt_stat>::ref cb, time_t timeout)
{
  if (timeout)
    return proxy->timedcall (timeout, REX_NULL, NULL, NULL, cb);
  else
    return proxy->call (REX_NULL, NULL, NULL, cb);
}

bool
rexsession::fail ()
{
  if (!suspended) {
    suspend ();
    if (failcb)
      return (*failcb) ();  // might call resume
    else if (endcb) {
      (*endcb) ();
      return false;
    }
  }
  return true;
}

void
rexsession::suspend ()
{
  if (!suspended) {
    suspended = true;
    proxy->stop ();
    rexserv->stop ();
    if (resumable) {
      ifchg_cb_clear ();
      silence_tmo_disable ();
    }
  }
}

void
rexsession::setresumable (bool mode)
{
  if (resumable != mode) {
    resumable = mode;
    if (resumable) {
      ifchg_cb_set ();
      silence_tmo_enable ();
    }
    else {
      ifchg_cb_clear ();
      silence_tmo_disable ();
    }
    rex_setresumable_arg arg (resumable);
    if (resumable)
      arg.secretid->set (secretid.base (), secretid.size ());
    proxy->call (REX_SETRESUMABLE, &arg, NULL, aclnt_cb_null);
  }
}

void
rexsession::resumed (ptr<axprt_crypt> xprt, ref<bool> resp,
                     ptr<aclnt> proxytmp, callback<void, bool>::ref cb,
                     clnt_stat err)
{
  proxytmp = NULL;
  if (err) {
    warn << "REX_RESUME RPC failed (" << err << ")\n";
    cb (false);
    return;
  }
  if (!*resp) {
    warn << "proxy couldn't resume rex session\n";
    cb (false);
    return;
  }
  proxyxprt = xprt;

  ifchg_cb_set ();
  silence_tmo_reset ();
  silence_tmo_enable ();
  proxy->post_resume ();

  suspended = false;
  cb (true);
}

callbase *
rexsession::resume (ptr<axprt_crypt> xprt, sfs_seqno seqno,
                    callback<void, bool>::ref cb)
{
  if (!resumable) {
    cb (false);
    return NULL;
  }

  suspend ();

  if (!(rexserv->resume (xprt))) {
    cb (false);
    return NULL;
  }
  if (!(proxy->pre_resume (xprt))) {
    cb (false);
    return NULL;
  }

  ref<aclnt> proxytmp = aclnt::alloc (xprt, rex_prog_1);
  ref<bool> resp = New refcounted<bool> ();
  rex_resume_arg arg;
  arg.seqno = seqno;
  arg.secretid.set (secretid.base (), secretid.size ());
  return proxytmp->timedcall (30, REX_RESUME, &arg, resp,
                                  wrap (this, &rexsession::resumed, xprt, resp,
                                        proxytmp, cb));
}

void
rexsession::rexcb_dispatch (svccb *sbp)
{
  if (!sbp) {
    fail ();
    return;
  }
      
  switch (sbp->proc ()) {
	
  case REXCB_NULL:
    sbp->reply (NULL);
    break;
	
  case REXCB_EXIT:
    {
      rex_int_arg *argp = sbp->Xtmpl getarg<rex_int_arg> ();
      rexchannel *chan = channels[argp->channel];
	  
      if (chan) {
	chan->got_exit_cb = true;
	chan->exited (argp->val);
      }
      else {	// the channel was already shutdown "by EOF's"
	chan = channels_pending_exit[argp->channel];
	assert (chan);
	chan->got_exit_cb = true;
	chan->exited (argp->val);
	remove_chan_pending_exit (argp->channel);
      }

      sbp->reply (NULL);
      break;
    }
	
  case REXCB_DATA:
    {
      rex_payload *argp = sbp->Xtmpl getarg<rex_payload> ();
      rexchannel *chan = channels[argp->channel];

      if (chan)
	chan->data (sbp);
      else			    
	sbp->replyref (false);

      break;
    }
	
  case REXCB_NEWFD:
    {
      rex_int_arg *argp = sbp->Xtmpl getarg<rex_int_arg> ();
      rexchannel *chan = channels[argp->channel];

      if (chan)
	chan->newfd (sbp);
      else
	sbp->replyref (false);

      break;
    }
    
  default:
    sbp->reject (PROC_UNAVAIL);
    break;
  }
}
  
void
rexsession::madechannel (ptr<rex_mkchannel_res> resp, 
			 ptr<rexchannel> newchan, clnt_stat err)
{
  assert (channelspending);
  channelspending--;

  if (err) {
    warn << "REX_MKCHANNEL RPC failed (" << err << ")\n";
    newchan->channelinit (0, proxy, 1);
  }
  else if (resp->err != SFS_OK) {
    warn << "REX_MKCHANNEL failed (" << int (resp->err) << ")\n";
    newchan->channelinit (0, proxy, 1);
  }
  else {
    cchan++;

    if (verbose) {
      warn << "made channel: ";
      vec<str> command = newchan->get_cmd ();
      for (size_t i = 0; i < command.size (); i++)
	warnx << command[i] << " ";
      warnx << "\n";
    }
    channels.insert (resp->resok->channel, newchan);
    newchan->channelinit (resp->resok->channel, proxy, 0);
  }

  if (!cchan && !channelspending)
    if (endcb)
      endcb ();
}
 
void
rexsession::makechannel (ptr<rexchannel> newchan, rex_env env)
{
  rex_mkchannel_arg arg;

  arg.av.setsize (newchan->get_cmd ().size ());
  for (size_t i = 0; i < arg.av.size (); i++)
    arg.av[i] = newchan->get_cmd ()[i];
  arg.nfds = newchan->get_initnfds ();
  arg.env = env;
    
  channelspending++;
  ref<rex_mkchannel_res> resp (New refcounted<rex_mkchannel_res> ());
  proxy->call (REX_MKCHANNEL, &arg, resp, wrap (this,
						&rexsession::madechannel,
						resp, newchan));
}

void
rexsession::remove_chan (int channo)
{
  // warn << "--reached remove_chan; cchan = " << cchan << "\n";
  if (channels[channo] && !channels[channo]->got_exit_cb) {
    // warn ("--remove_chan [%p]: removing chan but "
    //       "haven't seen REXCB_EXIT yet\n", channels[channo].get ());
    channels_pending_exit.insert (channo, channels[channo]);
    channels.remove (channo);
    return;
  }
  // warn ("--remove_chan [%p]\n", channels[channo].get ());
  channels.remove (channo);
  if (!--cchan && !channelspending) {
    // warn << "--remove_chan: removing last channel\n";
    if (endcb)
      endcb ();
  }
}

void
rexsession::remove_chan_pending_exit (int channo)
{
  // warn << "--reached remove_chan_pending_exit; cchan = " << cchan << "\n";
  channels_pending_exit.remove (channo);
  if (!--cchan && !channelspending) {
    // warn << "--remove_chan_pending_exit: removing last channel\n";
    if (endcb)
      endcb ();
  }
}

