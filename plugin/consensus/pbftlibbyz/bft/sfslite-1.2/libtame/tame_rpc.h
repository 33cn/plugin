
#include "arpc.h"
#include "tame_event.h"

namespace tame_rpc {

  inline void
  call (ptr<aclnt> c, u_int32_t procno, const void *arg,
	void *res, event<clnt_stat>::ref ev)
  {
    aclnt_cb cb = ev;
    c->call (procno, arg, res, cb);
  }

  inline callbase *
  rcall (ptr<aclnt> c, u_int32_t procno, const void *arg,
	void *res, event<clnt_stat>::ref ev)
  {
    aclnt_cb cb = ev;
    return c->call (procno, arg, res, cb);
  }

};
