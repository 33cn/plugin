/* $Id: prng.C 1117 2005-11-01 16:20:39Z max $ */

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

#include "sha1.h"
#include "prng.h"

const u_int32_t prng::initdat[5] = {
  0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476, 0xc3d2e1f0
};

prng::prng ()
  : inpos (input.bytes), inlim (inpos + sizeof (input.bytes))
{
  // For debugging, put in a deterministic state by default
  bzero (state.bytes, sizeof (state.bytes));
}

void
prng::transform (sumbuf<5> *output)
{
  output->set (initdat);
  if (inpos == input.bytes)
    sha1::transform (output->words, state.bytes);
  else {
    if (inpos != inlim)
      bzero (inpos, inlim - inpos);
    // XXX - template args required for egs bug
    sumbufadd<16, 16> (&input, &state);
    sha1::transform (output->words, input.bytes);
    inpos = input.bytes;
  }
  // XXX - template args required for egs bug
  sumbufadd<16, 5> (&state, output, true);
}

void
prng::seed (const u_char buf[64])
{
  state.set (buf);
}

void
prng::seed_oracle (sha1oracle *ora)
{
  const size_t bufsize = max<size_t> (ora->resultsize, 64);
  u_char *buf = New u_char[bufsize];

  bzero (buf, 64);
  getbytes (buf, bufsize);
  ora->update (buf, bufsize);

  ora->final (buf);
  seed (buf);

  ora->reset ();
  bzero (buf, bufsize);
  delete[] buf;
}

void
prng::getbytes (void *buf, size_t len)
{
  char *cp = static_cast<char *> (buf);
  sumbuf<5> out;

  // getclocknoise (this);
  while (len >= sizeof (out)) {
    transform (&out);
    memcpy (cp, out.chars, sizeof (out.chars));
    cp += sizeof (out);
    len -= sizeof (out);
  }
  if (len > 0) {
    transform (&out);
    memcpy (cp, out.chars, len);
  }
}

void
prng::update (const void *buf, size_t len)
{
  sumbuf<5> junk;
  const char *cp = static_cast<const char *> (buf);
  const char *end = cp + len;

  while (cp < end) {
    if (inpos == inlim)
      transform (&junk);
    size_t n = min (end - cp, inlim - inpos);
    memcpy (inpos, cp, n);
    cp += n;
    inpos += n;
  }
}
