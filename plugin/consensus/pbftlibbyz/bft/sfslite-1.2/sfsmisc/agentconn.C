/* $Id: agentconn.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2000 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#include "agentconn.h"
#include "rxx.h"
#include "rexcommon.h"

str
agentconn::lookup (str &file)
{
  str ret;

  str host = "localhost";
  int err;
  if (clnt_stat stat = cagent_ctl ()->scall (AGENTCTL_FORWARD, &host, &err)) {
    warn << "RPC error: " << stat << "\n";
    return ret;
  }
  if (err) {
    warn << "agent (ctl): " << strerror (err) << "\n";
    return ret;
  }

  sfsagent_lookup_res res;
  if (clnt_stat stat = cagent_cb ()->scall (AGENTCB_LOOKUP, &file, &res)) {
    warn << "agent (cb): " << stat << "\n";
    return ret;
  }
  if (res.type == LOOKUP_MAKELINK)
    return *res.path;
  return ret;
}

void
agentconn::authcb (sfsagent_auth_res *resp, cbv cb, clnt_stat stat)
{
  if (stat) {
    warn << "agent: " << stat << "\n";
    resp->set_authenticate (false);
  }
  (*cb) ();
}

void
agentconn::authinit (const sfsagent_authinit_arg *argp,
		     sfsagent_auth_res *resp, cbv cb)
{
  cagent_cb()->call (AGENTCB_AUTHINIT, argp, resp,
		     wrap (authcb, resp, cb));
}

void
agentconn::authmore (const sfsagent_authmore_arg *argp,
		     sfsagent_auth_res *resp, cbv cb)
{
  cagent_cb()->call (AGENTCB_AUTHMORE, argp, resp,
		     wrap (authcb, resp, cb));
}

#if 0
ptr<sfsagent_auth_res>
agentconn::auth (sfsagent_authinit_arg &arg)
{
  str host = "localhost";
  int err;
  ref<sfsagent_auth_res> res = New refcounted<sfsagent_auth_res> (FALSE);
  sfsagent_auth_res r;

  if (clnt_stat stat = cagent_ctl ()->scall (AGENTCTL_FORWARD, &host, &err)) {
    warn << "RPC error: " << stat << "\n";
    return res;
  }
  if (err) {
    warn << "agent (ctl): " << strerror (err) << "\n";
    return res;
  }

  if (clnt_stat stat = cagent_cb ()->scall (AGENTCB_AUTHINIT, &arg, &r)) {
    warn << "agent (cb): " << stat << "\n";
    return res;
  }
  else
    *res = r;
  return res;
}
#endif

ptr<sfsagent_rex_res>
agentconn::rex (str dest, str schost,
                bool forwardagent, bool blockactive, bool resumable)
{
  ref<sfsagent_rex_res> res = New refcounted<sfsagent_rex_res_w> (FALSE);
  sfsagent_rex_arg a;
  a.dest = dest;
  a.schost = schost;
  a.forwardagent = forwardagent;
  a.blockactive = blockactive;
  a.resumable = resumable;
  sfsagent_rex_res r;

  if (clnt_stat stat = cagent_ctl ()->scall (AGENTCTL_REX, &a, &r)) {
    warn << "agent (cb): " << stat << "\n";
    return NULL;
  }
  else
    *res = r;
  return res;
}

ptr<bool>
agentconn::keepalive (str schost)
{
  ref<bool> res = New refcounted<bool> (false);
  bool r;

  if (clnt_stat stat = cagent_ctl ()->scall (AGENTCTL_KEEPALIVE,
                                             &schost, &r)) {
    warn << "agent (cb): " << stat << "\n";
    return NULL;
  }
  else {
    *res = r;
    return res;
  }
}

bool
agentconn::isagentrunning ()
{
  return cagent_fd (false) >= 0;
}

ptr<aclnt>
agentconn::ccd (bool required)
{
  if (sfscdclnt || (ccddone && !required))
    return sfscdclnt;
  ccddone = true;
  int fd = required ? suidgetfd_required ("agent") : suidgetfd ("agent");
  if (fd >= 0) {
    sfscdxprt = axprt_unix::alloc (fd);
    sfscdclnt = aclnt::alloc (sfscdxprt, agent_prog_1);
  }
  return sfscdclnt;
}

int
agentconn::cagent_fd (bool required)
{
  if (agentfd >= 0)
    return agentfd;

  static rxx sockfdre ("^-(\\d+)?$");
  if (agentsock && sockfdre.search (agentsock)) {
    if (sockfdre[1])
      agentfd = atoi (sockfdre[1]);
    else
      agentfd = 0;
    if (!isunixsocket (agentfd))
      fatal << "fd specified with '-S' not unix domain socket\n";
  }
  else if (agentsock) {
    agentfd = unixsocket_connect (agentsock);
    if (agentfd < 0 && required)
      fatal ("%s: %m\n", agentsock.cstr ());
  }
  else if (ccd (false)) {
    int32_t res;
    if (clnt_stat err = ccd ()->scall (AGENT_GETAGENT, NULL, &res)) {
      if (required)
	fatal << "sfscd: " << err << "\n";
    }
    else if (res) {
      if (required)
	fatal << "connecting to agent via sfscd: " << strerror (res) << "\n";
    }
    else if ((agentfd = sfscdxprt->recvfd ()) < 0) {
      fatal << "connecting to agent via sfscd: "
	    << "could not get file descriptor\n";
    }
  }
  else {
    if (str sock = agent_usersock (true))
      agentfd = unixsocket_connect (sock);
    if (agentfd < 0 && required)
      fatal << "sfscd not running and no standalone agent socket\n";
  }
  return agentfd;
}

ref<aclnt>
agentconn::cagent_ctl ()
{
  if (agentclnt_ctl)
    return agentclnt_ctl;

  int fd = cagent_fd ();

  agentclnt_ctl = aclnt::alloc (axprt_stream::alloc (fd), agentctl_prog_1);
  return agentclnt_ctl;
}

ref<aclnt>
agentconn::cagent_cb ()
{
  if (agentclnt_cb)
    return agentclnt_cb;

  int fd = cagent_fd ();

  // str name (progname ? progname : str ("LOCAL"));
  str name ("LOCAL");		// XXX
  int32_t err;
  if (clnt_stat stat = cagent_ctl ()->scall (AGENTCTL_FORWARD, &name, &err))
    fatal << "agent: " << stat << "\n";
  if (err)
    fatal << "agent forwarding: " << strerror (err) << "\n";

  agentclnt_cb = aclnt::alloc (axprt_stream::alloc (fd), agentcb_prog_1);
  return agentclnt_cb;
}
