// -*-c++-*-
/* $Id: parseopt.h 3146 2007-12-20 17:44:02Z max $ */

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

#ifndef _ASYNC_PARSEOPT_
#define _ASYNC_PARSEOPT_ 1

#include "vec.h"
#include "str.h"
#include "ihash.h"
#include "amisc.h"

class parseargs {

  char *buf;
  const char *lim;
  const char *p; 

  void skipblanks ();
  void skiplwsp ();
  str getarg ();

protected:
  str filename;
  int lineno;

  virtual void error (str);

public:
  parseargs (str file, int fd = -1);
  virtual ~parseargs ();

  bool getline (vec<str> *args, int *linep = NULL);
};

int64_t strtoi64 (const char *nptr, char **endptr = NULL, int base = 0);

template<class T> bool
convertint (const char *cp, T *resp)
{
  if (!*cp)
    return false;
  char *end;
  T res = strtoi64 (cp, &end, 0);
  if (*end)
    return false;
  *resp = res;
  return true;
}

void mytolower (char *dest, const char *src);
str mytolower (const str &in);

class conftab_el {
public:
  conftab_el (const str &n) : 
    name (n), lcname (mytolower (n)), _was_set (false),
    _set_by_default (false) {}

  virtual ~conftab_el () {}
  virtual bool convert (const vec<str> &v, const str &loc, bool *e) = 0;
  virtual bool inbounds () = 0;
  virtual void set () = 0;
  inline bool count_args (const vec<str> &v, u_int l)
  { return (v.size () == l || (v.size () > l && v[l][0] == '#')); }

  const str &get_name () const { return name; }

  inline bool was_set () const { return _was_set; }
  void reset () { _was_set = false; _set_by_default = false; }
  void mark_set () { _was_set = true; }
  virtual bool apply_default () = 0;
  void mark_set_by_default () { _set_by_default = true; }
  bool was_set_by_default () const { return _set_by_default; }
  virtual void dump (strbuf &b) const = 0;

  const str name;
  const str lcname;
  ihash_entry<conftab_el> lnk;
private:
  bool _was_set, _set_by_default;
};

class conftab_ignore : public conftab_el {
public:
  conftab_ignore (const str &n) : conftab_el (n) {}
  bool convert (const vec<str> &v, const str &loc, bool *e) { return true; }
  bool inbounds () { return true; }
  void set () {}
  bool apply_default () { return false; }
  void dump (strbuf &b) const { b << "IGNORED"; }
};

typedef callback<void, vec<str>, str, bool *>::ref confcb;
class conftab_str : public conftab_el {
public:
  conftab_str (const str &n, str *d, bool c) 
    : conftab_el (n), dest (d), cnfcb (NULL), scb (NULL), check (c),
      _has_default (false) {}
  conftab_str (const str &n, confcb c)
    : conftab_el (n), dest (NULL), cnfcb (c), scb (NULL), check (false),
      _has_default (false) {}

  // XXX: reverse order to disambiguate
  conftab_str (cbs c, const str &n)
    : conftab_el (n), dest (NULL), cnfcb (NULL), scb (c), check (false),
      _has_default (false) {}

  // Can supply a default value for this string.
  conftab_str (const str &n, str *d, const str &def, bool c)
    : conftab_el (n), dest (d), cnfcb (NULL), scb (NULL), check (c),
      _default (def), _has_default (true) {}

  bool convert (const vec<str> &v, const str &l, bool *e);
  bool inbounds () { return true; }
  void set ();

  bool apply_default ()
  { if (_has_default) *dest = _default; return _has_default; }

  void dump (strbuf &b) const 
  {
    if (*dest) b << "\"" << *dest << "\"";
    else b << "(null)";
  }

private:
  str *const dest;
  confcb::ptr cnfcb;
  cbs::ptr scb;
  const bool check;

  vec<str> tmp;
  str tmp_s;
  str loc;
  bool *errp;

  const str _default;
  bool _has_default;
};

template<class T>
class conftab_int : public conftab_el {
public:

  // If no default value provided...
  conftab_int (const str &n, T *d, T l, T u)
    : conftab_el (n), dest (d), lb (l), ub (u),
      _default (0), _has_default (false) {}

  // If default value provided...
  conftab_int (const str &n, T *d, T l, T u, T def)
    : conftab_el (n), dest (d), lb (l), ub (u),
      _default (def), _has_default (true) {}
      
  bool convert (const vec<str> &v, const str &cf, bool *e)
  { return (count_args (v, 2) && convertint (v[1], &tmp)); }
  bool inbounds () { return (tmp >= lb && tmp <= ub); }
  void set () { *dest = tmp; }

  bool apply_default ()
  { if (_has_default) *dest = _default; return _has_default; }

  void dump (strbuf &b) const { b << *dest; }

private:
  T *const dest;
  const T lb;
  const T ub;
  T tmp;

  T _default;
  bool _has_default;
};

class conftab_bool : public conftab_el {
public:
  conftab_bool (const str &n, bool *b) 
    : conftab_el (n), dest (b), err (false), _has_default (false) {}

  conftab_bool (const str &n, bool *b, bool def) 
    : conftab_el (n), dest (b), err (false),
      _default (def), _has_default (true) {}

  bool convert (const vec<str> &v, const str &cf, bool *e);
  bool inbounds () { return !(err); }
  void set () { *dest = tmp; }

  bool apply_default ()
  { if (_has_default) *dest = _default; return _has_default; }

  void dump (strbuf &b) const { b << (*dest ? "True" : "False"); }

private:
  bool tmp;
  bool *dest;
  bool err;

  bool _default;
  bool _has_default;
};

enum {
  CONFTAB_VERBOSE = 0x1,
  CONFTAB_APPLY_DEFAULTS = 0x2
}; 

class conftab {
public:
  conftab () {}

  typedef enum { OK = 0,
		 ERROR = 1,
		 UNKNOWN = 2 } status_t;
  
  bool run (const str &file, u_int opts = 0, int fd = -1,
            status_t *sp = NULL);

  template<class P, class D> 
  conftab &add (const str &nm, P *dp, D lb, D ub)
  { return insert (New conftab_int<P> (nm, dp, lb, ub)); }

  template<class P, class D> 
  conftab &add (const str &nm, P *dp, D lb, D ub, P def)
  { return insert (New conftab_int<P> (nm, dp, lb, ub, def)); }

  template<class A>
  conftab &add (const str &nm, A a) 
  { return insert (New conftab_str (nm, a)); }

  conftab &ads (const str &nm, cbs c) // XXX: cannot overload
  { return insert (New conftab_str (c, nm)); }

  conftab &add (const str &nm, str *s)
  { return insert (New conftab_str (nm, s, false)); }

  conftab &add_check (const str &nm, str *s)
  { return insert (New conftab_str (nm, s, true)); }

  conftab &add (const str &nm, str *s, const str &def)
  { return insert (New conftab_str (nm, s, def, false)); }

  conftab &add_check (const str &nm, str *s, const str &def)
  { return insert (New conftab_str (nm, s, def, true)); }

  conftab &add (const str &nm, bool *b) 
  { return insert (New conftab_bool (nm, b)); }

  conftab &add (const str &nm, bool *b, bool def) 
  { return insert (New conftab_bool (nm, b, def)); }

  conftab &ignore (const str &m)
  { return insert (New conftab_ignore (m)); }

  void apply_defaults ();
  void report ();
  void report (vec<str> *out);
  void reset ();

  conftab &insert (conftab_el *e)
  {
    tab.insert (e);
    _v.push_back (e);
    return *this;
  }

  ~conftab () { tab.deleteall (); }

  bool match (const vec<str> &s, const str &cf, int ln, bool *err);
private:
  ihash<const str, conftab_el, &conftab_el::lcname, &conftab_el::lnk> tab;
  vec<conftab_el *> _v;
};


#endif /* !_ASYNC_PARSEOPT_ */
