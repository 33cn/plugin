
#include "dynenum.h"

bool
dynamic_enum_t::lookup (const str &s, int *vp) const
{
  bool ret = false;
  int v = _def_val;
  const int *i = s ? _tab[s] : NULL;
  if (i) {
    ret = true;
    v = *i;
  }
  if (vp) *vp = v;
  return ret;
}

int
dynamic_enum_t::lookup (const str &s, bool dowarn) const
{
  int ret = 0; // silence not-initialized warnings
  bool ok = lookup (s, &ret);
  if (!ok && dowarn)
    warn_not_found (s);
  return ret;
}

void
dynamic_enum_t::warn_not_found (str s) const
{
  if (!s)
    s = "(null)";

  str n = _enum_name;
  if (!n)
    n = "anonymous";

  warn << "XX dynamic enum (" << n << "): no value for key=" << s << "\n";
}


void
dynamic_enum_t::init (const pair_t pairs[], bool chk)
{
  for (const pair_t *p = pairs; p->n; p++) {
    _tab.insert (p->n, p->v);
  }

  if (chk) {
    // Do a spot check!
    for (const pair_t *p = pairs; p->n; p++) {
      assert ((*this)[p->n] == p->v);   
    }
  }
}
