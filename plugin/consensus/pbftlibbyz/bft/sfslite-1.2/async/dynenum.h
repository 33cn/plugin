
// -*-c++-*-
/* $Id: qhash.h 2885 2007-05-18 23:10:26Z max $ */

#include "async.h"
#include "qhash.h"

#ifndef _ASYNC_DYNENUM_H_INCLUDED_
#define _ASYNC_DYNENUM_H_INCLUDED_

class dynamic_enum_t {
public:
  struct pair_t {
    const char *n;
    int v;
  };

  dynamic_enum_t (int def, bool quiet = false, str n = NULL) 
    : _def_val (def), _quiet (quiet), _enum_name (n) {}
  virtual ~dynamic_enum_t () {}

  int operator[] (const str &s) const { return lookup (s, !_quiet); }

  bool lookup (const str &s, int *v) const;
  int lookup (const str &s, bool dowarn = true) const ;

protected:
  void init (const pair_t pairs[], bool chk = false);
  virtual void warn_not_found (str s) const;
private:
  const int _def_val;
  bool _quiet;
  const str _enum_name;
  qhash<str, int> _tab;
};

#endif /* _ASYNC_DYNENUM_H_INCLUDED_ */
