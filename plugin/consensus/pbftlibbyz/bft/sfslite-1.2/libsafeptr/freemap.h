

// -*- c++ -*-

#ifndef __LIBSAFEPTR_FREEMAP_H__
#define __LIBSAFEPTR_FREEMAP_H__

#include "async.h"
#include "itree.h"


#undef setbit

class freemap_t {
public:
  freemap_t ();
  ~freemap_t ();

  struct node_t {
    node_t (u_int32_t id);
    ~node_t () {}
    
    bool getbit (u_int i) const;
    void setbit (u_int i, bool b);
    bool is_empty () const;
    int topbit () const;
    int global_id (u_int i) const;
    int cmp (u_int32_t segid) const;
    size_t nfree () const;
    
    enum { n_bits = 64 };
    
    static u_int32_t segid (u_int i) { return i / n_bits; }
    static u_int bitid (u_int i) { return i % n_bits; }
    
    u_int32_t _id;
    itree_entry<node_t> _lnk;
  private:
    u_int64_t _bits;
  };

  int alloc ();
  void dealloc (u_int i);
  size_t nfree () const;

  static size_t fixed_overhead () { return sizeof (node_t); }

  static size_t bits_per_slot ()
  { return sizeof (node_t) / node_t::n_bits; }

private:
  node_t *find (u_int32_t s);
  node_t *findmax ();
  itree<u_int32_t, node_t, &node_t::_id, &node_t::_lnk> _segs;
};

#endif /* __LIBSAFEPTR_FREEMAP_H__ */
