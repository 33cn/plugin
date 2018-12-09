/* $Id: test_timecb.C 2 2003-09-24 14:35:33Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
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


#include "async.h"
#include "crypt.h"

void phase2 ();

class timetest {
public:
  static int ctr;
  static u_int ttno;

private:
  int ok;
  int no;
  timecb_t *t[2];

  void timeout (int i) {
    assert (ok);
    assert (t[i]);
    t[i] = NULL;
    delete this;
  }

  PRIVDEST ~timetest () {
    ok = 0;
    for (int i = 0; i < 2; i++)
      if (t[i]) {
	timecb_remove (t[i]);
	t[i] = NULL;
      }
    ctr--;
    if (!ctr)
      phase2 ();
    if (ctr < 0)
      panic ("too many callbacks\n");
  }

public:
  timetest () : ok (1), no (++ttno) {
    ctr++;
    t[0] = t[1] = NULL;
    timecb_t *tt;
    tt = timecb (time (NULL) + rnd.getword () % 8,
		wrap (this, &timetest::timeout, 0));
    if (!tt)
      return;
    t[0] = tt;
    tt = timecb (time (NULL) + rnd.getword () % 8,
		 wrap (this, &timetest::timeout, 1));
    if (!tt)
      return;
    t[1] = tt;
  }
};

int timetest::ctr;
u_int timetest::ttno;

static timespec onesecond;

static void
phase2cb (timespec mintime, int numtogo)
{
  timespec ts;
  clock_gettime (CLOCK_REALTIME, &ts);
  if (mintime > ts)
    panic ("callback too early\n");
  if (!--numtogo)
    exit (0);
  ts.tv_sec++;
  delaycb (1, wrap (phase2cb, ts, numtogo));
}

void
phase2 ()
{
  phase2cb (onesecond, 3);
}

static void
timeout (int)
{
  char msg[] = "lost a timecb\n";
  write (2, msg, sizeof (msg) - 1);
  abort ();
}

int
main (int argc, char **argv)
{
  setprogname (argv[0]);

  timetest::ctr++;
  for (int i = 0; i < 256; i++)
    New timetest ();
  if (!--timetest::ctr)
    return 0;
  signal (SIGALRM, timeout);
  alarm (15);
  amain ();
}
