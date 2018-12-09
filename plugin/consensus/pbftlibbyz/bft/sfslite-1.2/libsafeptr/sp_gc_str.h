

// -*- c++ -*-

#include "sp_gc.h"


namespace sp {
namespace gc {

  struct strobj {
    strobj () : _len (0) {}
    strobj (const char *s);
    strobj (const char *s, size_t len);
    size_t _len;
    ptr<char> _p;
  };

  class str {
  public:
    str () {}
    str (const str &s) : _o (s._o) {}
    str (const char *p) : _o (p) {}
    str (const char *p, size_t l) : _o (p, l) {}
    ~str () {}

    str &operator= (const str &s) { _o = s._o; return (*this); }
    str &operator= (const char *p) { _o = strobj (p); return (*this); }

    size_t len () const { return _o._len; }
    const char *volatile_cstr () const ;

    char operator[] (ptrdiff_t n) const;
    int cmp (const str &s) const;
    int cmp (const char *p) const;
    ::str copy () const { return ::str (volatile_cstr (), len ()); }
    operator bool () const { return _o._p; }
    
    bool operator== (const str &s) const;
    bool operator!= (const str &s) const { return !(*this == s); }
    bool operator< (const str &s) const { return cmp(s) < 0; }
    bool operator<= (const str &s) const { return cmp(s) <= 0; }
    bool operator> (const str &s) const { return cmp(s) > 0; }
    bool operator>= (const str &s) const { return cmp(s) >= 0; }

    bool operator== (const char *p) const;
    bool operator!= (const char *p) const { return ! (*this == p); }
    operator hash_t () const { return to_hash (); }
    hash_t to_hash () const;
  private:
    
    strobj _o;
  };

};
};
