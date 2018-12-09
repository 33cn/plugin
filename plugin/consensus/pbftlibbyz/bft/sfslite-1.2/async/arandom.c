/* $Id: arandom.c 1117 2005-11-01 16:20:39Z max $ */

/*
 *
 * Copyright (C) 2003 David Mazieres (dm@uun.org)
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

#include "sysconf.h"

/* XXX - need initializer because of broken Apple Darwin linker */
u_int32_t (*arandom_fn) () = 0;

#ifndef HAVE_ARC4RANDOM
/* This is a simple random number generator, based on the ARC4 stream
 * cipher.  THIS IS NOT CRYPTOGRAPHICALLY SECURE!  It is simply for
 * getting a somewhat random-looking series of bytes.  In order to get
 * cryptographically secure random numbers, you have to set the
 * arandom_fn function pointer to a real cryptographic pseudo-random
 * generator
 */

struct arc4rnd {
  u_char i;
  u_char j;
  u_char s[256];

};
typedef struct arc4rnd arc4rnd;

static inline u_int32_t
arc4rnd_getbyte (arc4rnd *a)
{
  u_char si, sj;
  a->i = (a->i + 1) & 0xff;
  si = a->s[a->i];
  a->j = (a->j + si) & 0xff;
  sj = a->s[a->j];
  a->s[a->i] = sj;
  a->s[a->j] = si;
  return a->s[(si + sj) & 0xff];
}

static void
arc4rnd_reset (arc4rnd *a)
{
  int n;
  a->i = 0xff;
  a->j = 0;
  for (n = 0; n < 0x100; n++)
    a->s[n] = n;
}

static void
_arc4rnd_setkey (arc4rnd *a, const u_char *key, size_t keylen)
{
  u_int n, keypos;
  u_char si;
  for (n = 0, keypos = 0; n < 256; n++, keypos++) {
    if (keypos >= keylen)
      keypos = 0;
    a->i = (a->i + 1) & 0xff;
    si = a->s[a->i];
    a->j = (a->j + si + key[keypos]) & 0xff;
    a->s[a->i] = a->s[a->j];
    a->s[a->j] = si;
  }
}

static void
arc4rnd_setkey (arc4rnd *a, const void *_key, size_t len)
{
  const u_char *key = (const u_char *) _key;
  arc4rnd_reset (a);
  while (len > 128) {
    len -= 128;
    key += 128;
    _arc4rnd_setkey (a, key, 128);
  }
  if (len > 0)
    _arc4rnd_setkey (a, key, len);
  a->j = a->i;
}

static arc4rnd bad_random_state;
struct bad_random_junk {
  struct timespec now;
  pid_t pid;
};
static void
bad_random_init ()
{
  int i;
  struct bad_random_junk brj;
#ifdef SFS_DEV_RANDOM
  int fd = open (SFS_DEV_RANDOM, O_RDONLY);
  if (fd >= 0) {
    char buf[256];
    int n;
    if ((n = fcntl (fd, F_GETFL)) != -1
	&& fcntl (fd, F_SETFL, n | O_NONBLOCK) != -1
	&& (n = read (fd, buf, sizeof (buf))) >= 12) {
      close (fd);
      arc4rnd_setkey (&bad_random_state, buf, n);
      return;
    }
    close (fd);
  }
#endif /* SFS_DEV_RANDOM */
  clock_gettime (CLOCK_REALTIME, &brj.now);
  brj.pid = getpid ();
  arc4rnd_setkey (&bad_random_state, &brj, sizeof (brj));
  for (i = 0; i < 256; i++)
    arc4rnd_getbyte (&bad_random_state);
}
static u_int32_t
bad_random ()
{
  return arc4rnd_getbyte (&bad_random_state)
    | arc4rnd_getbyte (&bad_random_state) << 8
    | arc4rnd_getbyte (&bad_random_state) << 16
    | arc4rnd_getbyte (&bad_random_state) << 24;
}

#endif /* !HAVE_ARC4RANDOM */

u_int32_t
arandom ()
{
  if (!arandom_fn) {
#ifdef HAVE_ARC4RANDOM
    arandom_fn = arc4random;
#else /* !HAVE_ARC4RANDOM */
    bad_random_init ();
    arandom_fn = bad_random;
#endif /* !HAVE_ARC4RANDOM */
  }
  return (*arandom_fn) ();
}
