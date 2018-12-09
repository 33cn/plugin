
#include "sp_gc_str.h"

namespace sp {
namespace gc {

  //=======================================================================

  strobj::strobj (const char *p) 
    : _len (strlen (p)),
      _p (sp::gc::vecalloc<char> (_len + 1))
  {
    if (_p) {
      memcpy (_p.volatile_ptr (), p, _len);
      _p.volatile_ptr()[_len] = '\0';
    }
  }

  //-----------------------------------------------------------------------

  strobj::strobj (const char *p, size_t l)
    : _len (l),
      _p (sp::gc::vecalloc<char> (_len + 1))
  {
    if (_p) {
      memcpy (_p.volatile_ptr (), p, _len);
      _p.volatile_ptr()[_len] = '\0';
    }
  }

  //=======================================================================

  const char *
  str::volatile_cstr () const
  {
    return _o._p ? _o._p.volatile_ptr () : NULL;
  }

  //-----------------------------------------------------------------------

  char 
  str::operator[] (ptrdiff_t n) const
  {
    assert (_o._p);
    assert (size_t (n) <= _o._len);
    return volatile_cstr()[n];
  }

  //-----------------------------------------------------------------------

  int 
  str::cmp (const str &s) const
  {
    if (int r = memcmp (volatile_cstr (), s.volatile_cstr (), 
			min (len (), s.len ())))
      return r;
    return len () - s.len ();
  }

  //-----------------------------------------------------------------------
  
  int 
  str::cmp (const char *p) const {
    const char *s = volatile_cstr ();
    const char *e = s + len ();
    while (*s == *p)
      if (!*p++)
	return e - s;
      else if (s++ == e)
	return -1;
    return (u_char) *s - (u_char) *p;
  }

  //-----------------------------------------------------------------------

  bool 
  str::operator== (const str &s) const
  {
    return (len () == s.len () && 
	    !memcmp (volatile_cstr (), s.volatile_cstr (), len ()));
  }

  //-----------------------------------------------------------------------

  bool
  str::operator== (const char *p) const
  {
    if (!p && !_o._p) return true;
    else if (!p || !_o._p) return false;
    else return cmp(p) == 0;
  }

  //-----------------------------------------------------------------------

  hash_t
  str::to_hash () const
  {
    const char *s = volatile_cstr ();
    assert (s);
    return hash_bytes (s, len ());
  }

  //=======================================================================

};
};
