/* $Id: sfstty.C 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2000-2001 Eric Peterson (ericp@lcs.mit.edu)
 * Copyright (C) 2001 Michael Kaminsky (kaminsky@lcs.mit.edu)
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

#ifndef TIOCGWINSZ
#include <sys/ioctl.h>
#endif

#include "async.h"
#include "sfstty.h"

/* Terminal modes, as saved by enter_raw_mode. */
static struct termios saved_tio;
static bool in_raw_mode = false;

#if 0
EXITFN (leave_raw_mode);
#endif

str
windowsizetostring (struct winsize *size)
{
  if (ioctl (STDIN_FILENO, TIOCGWINSZ, (char *) size) < 0)
    fatal ("TIOCGWINSZ error\n");

  return strbuf() << size->ws_row << "\n" << size->ws_col <<
    "\n" << size->ws_xpixel << "\n" << size->ws_ypixel << "\n";
}

/* 
 * The following code (2 functions) is from the OpenSSH source.
 */

void
leave_raw_mode()
{
  if (!in_raw_mode)
    return;

  if (tcsetattr (fileno (stdin), TCSADRAIN, &saved_tio) < 0)
    warn << "leave_raw_mode: tcsetattr: " << strerror (errno) << "\n";
  else
    in_raw_mode = false;
}

void
enter_raw_mode()
{   
  struct termios tio;   

  if (in_raw_mode)
    return;

  if (tcgetattr (fileno (stdin), &tio) < 0)
    warn << "enter_raw_mode: tcgetattr: " << strerror (errno) << "\n";

  saved_tio = tio;
  tio.c_iflag |= IGNPAR;
  tio.c_iflag &= ~(ISTRIP | INLCR | IGNCR | ICRNL | IXON | IXANY | IXOFF);
  tio.c_lflag &= ~(ISIG | ICANON | ECHO | ECHOE | ECHOK | ECHONL);
#ifdef IEXTEN
  tio.c_lflag &= ~IEXTEN;
#endif /* IEXTEN */
  tio.c_oflag &= ~OPOST;
  tio.c_cc[VMIN] = 1;
  tio.c_cc[VTIME] = 0;

  if (tcsetattr (fileno (stdin), TCSADRAIN, &tio) < 0)
    warn << "enter_raw_mode: tcsetattr: " << strerror (errno) << "\n";
  else
    in_raw_mode = true;
}

