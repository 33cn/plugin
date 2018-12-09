/* $Id: mdblock.C 1117 2005-11-01 16:20:39Z max $ */

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


#include "crypthash.h"

void
mdblock::update (const void *_dp, size_t len)
{
  const u_char *data = static_cast<const u_char *> (_dp);
  size_t i;
  u_int bcount = count % blocksize;

  count += len;
  if (bcount + len < blocksize) {
    memcpy (&buffer[bcount], data, len);
    return;
  }
  /* copy first chunk into context, do the rest (if any) */
  /* directly from data array */
  if (bcount) {
    int j = blocksize - bcount;
    memcpy (&buffer[bcount], data, j);
    consume (buffer);
    i = j;
    len -= j;
  }
  else
    i = 0;

  while (len >= blocksize) {
    consume (&data[i]);
    i += blocksize;
    len -= blocksize;
  }
  memcpy (buffer, &data[i], len);
}


/* Add padding */

void
mdblock::finish_le ()
{
  u_char *dp;
  u_char *end;
  u_int bcount = count % blocksize;

  if (bcount > blocksize - 9) {
    /* need to split padding bit and count */
    u_int8_t tmp[blocksize];
    bzero (tmp, blocksize - bcount);
    /* add padding bit */
    tmp[0] = 0x01;
    update (tmp, blocksize - bcount);
    /* don't count padding in length of string */
    count -= blocksize - bcount;
    dp = buffer;
  }
  else {
    dp = &buffer[bcount];
    *dp++ = 0x01;
  }
  end = &buffer[blocksize - 8];
  while (dp < end)
    *dp++ = 0;
  count <<= 3;			/* make bytecount bitcount */

  *dp++ = (count >> 0) & 0xff;
  *dp++ = (count >> 8) & 0xff;
  *dp++ = (count >> 16) & 0xff;
  *dp++ = (count >> 24) & 0xff;
  *dp++ = (count >> 32) & 0xff;
  *dp++ = (count >> 40) & 0xff;
  *dp++ = (count >> 48) & 0xff;
  *dp++ = (count >> 56) & 0xff;

  /* Wipe variables */
  dp = end = 0;

  consume (buffer);
}

void
mdblock::finish_be ()
{
  u_char *dp;
  u_char *end;
  u_int bcount = count % blocksize;

  if (bcount > blocksize - 9) {
    /* need to split padding bit and count */
    u_int8_t tmp[blocksize];
    bzero (tmp, blocksize - bcount);
    /* add padding bit */
    tmp[0] = 0x80;
    update (tmp, blocksize - bcount);
    /* don't count padding in length of string */
    count -= blocksize - bcount;
    dp = buffer;
  }
  else {
    dp = &buffer[bcount];
    *dp++ = 0x80;
  }
  end = &buffer[blocksize - 8];
  while (dp < end)
    *dp++ = 0;
  count <<= 3;			/* make bytecount bitcount */

  *dp++ = (count >> 56) & 0xff;
  *dp++ = (count >> 48) & 0xff;
  *dp++ = (count >> 40) & 0xff;
  *dp++ = (count >> 32) & 0xff;
  *dp++ = (count >> 24) & 0xff;
  *dp++ = (count >> 16) & 0xff;
  *dp++ = (count >> 8) & 0xff;
  *dp++ = (count >> 0) & 0xff;

  /* Wipe variables */
  dp = end = 0;

  consume (buffer);
}

void
mdblock::updatev (const iovec *iov, u_int cnt)
{
  for (const iovec *end = iov + cnt; iov < end; iov++)
    update (iov->iov_base, iov->iov_len);
}
