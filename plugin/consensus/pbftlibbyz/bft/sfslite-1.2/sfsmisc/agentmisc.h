// -*-c++-*-
/* $Id: agentmisc.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
 * Copyright (C) 1999 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#ifndef _SFSMISC_AGENTMISC_H_
#define _SFSMISC_AGENTMISC_H_ 1

#include "sfsagent.h"

extern str dotsfs; 
extern str agentsock;
extern str userkeysdir;

str agent_userdir (u_int32_t uid, bool create);
str agent_usersock (bool create_dir = false);
void agent_setsock ();
void agent_ckdir (bool fail_on_keysdir = true);
bool agent_ckdir (const str &dir);
void agent_mkdir ();
void agent_mkdir (const str &dir);
str defkey ();
void rndaskcd ();
void agent_spawn (bool opt_verbose = false);
bool isagentrunning ();

#endif /* !_SFSMISC_AGENTMISC_H_ */
