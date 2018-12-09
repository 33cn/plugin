/* $Id */

/*
 *
 * Copyright (C) 2004 Maxwell Krohn (max@okcupid.com)
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


#include <sys/time.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdio.h>
#include <errno.h>
#include "sysconf.h"

#define CLCKD_INTERVAL 10000
#define THOUSAND 1000
#define MILLION THOUSAND*THOUSAND

const char *progname;

static void
usage (int argc, char *argv[])
{
  fprintf (stderr, "usage: %s clockfile\n", argv[0]);
  exit (1);
}

static int 
timespec_diff (struct timespec a, struct timespec b)
{
  return  (a.tv_nsec - b.tv_nsec) / THOUSAND -
    (b.tv_sec - a.tv_sec) * MILLION;
}

static void 
mmcd_shutdown (void *rgn, size_t sz, int fd, char *fn, int sig)
{
  fprintf (stderr, "%s: exiting on signal %d\n", progname, sig);
  munmap (rgn, sz);
  close (fd);

  /*
   *  XXX - this might segfault other processes ; not sure
   */ 
  unlink (fn);
}

void
setprogname (const char *in)
{
  const char *slashp = NULL;
  const char *cp;

  for (cp = in; *cp; cp++) 
    if (*cp == '/') 
      slashp = cp;

  progname = slashp ? slashp + 1 : in;
}

#define BUFSZ 1024

static int
mmcd_gettime (struct timespec *tp)
{
  struct timeval tv;
  if (gettimeofday (&tv, NULL) < 0)
      return -1;
  tp->tv_sec = tv.tv_sec;
  tp->tv_nsec = tv.tv_usec * 1000;
  return 0;
}

/*
 * global variables that can be accessed from within signal handlers!
 */
char *mmap_rgn;
size_t mmap_rgn_sz;
int mmap_fd;
char *mmap_file;

static void
handle_sigint (int i)
{
  if (i == SIGTERM || i == SIGINT || i == SIGKILL) {
    mmcd_shutdown (mmap_rgn, mmap_rgn_sz, mmap_fd, mmap_file, i);
    exit (0);
  } else {
    fprintf (stderr, "unexpected signal caught: %d; ignoring it\n", i);
  }
}

static void
set_signal_handlers ()
{
  struct sigaction sa;
  memset (&sa, 0, sizeof (sa));
  sa.sa_handler = handle_sigint;
  sa.sa_flags = 0;
  if (sigaction (SIGTERM, &sa, NULL) < 0) {
    fprintf (stderr, "bad sigaction; errno=%d\n", errno);
    exit (-1);
  }
  if (sigaction (SIGINT, &sa, NULL) < 0) {
    fprintf (stderr, "bad sigaction; errno=%d\n", errno);
    exit (-1);
  }
}
 
int
main (int argc, char *argv[])
{
  struct timespec ts[2];
  struct timespec *targ;
  struct timeval wt;
  int d;
  int rc;

  setprogname (argv[0]);

  /* 
   * make the file the right size by writing our time
   * there
   */
  mmcd_gettime (ts);
  ts[1]  = ts[0];

  if (argc != 2) 
    usage (argc, argv);

  mmap_file = argv[1];
  mmap_fd = open (mmap_file, O_RDWR|O_CREAT, 0644);

  if (mmap_fd < 0) {
    fprintf (stderr, "mmcd: %s: cannot open file for reading\n", argv[1]);
    exit (1);
  }
  if (write (mmap_fd, (char *)ts, sizeof (ts)) != sizeof (ts)) {
    fprintf (stderr, "mmcd: %s: short write\n", argv[1]);
    exit (1);
  }

  mmap_rgn_sz = sizeof (struct timespec )  * 2;
  mmap_rgn = mmap (NULL, mmap_rgn_sz, PROT_READ|PROT_WRITE, MAP_SHARED, mmap_fd, 0); 
  if (mmap_rgn == MAP_FAILED) {
    fprintf (stderr, "mmcd: mmap failed: %d\n", errno);
    exit (1);
  }

  fprintf (stderr, "%s: starting up; file=%s; pid=%d\n", 
	   progname, mmap_file, getpid ());

  set_signal_handlers ();
  
  targ = (struct timespec *) mmap_rgn;
  while (1) {
    
    wt.tv_sec = 0;
    wt.tv_usec = CLCKD_INTERVAL;

    mmcd_gettime (ts); 
    targ[0] = ts[0];
    targ[1] = ts[0];
      
    rc = select (0, NULL, NULL, NULL, &wt);
    if (rc < 0) {
      fprintf (stderr, "%s: select failed: %d\n", progname, errno);
      continue;
    } 

    mmcd_gettime (ts+1); 
    d = timespec_diff (ts[1], ts[0]);
    /*
     * long sleeps will hurt the accuracy of the clock; may as well
     * report them, although eventually it would be nice to do something
     * about them on the client side;
     *
     */
    if (d > 10 * CLCKD_INTERVAL)
      fprintf (stderr, "%s: %s: very long sleep: %d\n", progname, argv[0], d);
  }
  return 0;
}
