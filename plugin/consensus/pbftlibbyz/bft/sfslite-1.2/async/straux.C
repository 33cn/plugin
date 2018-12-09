/* $Id: straux.C 1117 2005-11-01 16:20:39Z max $ */

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


#include "amisc.h"

/* This is the memory equivalent of strpbrk().  It finds the first
 * occurance of any character from s2 in s1, but takes a length
 * argument for s1 rather than stopping at a null byte. */
char *
mempbrk (char *s1, const char *s2, int len)
{
  const char *const eom = s1 + len;
  register const char *cp;
  register int i1, i2;

  while (s1 < eom) {
    i1 = *s1++;
    cp = s2;
    while ((i2 = *cp++))
      if (i1 == i2)
	return (s1 - 1);
  }
  return (NULL);
}

/* This is a strsep function that returns a null field for adjacent
 * separators.  This is the same as the 4.4BSD strsep, but different
 * from the one in the GNU libc. */
char *
xstrsep(char **str, const char *delim)
{
  char *s, *e;

  if (!**str)
    return (NULL);

  s = *str;
  e = s + strcspn (s, delim);

  if (*e != '\0')
    *e++ = '\0';
  *str = e;

  return (s);
}

/* Get the next non-null token (like GNU strsep).  Strsep() will
 * return a null token for two adjacent separators, so we may have to
 * loop. */
char *
strnnsep (char **stringp, const char *delim)
{
  char *tok;

  do {
    tok = xstrsep (stringp, delim);
  } while (tok && *tok == '\0');
  return (tok);
}

#ifndef HAVE_STRCASECMP

/*
 * This array is designed for mapping upper and lower case letter
 * together for a case independent comparison.  The mappings are
 * based upon ascii character sequences.
 */
static const u_char charmap[] =
{
  '\000', '\001', '\002', '\003', '\004', '\005', '\006', '\007',
  '\010', '\011', '\012', '\013', '\014', '\015', '\016', '\017',
  '\020', '\021', '\022', '\023', '\024', '\025', '\026', '\027',
  '\030', '\031', '\032', '\033', '\034', '\035', '\036', '\037',
  '\040', '\041', '\042', '\043', '\044', '\045', '\046', '\047',
  '\050', '\051', '\052', '\053', '\054', '\055', '\056', '\057',
  '\060', '\061', '\062', '\063', '\064', '\065', '\066', '\067',
  '\070', '\071', '\072', '\073', '\074', '\075', '\076', '\077',
  '\100', '\141', '\142', '\143', '\144', '\145', '\146', '\147',
  '\150', '\151', '\152', '\153', '\154', '\155', '\156', '\157',
  '\160', '\161', '\162', '\163', '\164', '\165', '\166', '\167',
  '\170', '\171', '\172', '\133', '\134', '\135', '\136', '\137',
  '\140', '\141', '\142', '\143', '\144', '\145', '\146', '\147',
  '\150', '\151', '\152', '\153', '\154', '\155', '\156', '\157',
  '\160', '\161', '\162', '\163', '\164', '\165', '\166', '\167',
  '\170', '\171', '\172', '\173', '\174', '\175', '\176', '\177',
  '\200', '\201', '\202', '\203', '\204', '\205', '\206', '\207',
  '\210', '\211', '\212', '\213', '\214', '\215', '\216', '\217',
  '\220', '\221', '\222', '\223', '\224', '\225', '\226', '\227',
  '\230', '\231', '\232', '\233', '\234', '\235', '\236', '\237',
  '\240', '\241', '\242', '\243', '\244', '\245', '\246', '\247',
  '\250', '\251', '\252', '\253', '\254', '\255', '\256', '\257',
  '\260', '\261', '\262', '\263', '\264', '\265', '\266', '\267',
  '\270', '\271', '\272', '\273', '\274', '\275', '\276', '\277',
  '\300', '\301', '\302', '\303', '\304', '\305', '\306', '\307',
  '\310', '\311', '\312', '\313', '\314', '\315', '\316', '\317',
  '\320', '\321', '\322', '\323', '\324', '\325', '\326', '\327',
  '\330', '\331', '\332', '\333', '\334', '\335', '\336', '\337',
  '\340', '\341', '\342', '\343', '\344', '\345', '\346', '\347',
  '\350', '\351', '\352', '\353', '\354', '\355', '\356', '\357',
  '\360', '\361', '\362', '\363', '\364', '\365', '\366', '\367',
  '\370', '\371', '\372', '\373', '\374', '\375', '\376', '\377',
};

int
strcasecmp (const char *s1, const char *s2)
{
  const u_char *cm = charmap,
    *us1 = (const u_char *) s1,
    *us2 = (const u_char *) s2;

  while (cm[*us1] == cm[*us2++])
    if (*us1++ == '\0')
      return (0);
  return (cm[*us1] - cm[*--us2]);
}

int
strncasecmp (const char *s1, const char *s2, int n)
{
  if (n != 0) {
    const u_char *cm = charmap,
      *us1 = (const u_char *) s1,
      *us2 = (const u_char *) s2;
    do {
      if (cm[*us1] != cm[*us2++])
	return (cm[*us1] - cm[*--us2]);
      if (*us1++ == '\0')
	break;
    } while (--n != 0);
  }
  return (0);
}

#endif /* !HAVE_STRCASECMP */
