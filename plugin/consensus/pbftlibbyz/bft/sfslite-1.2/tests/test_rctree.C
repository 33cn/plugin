
#include "async.h"
#include "rctree.h"


struct dat_t {
  dat_t (int d) : _d (d) {}
  static ptr<dat_t> alloc (int i) { return New refcounted<dat_t> (i); }
  const int _d;
  void dump () { warn << "dump: " << _d << "\n"; }
  int to_int() const { return _d; }
  rctree_entry_t<dat_t> _lnk;
};

template<class T> void insert (T *t, int v) { t->insert (v, dat_t::alloc (v)); }


void
tree_test (void)
{
  rctree_t<int, dat_t, &dat_t::_lnk> tree;

  int vals[] = { 1, 1, 2, 3, 4, 5, 10, 12, 20, 120, 200, 300, 12, 12, 12, -1 };

  for (int *vp = vals; *vp >= 0; vp++ ){
    insert (&tree, *vp);
  }
  
  for (int *vp = vals; *vp >= 0; vp++) {
    assert (tree[*vp]->to_int () == *vp);
  }
 
}

int
main (int argc, char *argv)
{
  tree_test ();
  


}
