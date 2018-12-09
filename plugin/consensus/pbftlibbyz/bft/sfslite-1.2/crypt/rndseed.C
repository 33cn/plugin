/* $Id: rndseed.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "crypt.h"

prng rnd;
sha1oracle rnd_input (64, 0, 8);

enum { seedsize = 48 };
const int mapsize = sysconf (_SC_PAGESIZE);

static void *seed;
static bool initialized;
static u_int64_t nupdates;

EXITFN(saveseed);

static void
saveseed ()
{
  if (seed)
    random_update ();
}

void
random_update ()
{
  if (seed)
    rnd_input.update (seed, seedsize);
  getclocknoise (&rnd_input);
  rnd.seed_oracle (&rnd_input);
  if (seed)
    rnd.getbytes (seed, seedsize);
  nupdates++;
}

static void
random_stir ()
{
  getsysnoise (&rnd_input, wrap (random_update));
}

static void
random_timer ()
{
  random_stir ();
  timecb (time (NULL) + 1800 + rnd.getword () % 1800, wrap (random_timer));
}

static u_int32_t
random_word ()
{
  return rnd.getword ();
}

void
random_start ()
{
  if (!initialized) {
    initialized = true;
    random_update ();
    arandom_fn = random_word;
    random_timer ();
  }
}

void
random_init ()
{
  if (initialized)
    random_update ();
  else {
    random_start ();
    while (!nupdates)
      acheck ();
  }
}

void
random_set_seedfile (str path)
{
  if (!path) {
    if (seed) {
      munmap (reinterpret_cast<char *> (seed), mapsize);
      seed = NULL;
    }
    return;
  }

  if (path[0] == '~' && path[1] == '/') {
    const char *home = getenv ("HOME");
    if (!home) {
      warn ("$HOME not set in environment\n");
      return;
    }
    path = strbuf () << home << (path.cstr () + 1);
  }

  int fd = open (path, O_CREAT|O_RDWR, 0600);
  if (fd < 0) {
    warn ("%s: %m\n", path.cstr ());
    return;
  }
  struct stat sb;
  char c;
  if (read (fd, &c, 1) < 0 || fstat (fd, &sb) < 0
      || lseek (fd, mapsize - 1, SEEK_SET) == -1
      || write (fd, "", 1) < 0) {
    /* The read call avoids a segfault on NFS 2.  Specifically, if we
     * are root and the random_seed file is over NFS 2, the open will
     * succeed even though read returns EACCES.  If we map the file
     * when we can't read it--bingo, segfault.  (In fact, on some OSes
     * it also seems to cause a kernel panic when you examine the
     * mmapped memory from the debugger.) */
    close (fd);
    warn ("%s: %m\n", path.cstr ());
    return;
  }
  if ((sb.st_mode & 07777) != 0600)
    warn ("%s: mode 0%o should be 0600\n", path.cstr (), sb.st_mode & 07777);

  if (seed)
    munmap (reinterpret_cast<char *> (seed), mapsize);
  seed = mmap (NULL, (size_t) mapsize, PROT_READ|PROT_WRITE,
	       MAP_FILE|MAP_SHARED, fd, 0);
  if (seed == reinterpret_cast<void *> (MAP_FAILED)) {
    warn ("mmap: %s: %m\n", path.cstr ());
    seed = NULL;
  }
  else
    rnd_input.update (seed, seedsize);
  close (fd);
}

void
random_init_file (str path)
{
  random_set_seedfile (path);
  random_init ();
}

u_int32_t
random_getword ()
{
  if (!initialized)
    random_init ();
  return rnd.getword ();
}
