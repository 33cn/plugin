/* $Id: rxx.C 3467 2008-07-01 03:49:10Z max $ */

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

#include "rxx.h"
#include "amisc.h"

// the default behavior is to panic on any rxx errors.
bool sfs_rxx_panic = true;


struct rcbase {
  enum { magicval = int (0xa5e10288) };
  int32_t cnt;
  int32_t magic;
};

void *
rcmalloc (size_t n)
{
  rcbase *r = static_cast<rcbase *> (xmalloc (sizeof (rcbase) + n));
  r->cnt = 1;
  r->magic = rcbase::magicval;
  // warn ("alloc %p\n", r);
  return r + 1;
}

void
rcfree (void *p)
{
  if (p) {
    rcbase *r = static_cast<rcbase *> (p) - 1;
    assert (r->magic == rcbase::magicval);
    if (!--r->cnt) {
      // warn ("free %p\n", r);
      r->magic = 0;
      xfree (r);
    }
    else
      assert (r->cnt > 0);
  }
}

void *
rccopy (void *p)
{
  if (p) {
    rcbase *r = static_cast<rcbase *> (p) - 1;
    assert (r->magic == rcbase::magicval);
    r->cnt++;
  }
  return p;
}

int rxxinit::count;
void
rxxinit::start ()
{
  pcre_malloc = rcmalloc;
  pcre_free = rcfree;
}
void
rxxinit::stop ()
{
}

str
rxx::init (const char *pat, const char *opt)
{
  extra = NULL;
  nsubpat = 0;
  ovector = NULL;
  bool studyit = false;

  int options = 0;
  for (; *opt; opt++)
    switch (*opt) {
    case '^':
      options |= PCRE_ANCHORED;
      break;
    case 'i':
      options |= PCRE_CASELESS;
      break;
    case 's':
      options |= PCRE_DOTALL;
      break;
    case 'm':
      options |= PCRE_MULTILINE;
      break;
    case 'x':
      options |= PCRE_EXTENDED;
      break;
    case 'U':
      options |= PCRE_UNGREEDY;
      break;
    case 'X':
      options |= PCRE_EXTRA;
      break;
    case 'S':
      studyit = true;
      break;
    default:
      return strbuf ("invalid regular expression option '%c'\n", *opt);
    }

  const char *errptr;
  int erroffset;
  re = pcre_compile (pat, options, &errptr, &erroffset, NULL);
  if (!re) {
    strbuf err;
    err << "Invalid regular expression:\n"
	<< "   " << pat << "\n";
    suio_fill (err.tosuio (), ' ', erroffset);
    err << "   ^\n"
	<< errptr << ".\n";
    return err;
  }
  if (studyit) {
    str err = study ();
    if (err)
      return strbuf () << "Could not study regular expression: " << err;
  }

  int ns = pcre_info (re, NULL, NULL);
  assert (ns >= 0);
  ovecsize = (ns + 1) * 3;
  return NULL;
}

bool
rxx::_exec (const char *p, size_t len, int options)
{
  bool ok = true;
  subj = NULL;
  _errcode = 0;
		
  if (!ovector)
    ovector = New int[ovecsize];
  nsubpat = pcre_exec (re, extra, p, len, 0,
		       options, ovector, ovecsize);
  if (nsubpat <= 0 && nsubpat != PCRE_ERROR_NOMATCH)  {
    _errcode = nsubpat;
    ok = false;
    if (sfs_rxx_panic) {
      panic ("rxx/pcre_exec error %d\n", nsubpat);
    } else {
      warn ("rxx/pcre_exec error %d\n", nsubpat);
      nsubpat = 0;
    }
  }
  return ok;
}

bool
rxx::exec (str s, int options)
{
  bool ok = true;
  subj = s;
  _errcode = 0;

  if (!ovector)
    ovector = New int[ovecsize];
  nsubpat = pcre_exec (re, extra, s.cstr (), s.len (), 0,
		       options, ovector, ovecsize);
  if (nsubpat <= 0 && nsubpat != PCRE_ERROR_NOMATCH) {
    _errcode = nsubpat;
    ok = false;
    if (sfs_rxx_panic) {
      panic ("rxx/pcre_exec error %d\n", nsubpat);
    } else {
      warn ("rxx/pcre_exec error %d\n", nsubpat);
      nsubpat = 0;
    }
  }
  return ok;
}

str
rxx::at (ptrdiff_t n) const
{
  assert (n >= 0);
  if (n >= nsubpat)
    return NULL;
  size_t i = 2 * n;
  if (ovector[i] == -1)
    return NULL;
  return str (subj.cstr () + ovector[i], ovector[i+1] - ovector[i]);
}

rxx &
rxx::operator= (const rxx &r)
{
  if (&r != this) {
    this->~rxx ();
    copy (r);
  }
  return *this;
}

int
split (vec<str> *out, rxx pat, str expr, size_t lim, bool emptylast)
{
  const char *p = expr;
  const char *const e = p + expr.len ();
  size_t n;
  if (out)
    out->clear ();

  for (n = 0; n + 1 < lim; n++) {
    if (!pat._exec (p, e - p, 0)) {
      return 0;
    }
    if (!pat.success ())
      break;
    if (out)
      out->push_back (str (p, pat.start (0)));
    p += max (pat.end (0), 1);
  }

  if (lim && (p < e || emptylast)) {
    n++;
    if (out)
      out->push_back (str (p, e - p));
  }
  return n;
}

str
join (str sep, const vec<str> &v)
{
  strbuf sb;
  const str *sp = v.base ();
  if (sp < v.lim ()) {
    sb.cat (*sp++);
    while (sp < v.lim ())
      sb.cat (sep).cat (*sp++);
  }
  return sb;
}
