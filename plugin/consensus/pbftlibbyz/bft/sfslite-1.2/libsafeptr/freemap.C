
#include "freemap.h"

//=======================================================================

freemap_t::node_t::node_t (u_int32_t i) : _id (i), _bits (0) {}

bool
freemap_t::node_t::getbit (u_int i) const
{
  assert (i < n_bits);
  return (_bits & (1 << i));
}

//-----------------------------------------------------------------------

size_t
freemap_t::node_t::nfree () const 
{
  size_t r = 0;
  u_int64_t b = _bits;
  for (int i = 0; i < n_bits; i++) {
    r += (b & 1);
    b = b >> 1;
  }
  return r;
}

//-----------------------------------------------------------------------

void
freemap_t::node_t::setbit (u_int i, bool b) 
{
  assert (i < n_bits);
  if (b) {
    _bits = _bits | (1 << i);
  } else {
    _bits = _bits & (~(1 << i));
  }
}

//-----------------------------------------------------------------------

int
freemap_t::node_t::topbit () const
{
  int ret = -1;
  if (!is_empty ()) {
    for (int i = n_bits - 1; ret < 0 && i >= 0; i--) {
      if (getbit (i)) 
	ret = i;
    }
  }
  return ret;
}

//-----------------------------------------------------------------------

bool
freemap_t::node_t::is_empty () const
{
  return (_bits == 0);
}

//-----------------------------------------------------------------------

int
freemap_t::node_t::global_id (u_int i) const
{
  assert (i < n_bits);
  return _id * n_bits + i;
}

//-----------------------------------------------------------------------

int 
freemap_t::node_t::cmp (u_int32_t segid) const
{
  return (segid - _id);
}

//=======================================================================

void
freemap_t::dealloc (u_int i)
{
  u_int32_t segid = node_t::segid (i);
  u_int bitid = node_t::bitid (i);
  node_t *n = find (segid);
  if (!n) {
    n = New node_t (segid);
    _segs.insert (n);
  }
  assert (n);
  assert (!n->getbit (bitid));
  n->setbit (bitid, true);
}

//-----------------------------------------------------------------------

int
freemap_t::alloc () 
{
  int ret;
  node_t *n = findmax ();
  if (!n) {
    ret = -1;
  } else {
    int b = n->topbit ();
    assert (b >= 0);
    n->setbit (b, false);
    ret = n->global_id (b);
    if (n->is_empty ()) {
      _segs.remove (n);
      delete n;
    }
  }
  return ret;
}

//-----------------------------------------------------------------------

freemap_t::freemap_t () {}

//-----------------------------------------------------------------------

freemap_t::~freemap_t ()
{
  _segs.deleteall ();
}

//-----------------------------------------------------------------------

freemap_t::node_t *
freemap_t::findmax ()
{
  freemap_t::node_t *n, *nn;
  for (n = _segs.root (); 
       n && ((nn = _segs.right (n)) || (nn = _segs.left (n))); 
       n = nn) ;
  return n;
}

static int find_fn (u_int32_t id, const freemap_t::node_t *n)
{
  return n->cmp (id);
}

//-----------------------------------------------------------------------

freemap_t::node_t *
freemap_t::find (u_int32_t segid)
{
  return _segs.search (wrap (find_fn, segid));
}

//-----------------------------------------------------------------------

size_t
freemap_t::nfree () const
{
  size_t s = 0;
  const node_t *n, *nn;
  for (n = _segs.first (); n; n = nn) {
    nn = _segs.next (n);
    s += n->nfree ();
  }
  return s;
}

//-----------------------------------------------------------------------

