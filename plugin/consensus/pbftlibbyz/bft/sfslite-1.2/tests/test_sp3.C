
#include "async.h"
#include "sp_gc.h"
#include "sp_wkref.h"
#include "sp_gc_str.h"
#include "crypt.h"

class rand_str_t {
public:
  rand_str_t (const char *c, size_t o, size_t len)
    : _s (c, len), _offset (o), _len (len) {}

  sp::gc::str gets () const { return _s; }
  size_t offset () const { return _offset; }
  size_t len () const { return _len; }

private:
  sp::gc::str _s;
  const char *_p;
  size_t _offset;
  size_t _len;
};

class rand_data_t {
public:
  rand_data_t (size_t mx) : _maxsz (mx)
  {
    if (mx >= bufsz) mx = bufsz / 2;
    rnd.getbytes (_buf, bufsz);
  }

  const char *dat (size_t off, size_t len) const
  {
    assert (off < bufsz);
    assert (off + len < bufsz);
    return _buf + off;
  }

  sp::gc::ptr<rand_str_t> produce () const
  {
    size_t sz = rnd.getword () % _maxsz;
    u_int32_t w = rnd.getword ();
    w = w % (bufsz - sz);
    sp::gc::ptr<rand_str_t> r = sp::gc::alloc<rand_str_t> (dat (w, sz), w, sz);
    assert (r);
    assert (r->gets ());
    return r;
  }

  bool check (const rand_str_t &s) const
  {
    if (s.len () != s.gets ().len ()) return false;

    return (memcmp (s.gets ().volatile_cstr (),
		    _buf + s.offset (), 
		    s.len ()) == 0);
  }

private:
  size_t _maxsz;
  enum { bufsz = 0x10000 };
  char _buf[bufsz];
};

static void
check (const rand_data_t &rd, const vec<sp::gc::ptr<rand_str_t> > &v)
{
  sp::gc::ptr<rand_str_t> x;
  for (size_t i = 0; i < v.size (); i++) {
    x = v[i];
    if (x) {
      rand_str_t tmp = *x;
      assert (rd.check (tmp));
    }
  }
}

static void
test0 (void)
{
  sp::gc::smallobj_sizer_t sz;
  for (size_t s = 1; s < 1025; s++) {
    int i;
    size_t t = sz.find (s, &i);
    size_t u = sz.ind2size (i);
    size_t l = 0;
    if (i > 0)
      l = sz.ind2size (i-1);
    assert (t == u);
    assert (s <= u);
    assert (s > l);
  }

}

static void
test1 (void)
{
  vec<sp::gc::ptr<rand_str_t> > v;
  sp::gc::ptr<rand_str_t> x;
  rand_data_t rd (0x2000);
  for (size_t i = 0; i < 1000; i++) {
    x = rd.produce ();
    assert (x);
    v.push_back (x);
  }

  check (rd, v);

  for (size_t i = 0; i < v.size (); i += 2) {
    v[i] = NULL;
  }

  check (rd, v);

  for (size_t i = 0; i < v.size (); i += 13) {
    v[i] = NULL;
  }

  sp::gc::mgr_t<>::get ()->sanity_check();
  check (rd, v);

  for (size_t i = 0; i < v.size (); i++ ) {
    if (!v[i]) {
      sp::gc::mgr_t<>::get ()->sanity_check();
      x = rd.produce ();
      assert (x);
      v[i] = x;
      sp::gc::mgr_t<>::get ()->sanity_check();
    }
  }

  check (rd, v);
  sp::gc::mgr_t<>::get ()->sanity_check();

  for (size_t i = 0; i < v.size (); i++ ) {
    sp::gc::mgr_t<>::get ()->sanity_check();
    x = rd.produce ();
    assert (x);
    v[i] = x;
  }

  check (rd, v);
}

static void
test2 (void)
{
  rand_data_t rd (0x10000);
  size_t len = 0x2000;
  const char *s = rd.dat (0, len);
  sp::gc::str ss (s, len);
}


int
main (int argc, char *argv[])
{
  test0 ();

  random_start ();
  random_init ();

  sp::gc::std_cfg_t cfg;
  cfg._n_b_arenae = 1024;
  cfg._size_b_arenae = 0x10000;
  cfg._smallobj_lim = 1024;
  cfg._smallobj_min_obj_per_arena = 1;
  sp::gc::mgr_t<>::set (New sp::gc::std_mgr_t<> (cfg));

  sp::gc::debug_warnings = true;

  test2 ();
  for (int i = 0; i < 10; i++) {
    sp::gc::mgr_t<>::get ()->sanity_check();
    test1();
  }
}

