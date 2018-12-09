/* $Id: agentconn.h 1754 2006-05-19 20:59:19Z max $ */

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

#ifndef _SFSMISC_AGENTCONN_H_
#define _SFSMISC_AGENTCONN_H_ 1

#include "async.h"
#include "arpc.h"
#include "sfsmisc.h"
#include "sfsagent.h"
#include "agentmisc.h"
#include "sfssesscrypt.h"

class agentconn : public virtual refcount, public virtual sfs_authorizer {
private:
  bool ccddone;
  int agentfd;
  ptr<aclnt> agentclnt_ctl;
  ptr<aclnt> agentclnt_cb;
  ptr<axprt_unix> sfscdxprt;
  ptr<aclnt> sfscdclnt;

  static void authcb (sfsagent_auth_res *resp, cbv cb, clnt_stat stat);

public:
  agentconn ()
    : ccddone (false), agentfd (-1) {}
  ~agentconn () {}

  int cagent_fd (bool required = true);
  ptr<aclnt> ccd (bool required = true);
  ref<aclnt> cagent_ctl ();
  ref<aclnt> cagent_cb ();

  str lookup (str &hostname);
  ptr<sfsagent_rex_res> rex (str dest, str schost, bool forwardagent,
                             bool agentconnect, bool resumable);
  ptr<bool> keepalive (str schost);
  bool isagentrunning ();

  void authinit (const sfsagent_authinit_arg *argp,
			 sfsagent_auth_res *resp, cbv cb);
  void authmore (const sfsagent_authmore_arg *argp,
		 sfsagent_auth_res *resp, cbv cb);
};

#endif /* _SFSMISC_AGENTCONN_H_ */
