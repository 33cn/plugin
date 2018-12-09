/* $Id: auth_helper_pam.c 2091 2006-07-17 19:29:32Z max $ */

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
#include "suio.h"

#ifdef HAVE_PAM_PAM_APPL_H
#include <pam/pam_appl.h>
#else /* !HAVE_PAM_PAM_APPL_H */
#include <security/pam_appl.h>
#endif /* !HAVE_PAM_PAM_APPL_H */

static suio *chatter;
static int conversed;

static int
rpc_conv (int num_msg, const struct pam_message **msg,
	  struct pam_response **resp, void *appdata_ptr)
{
  struct pam_response *res = malloc (num_msg * sizeof (*res));
  int i;

  if (!res) {
    const char msg[] = "malloc failed\n";
    write (2, msg, sizeof (msg) - 1);
    abort ();
  }
#if !defined (__linux__) && !defined (__FreeBSD__)
  if (num_msg > 1) {
    fprintf (stderr, "%s(%s,%s): PAM conversation w. more than one message\n"
	     "Conservatively refusing to authenticate because "
	     "incompatibilities\n"
	     "in different OS APIs could lead to memory corruption\n",
	     progname, user, service);
    exit (1);
  }
#endif /* not linux or FreeBSD */

  bzero (res, num_msg * sizeof (*res));
  for (i = 0; i < num_msg; i++)
    switch (msg[i]->msg_style) {
    case PAM_PROMPT_ECHO_OFF:
    case PAM_PROMPT_ECHO_ON:
      {
	char *text;
	conversed = 1;
	suio_copy (chatter, msg[i]->msg, strlen (msg[i]->msg));
	suio_copy (chatter, "", 1);
	text = suio_flatten (chatter);
	suio_clear (chatter);
	res[i].resp = authhelp_input (text,
				      msg[i]->msg_style == PAM_PROMPT_ECHO_ON);
	xfree (text);
	break;
      }
    case PAM_ERROR_MSG:
      fprintf (stderr, "%s(%s,%s): %s\n", progname, service, user,
	       msg[i]->msg);
      break;
    case PAM_TEXT_INFO:
      suio_copy (chatter, msg[i]->msg, strlen (msg[i]->msg));
      suio_copy (chatter, "\n", 1);
      break;
    default:
      fprintf (stderr, "%s(%s,%s): unsupported PAM message type %d\n",
	       progname, user, service, msg[i]->msg_style);
      exit (1);
      break;
    }

  *resp = res;
  return PAM_SUCCESS;
}

int
authhelp_go (void)
{
  pam_handle_t *pamh = NULL;
  struct pam_conv conv;
  int err;
  char *u = NULL;

  chatter = suio_alloc ();
  bzero (&conv, sizeof (conv));
  conv.conv = rpc_conv;
  err = pam_start (service, user, &conv, &pamh);
  if (!user)
    user = "unknown user";
  if (err == PAM_SUCCESS)
    err = pam_authenticate (pamh, 0);
  if (err == PAM_SUCCESS)
    err = pam_acct_mgmt (pamh, 0);
  if (err == PAM_SUCCESS) {
    const void *vv;
    pam_get_item (pamh, PAM_USER, &vv);
    u = (char *) vv;
  }

  if (err == PAM_SUCCESS) {
    char *text;
    suio_copy (chatter, "", 1);
    text = suio_flatten (chatter);
    suio_clear (chatter);

    authhelp_approve (u, text);

    xfree (text);
  }
  else if (!conversed) {
    fprintf (stderr, "%s PAM(%s): %s\n", progname,
	     service, pam_strerror (pamh, err));
    if (!chdir ("/etc/pam.d") && access (service, F_OK))
      fprintf (stderr, "%s PAM(%s): %s\n", progname, service,
	       "Try running: ln -s system-auth /etc/pam.d/sfs");
    else
      fprintf (stderr, "%s PAM(%s): make PAM config allows service %s"
	       " or other\n", progname, service, service);
  }

#if 0
  if (err == PAM_SUCCESS)
    fprintf (stderr, "%s(%s,%s): authenticated %s\n", progname,
	     user, service, u);
  else
    fprintf (stderr, "%s(%s,%s): %s\n", progname,
	     user, service, pam_strerror (pamh, err));
#endif

  pam_end (pamh, err);
  return err == PAM_SUCCESS ? 0 : -1;
}
