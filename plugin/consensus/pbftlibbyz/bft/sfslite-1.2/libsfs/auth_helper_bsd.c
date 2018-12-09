/* $Id: auth_helper_bsd.c 435 2004-06-02 15:46:36Z max $ */

/*
 *
 * Copyright (C) 2004 David Mazieres (dm@uun.org)
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

#include "auth_helper.h"
#include <login_cap.h>
#include <bsd_auth.h>

int
authhelp_go (void)
{
  auth_session_t *as;
  char *chal = NULL;
  char *pass;
  char *p;
  int res;

  as = auth_userchallenge (user, NULL, NULL, &chal);
  pass = authhelp_input (chal ? chal : "Password: ", 0);
  if (!as)
    return -1;
  res = auth_userresponse (as, pass, 0);
  xfree (pass);

  if ((p = strchr (user, ':')))
    *p = '\0';
  if (res)
    authhelp_approve (user, NULL);

  return res ? 0 : -1;
}
