
#include "async.h"

class my_resizer_t : public vec_resizer_t {
public:
  my_resizer_t () : vec_resizer_t () {}
  size_t resize (u_int nalloc, u_int nwanted, int objid);
};

size_t
my_resizer_t::resize (u_int nalloc, u_int nwanted, int objid)
{
  int exponent = fls (max (nalloc, nwanted));

  int step;

  if (exponent < 3) step = 1;
  else if (exponent < 8) step = 3;
  else if (exponent < 10) step = 2;
  else step = 1;

  exponent = ((exponent - 1) / step + 1) * step;
  size_t ret = 1 << exponent;

  // If you want to know the pattern...
  warn << "resize: " << nalloc << "," << nwanted << "," << objid 
       << " -> " << ret << "\n";

  return ret;
}

template<>
struct vec_obj_id_t<int>
{
  vec_obj_id_t (){}
  int operator() (void) const { return 1; }
};

static void
vec_test (vec<int> &v, int n)
{
  for (int i = 0; i < n; i++) {
    v.push_back (i);
  }
  for (int i = n - 1; i >= 0; i--) {
    assert (v.pop_back () == i);
  }
}

static void
vec_test (void)
{
  vec<int> v1, v2;
  int n = 100;

  vec_test (v1, n);
  set_vec_resizer (New my_resizer_t ());
  vec_test (v2, n);
}


int
main (int argc, char *argv[])
{
  vec_test ();
  return 0;
}
