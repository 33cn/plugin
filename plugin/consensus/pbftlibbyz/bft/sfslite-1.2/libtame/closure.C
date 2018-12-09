
#include "tame_closure.h"
#include "tame_rendezvous.h"


bool
closure_t::block_dec_count (const char *loc)
{
  bool ret = false;
  if (_block._count <= 0) {
    tame_error (loc, "too many triggers for wait environment.");
  } else if (--_block._count == 0) {
    ret = true;
  }
  return ret;
}

str
closure_t::loc (int l) const
{
  strbuf b;
  b << _filename << ":" << l << " in function " << _funcname;
  return b;
}

closure_t::closure_t (const char *file, const char *fun)
  : _jumpto (0), 
    _id (++closure_serial_number),
    _filename (file),
    _funcname (fun)
{
  g_stats->did_mkclosure ();
}

static void
report_rv_problems (const vec<weakref<rendezvous_base_t> > &rvs)
{
  for (u_int i = 0; i < rvs.size (); i++) {
    u_int n;
    const rendezvous_base_t *p = rvs[i].pointer ();
    if (p && (n = p->n_triggers_left ())) {
      strbuf b ("rendezvous still active with %u trigger%s after control "
		"left function", 
		n, 
		n > 1 ? "s" : "");
      str s = b;
      tame_error (p->loc (), s.cstr ());
    }
  }
}

static void
end_of_scope_checks (vec<weakref<rendezvous_base_t> > rvs)
{
  report_rv_problems (rvs);
}

void 
closure_t::end_of_scope_checks (int line)
{
  if (tame_check_leaks ()) {
    // Unfortunately, we can only perform these end of scope checks
    // from the event loop, since we need to wait for the callstack
    // to unwind.  Of course, we can't hold a reference to the
    // closure since that will keep it from going out of scope.
    // So instead, we hold onto the relevant pieces inside the class,
    // with an expensive copy in the case of the _rvs.
    delaycb (0, 0, wrap (::end_of_scope_checks, _rvs));
  }
}

void
closure_t::error (int lineno, const char *msg)
{
  str s = loc (lineno);
  tame_error (s.cstr(), msg);
}

void
closure_t::init_block (int blockid, int lineno)
{
  _block._id = blockid;
  _block._count = 1;
  _block._lineno = lineno;
}

