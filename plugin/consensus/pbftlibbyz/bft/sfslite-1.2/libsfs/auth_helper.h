/* $Id: auth_helper.h 435 2004-06-02 15:46:36Z max $ */

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

#ifndef _LIBSFS_AUTH_HELPER_H_
#define _LIBSFS_AUTH_HELPER_H_ 1

#include "sfs-internal.h"
#include "auth_helper_prot.h"

extern char *progname;
extern char *service;
extern char *user;
char *authhelp_input (char *prompt, int echo);
void authhelp_approve (char *user, char *hello);
int authhelp_go (void);

#endif /* !_LIBSFS_AUTH_HELPER_H_ */
