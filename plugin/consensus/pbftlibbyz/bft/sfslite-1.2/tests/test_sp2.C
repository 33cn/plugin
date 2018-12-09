
#include "async.h"
#include "sp_gc.h"
#include "sp_wkref.h"
#include "sp_gc_str.h"

class foo_t {
public:
  foo_t (int a) : _x (a) {}
  foo_t (int a, int b) : _x (a+b) {}
  void bar () { warn << "foo_t() => " << _x << "\n"; }
  int baz () const { return _x; }
private:
  int _x;
  char _pad[31];
};

static void
test2(void) {

  sp::gc::ptr<foo_t> x = sp::gc::alloc<foo_t> (10,23);
  sp::gc::ptr<foo_t> y = x;
  assert (y->baz () == 33);
  assert (x->baz () == 33);
  x = y = NULL;
  x = sp::gc::alloc<foo_t> (40,50);

  vec<sp::gc::ptr<foo_t> > v;
  for (int j = 0; j < 10; j++) {
    for (size_t i = 0; i < 100; i++) {
      x = sp::gc::alloc<foo_t> (i, 0);
      assert (x);
      v.push_back (x);
    }
    sp::gc::mgr_t<>::get ()->report ();
    v.clear ();
  }


  for (size_t i = 0; i < 100; i++) {
    v.push_back (sp::gc::alloc<foo_t> (2*i, 0));
  }

  for (size_t i = 0; i < 20; i++) {
    v[i*5] = NULL;
  }

  for (size_t i = 0; i < 20; i++) {
    x = sp::gc::alloc<foo_t> (i,300);
    assert (x);
    v[i*5] = x;
  }

  for (size_t i = 0; i < 100; i++) {
    size_t x = v[i]->baz ();
    if (i % 5 == 0) {
      assert (x == 300 + i/5);
    } else {
      assert (x == 2*i);
    }
  }
}

int
main (int argc, char *argv[])
{
  sp::gc::std_cfg_t cfg;
  cfg._n_b_arenae = 2;
  cfg._size_b_arenae = 1;
  cfg._smallobj_lim = 1024;
  cfg._smallobj_min_obj_per_arena = 1;
  sp::gc::mgr_t<>::set (New sp::gc::std_mgr_t<> (cfg));

  for (int i = 0; i < 10; i++) {
    sp::gc::mgr_t<>::get ()->sanity_check();
    test2();
  }
}

