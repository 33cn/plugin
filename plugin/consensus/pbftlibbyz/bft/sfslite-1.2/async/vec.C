
#include "vec.h"

static vec_resizer_t *vec_resizer;

void
set_vec_resizer (vec_resizer_t *v)
{
  if (vec_resizer)
    delete vec_resizer;
  vec_resizer = v;
}

size_t
vec_resize_fn (u_int nalloc, u_int nwanted, int objid)
{
  size_t ret; 
  if (vec_resizer) {
    ret = vec_resizer->resize (nalloc, nwanted, objid);
  } else {
    // The original resize operation from SFS
    ret = (1 << fls (max (nalloc, nwanted)));
  }
  return ret;
}
