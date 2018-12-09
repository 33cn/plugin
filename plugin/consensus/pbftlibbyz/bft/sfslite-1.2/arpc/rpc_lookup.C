
#include "qhash.h"
#include "arpc.h"

static qhash<const char *, qhash<const char *, u_int32_t> > rpc_lookup_tab;

bool
rpc_program::lookup (const char *rpc, u_int32_t *out) const
{
  qhash<const char *, u_int32_t> *t = rpc_lookup_tab[name];
  if (!t) {
    rpc_lookup_tab.insert (rpc);
    t = rpc_lookup_tab[name];
    assert (t);
    for (size_t i = 0; i < nproc; i++) {
      if (tbl[i].name) {
	t->insert (tbl[i].name, i);
      }
    }
  }
  u_int32_t *val = (*t)[rpc];
  if (val) {
    *out = *val;
    return true;
  }
  return false;
}
