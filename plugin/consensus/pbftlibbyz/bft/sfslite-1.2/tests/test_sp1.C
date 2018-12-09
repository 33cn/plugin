
#include "async.h"
#include "sp_gc.h"
#include "sp_wkref.h"
#include "sp_gc_str.h"

class bar_t {
public:
  bar_t (int a) : _x (a) {}
  int val () const { return _x; }
private:
  int _x;
  char _pad[1501];
};

static void
test1(void)
{
  for (size_t j = 0; j < 10; j++) {

    // Note; this x must be deallocated by the end of the loop, otherwise
    // it will hold a reference to an bar_t and we won't have spcae.
    // There's only enough room for exactly 4 of these things.
    sp::gc::ptr<bar_t> x; 
    vec<sp::gc::ptr<bar_t> > v;
    for (size_t i = 0; i < 4; i++) {
      sp::gc::ptr<bar_t> x = sp::gc::alloc<bar_t> (i);
      assert (x);
      v.push_back (x);
    }

    // Must clear v[0] before allocating new one, otherwise we're out of 
    // space.
    v[0] = NULL;
    x = sp::gc::alloc<bar_t> (100);
    assert (x);
    v[0] = x;

    for (int i = 1; i < 4; i++) {
      assert (v[i]->val () == i);
    }

    assert (v[0]->val () == 100);
  }
}

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

static void
test3()
{
  sp::gc::ptr<bool> b = sp::gc::alloc<bool> (true);
  *b = false;
  sp::gc::ptr<char> c = sp::gc::vecalloc<char> (1000);
}

class wrobj_t : public sp::referee<wrobj_t> 
{
public:
  wrobj_t (int i) : _i (i) {}
  int val () const { return _i; }
private:
  int _i;
};

static void
test0 ()
{
  if (true) {
    int val = 10;
    wrobj_t foo (val);
    sp::wkref<wrobj_t> r (foo);
    assert (r->val () == val);
  }

  if (true) {
    int val = 20;
    sp::man_ptr<wrobj_t> x = sp::man_alloc<wrobj_t> (val);
    sp::man_ptr<wrobj_t> y = x;
    assert (x->val () == val);
    assert (y->val () == val);
    x.dealloc ();
    assert (!x);
    assert (!y);
  }
}

static void
test4 ()
{
  const char *s = "hello everyone; WTF is up?" ;
  sp::gc::str foo (s);
  assert (foo == s);

  // test that operator[] works
  mstr m (foo.len ());
  for (size_t i = 0; i < foo.len (); i++) { m[i] = foo[i]; }
  ::str s2 = m;
  assert (s2 == s);
  assert (foo == s2);

  ::str s3 = foo.copy ();
  assert (s3 == s2);
  assert (s3 == s);
  assert (foo == s3);
}


int
main (int argc, char *argv[])
{
  sp::gc::std_cfg_t cfg;
  cfg._n_b_arenae = 2;
  cfg._size_b_arenae = 1;
  cfg._smallobj_lim = 0;
  sp::gc::mgr_t<>::set (New sp::gc::std_mgr_t<> (cfg));

  test0 ();

  for (int i = 0; i < 10; i++) {
    sp::gc::mgr_t<>::get ()->sanity_check();
    test1();
    sp::gc::mgr_t<>::get ()->sanity_check();
    test2();
    sp::gc::mgr_t<>::get ()->sanity_check();
  }
  
  if (0) test3();
  test4();
}

