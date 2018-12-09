// -*-c++-*-

#include <Python.h>
#include "structmember.h"
#include "async.h"
#include "arpc.h"
#include "py_rpctypes.h"
#include "py_gen.h"
#include "py_util.h"

static PyObject *AsyncXDR_Exception;
static PyObject *AsyncRPC_Exception;


struct py_axprt_t {
  PyObject_HEAD
  ptr<axprt> x;
  PyObject *sock;
};

struct py_axprt_stream_t : public py_axprt_t {};

struct py_axprt_dgram_t : public py_axprt_t {};

struct py_aclnt_t {
  PyObject_HEAD
  py_rpc_program_t *py_prog;
  PyObject *x;
  ptr<aclnt> cli;
};

struct py_asrv_t {
  PyObject_HEAD
  PyObject *x;
  py_rpc_program_t *py_prog;
  ptr<asrv> srv;
};

struct py_svccb_t {
  PyObject_HEAD
  svccb *sbp;
  wrap_fn_t wrap_res;
};

PY_ABSTRACT_CLASS(py_rpc_program_t, "arpc.prc_program");
PY_ABSTRACT_CLASS(py_axprt_t, "arpc.axprt");


//-----------------------------------------------------------------------
//
//  axprt_stream
//

PY_CLASS_NEW(py_axprt_stream_t);

static int
py_axprt_stream_t_init (py_axprt_stream_t *self, PyObject *args, 
			PyObject *kwds)
{
  static char *kwlist[] = { "fd", "defps", "sock", NULL };
  int fd;
  int defps = 0;
  PyObject *sock = NULL;

  if (! PyArg_ParseTupleAndKeywords (args, kwds, "i|Oi", kwlist,
				     &fd, &sock, &defps))
    return -1;

  if (defps) {
    self->x = axprt_stream::alloc (fd, defps);
  } else {
    self->x = axprt_stream::alloc (fd);
  }

  Py_XINCREF (sock);
  self->sock = sock;

  return 0;
}

static void
py_axprt_stream_t_dealloc (py_axprt_stream_t *self)
{
  warnx << "~stream\n";
  self->x = NULL;
  Py_XDECREF (self->sock);
  self->ob_type->tp_free ((PyObject *)self);
}

PY_CLASS_DEF(py_axprt_stream_t, "arpc.axprt_stream", 1, dealloc, -1,
	     "py_axprt_stream_t object", 0, 0, 0, init, new, 
	     &py_axprt_t_Type);
//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
//
//  svccb
//

static void
py_svccb_t_dealloc (py_svccb_t *self)
{
  if (self->sbp) {
    self->sbp->reject (PROC_UNAVAIL);
    self->sbp = NULL;
  }
  warn << "~svccb\n";
  self->ob_type->tp_free ((PyObject *)self);
}

PY_CLASS_NEW (py_svccb_t);

static PyObject *
py_svccb_t_eof (py_svccb_t *o)
{
  PyObject *ret;
  ret = o->sbp ? Py_False : Py_True;
  Py_INCREF (ret);
  return ret;
}

static bool
sbp_check (py_svccb_t *o)
{
  if (!o->sbp) {
    PyErr_SetString (AsyncRPC_Exception, "no svccb active");
    return false;
  }
  return true;
}

static PyObject *
py_svccb_t_proc (py_svccb_t *o)
{
  long l = 0;
  if (o->sbp)
    l = long (o->sbp->proc ());
  return PyInt_FromLong (l);
}

static PyObject *
py_svccb_t_getarg (py_svccb_t *o)
{
  if (!sbp_check (o)) return NULL;
  pyw_base_t *arg = o->sbp->Xtmpl getarg<pyw_base_t> ();
  if (!arg) {
    // XXX maybe return Py_None?
    PyErr_SetString (AsyncRPC_Exception, "no argument given");
    return NULL;
  }

  // keeps the arg around in case we need to call this guy again
  // Obviously, get_obj () increases the reference count, so
  // we're returning a new reference
  return arg->get_obj ();
}

static PyObject *
py_svccb_t_getres (py_svccb_t *o)
{
  if (!sbp_check (o)) return NULL;
  pyw_base_t *res = o->sbp->Xtmpl getres<pyw_base_t> ();
  return res->unwrap ();
}

static PyObject *
py_svccb_t_reject (py_svccb_t *o, PyObject *args)
{
  warn << "in reject\n";
  if (!sbp_check (o)) return NULL;
  int stat = 0;
  char auth = 0;
  PyObject *ret = NULL;

  if (!PyArg_ParseTuple (args, "i|b", &stat, &auth))
    goto done;

  if (auth) {
    o->sbp->reject (auth_stat (stat));
  } else {
    o->sbp->reject (accept_stat (stat));
  }
  o->sbp = NULL;

  Py_INCREF (Py_None);
  ret = Py_None;

 done:
  return ret;
}

static PyObject *
py_svccb_t_ignore (py_svccb_t *o)
{
  if (!sbp_check (o)) return NULL;

  o->sbp->ignore ();
  o->sbp = NULL;

  Py_INCREF (Py_None);
  return Py_None;
}

static PyObject *
py_svccb_t_reply (py_svccb_t *o, PyObject *args)
{
  if (!sbp_check (o)) return NULL;
  PyObject *obj = NULL;
  if (!PyArg_ParseTuple (args, "O", &obj))
    return NULL;

  pyw_base_t *res = o->wrap_res (obj, AsyncRPC_Exception);
  if (!res)
    return NULL;

  o->sbp->reply (res);
  o->sbp = NULL;

  delete res; // DM's classes don't do this for us!

  Py_INCREF (Py_None);
  return Py_None;
}

static PyMethodDef py_svccb_t_methods[] = {
  { "reply", (PyCFunction)py_svccb_t_reply, METH_VARARGS,
    PyDoc_STR ("svccb::reply from libasync") },
  { "ignore", (PyCFunction)py_svccb_t_ignore, METH_NOARGS,
    PyDoc_STR ("svccb::ignore from libasync") },
  { "eof", (PyCFunction)py_svccb_t_eof, METH_NOARGS,
    PyDoc_STR ("check if server is at EOF") },
  { "proc", (PyCFunction)py_svccb_t_proc, METH_NOARGS,
    PyDoc_STR ("svccb::proc from libasync") },
  { "getarg", (PyCFunction)py_svccb_t_getarg, METH_NOARGS,
    PyDoc_STR ("svccb::getarg from libasync") },
  { "getres", (PyCFunction)py_svccb_t_getres, METH_NOARGS,
    PyDoc_STR ("svccb::getres from libasync") },
  { "reject", (PyCFunction)py_svccb_t_reject, METH_VARARGS,
    PyDoc_STR ("svccb::reject from libasync") },
  {NULL}
};

PY_CLASS_DEF (py_svccb_t, "arpc.svccb", 1, dealloc, -1,
	      "svccb object wrapper", methods, 0, 0, 0, new, 0);

//
//
//-----------------------------------------------------------------------


//-----------------------------------------------------------------------
//
//  asrv
//

static void
py_asrv_t_dealloc (py_asrv_t *self)
{
  warn << "~asrv\n";
  Py_DECREF (self->x);
  self->srv = NULL;
  self->ob_type->tp_free ((PyObject *)self);
}
  
PY_CLASS_NEW(py_asrv_t);

struct asrvbun_t {

  asrvbun_t (py_rpc_program_t *p, PyObject *c)
    : prog (p), cb (c) { Py_INCREF (prog); Py_XINCREF (cb); }

  ~asrvbun_t () { Py_DECREF (prog); Py_XDECREF (cb); }

  static ptr<asrvbun_t> alloc (py_rpc_program_t *p, PyObject *c) 
  { return New refcounted<asrvbun_t> (p, c); }

  py_rpc_program_t *prog;
  PyObject *cb;
};

static void
py_asrv_t_dispatch (ptr<asrvbun_t> bun, svccb *sbp)
{
  if (!bun->cb) {
    sbp->ignore ();
    return;
  }

  py_svccb_t *py_sbp =  PyObject_New (py_svccb_t, &py_svccb_t_Type);

  if (!py_sbp) {
    sbp->reject (PROC_UNAVAIL);
  } else {
    py_sbp->sbp = sbp;
    if (sbp)
      py_sbp->wrap_res = bun->prog->pytab[sbp->proc ()].wrap_res;
    PyObject *arglist = Py_BuildValue ("(O)", py_sbp);
    PyObject *pres = PyObject_CallObject (bun->cb, arglist);

    // will cause a stack trace and exit if there is an error
    py_throwup (pres);

    Py_XDECREF (pres);
    Py_DECREF (arglist);
    Py_DECREF (py_sbp);

  }
}


static PyObject *
py_asrv_t_setcb (py_asrv_t *self, PyObject *args, PyObject *kwds)
{
  PyObject *cb = NULL;
  static char *kwlist[] = { "cb", NULL };
  
  if (args && !PyArg_ParseTupleAndKeywords (args, kwds, "|O", kwlist, &cb))
    return NULL;

  if (cb && cb == Py_None)
    cb = NULL;

  if (!cb) {
    self->srv->setcb (NULL);
  } else {
    if (!assure_callable (cb)) return NULL;
    self->srv->setcb (wrap (py_asrv_t_dispatch,
			    asrvbun_t::alloc (self->py_prog, cb)));
  }

  Py_INCREF (Py_None);
  return Py_None;
}

static PyObject *
py_asrv_t_clearcb (py_asrv_t *self)
{
  return py_asrv_t_setcb (self, NULL, NULL);
}

static PyObject *
py_asrv_t_stop (py_asrv_t *self)
{
  self->srv->stop();
  Py_INCREF (Py_None);
  return Py_None;
}

static PyObject *
py_asrv_t_start (py_asrv_t *self)
{
  self->srv->start ();
  Py_INCREF (Py_None);
  return Py_None;
}


static int
py_asrv_t_init (py_asrv_t *self, PyObject *args, PyObject *kwds)
{
  PyObject *x = NULL;
  PyObject *prog = NULL;
  PyObject *cb = NULL;
  PyObject *tmp;
  static char *kwlist[] = { "x", "prog", "cb", NULL };

  self->py_prog = NULL;
  if (!PyArg_ParseTupleAndKeywords (args, kwds, "OOO", kwlist, 
				    &x, &prog, &cb))
    return -1;

  if (!x || !PyObject_IsInstance (x, (PyObject *)&py_axprt_t_Type)) {
    PyErr_SetString (PyExc_TypeError,
		     "asrv expects arg 1 as an axprt transport");
    return -1;
  }

  self->x = x;
  Py_INCREF (x);

  if (!prog || 
      !PyObject_IsInstance (prog, (PyObject *)&py_rpc_program_t_Type)) {
    PyErr_SetString (PyExc_TypeError,
		     "asrv expects arg 2 as an rpc_program");
    return -1;
  }

  if (cb && !PyCallable_Check (cb)) {
    PyErr_SetString (PyExc_TypeError,
		     "asrv expects arg 3 as a callback");
    return -1;
  }
  py_axprt_t *p_x = (py_axprt_t *)x;
  self->py_prog = reinterpret_cast<py_rpc_program_t *> (prog);
  Py_INCREF (self->py_prog);

  self->srv = asrv::alloc (p_x->x, *self->py_prog->prog,
			   wrap (py_asrv_t_dispatch, 
				 asrvbun_t::alloc (self->py_prog, cb)));
  return 0;
}

static PyMethodDef py_asrv_t_methods[] = {
  { "setcb", (PyCFunction)py_asrv_t_setcb, METH_VARARGS,
    PyDoc_STR ("asrv::setcb function from libasync") },
  { "stop", (PyCFunction)py_asrv_t_stop, METH_NOARGS,
    PyDoc_STR ("asrv::stop function from libasync") },
  { "start", (PyCFunction)py_asrv_t_start, METH_NOARGS,
    PyDoc_STR ("asrv::start function from libasync") },
  { "clearcb", (PyCFunction)py_asrv_t_clearcb, METH_NOARGS,
    PyDoc_STR ("asrv::setcb (NULL) from libasync") },
  {NULL}
};


PY_CLASS_DEF(py_asrv_t, "arpc.asrv", 1, dealloc, -1,
	     "asrv object wrapper", methods, 0, 0, init, new, 0);

//
//-----------------------------------------------------------------------


//-----------------------------------------------------------------------
//
//  aclnt
//

PY_CLASS_NEW(py_aclnt_t);

static void
py_aclnt_t_dealloc (py_aclnt_t *self)
{
  Py_DECREF (self->x);
  self->cli = NULL;
  self->ob_type->tp_free ((PyObject *)self);
}


static int
py_aclnt_t_init (py_aclnt_t *self, PyObject *args, PyObject *kwds)
{
  PyObject *x = NULL;
  PyObject *prog = NULL;
  PyObject *tmp;
  static char *kwlist[] = { "x", "prog", NULL };
  
  self->py_prog = NULL;

  if (!PyArg_ParseTupleAndKeywords (args, kwds, "OO", kwlist, &x, &prog))
    return -1;

  if (!x || !PyObject_IsInstance (x, (PyObject *)&py_axprt_t_Type)) {
    PyErr_SetString (PyExc_TypeError,
		     "aclnt expects arg 1 as an axprt transport");
    return -1;
  }

  self->x = x;
  Py_INCREF (x);

  if (!prog || 
      !PyObject_IsInstance (prog, (PyObject *)&py_rpc_program_t_Type)) {
    PyErr_SetString (PyExc_TypeError,
		     "aclnt expects arg 2 as an rpc_program");
    return -1;
  }
  py_axprt_t *p_x = reinterpret_cast<py_axprt_t *> (x);
  self->py_prog = reinterpret_cast<py_rpc_program_t *> (prog);
  Py_INCREF (self->py_prog);

  self->cli = aclnt::alloc (p_x->x, *self->py_prog->prog);
  
  return 0;
}

struct aclntbun_t {

  aclntbun_t (pyw_base_t *a, pyw_base_t *r, PyObject *c)
    : arg (a), res (r), cb (c) 
  {
    Py_XINCREF (c);
  }

  ~aclntbun_t ()
  {
    Py_XDECREF (cb);
    if (arg) delete arg;
    if (res) delete res;
  }

  pyw_base_t *arg;
  pyw_base_t *res;
  PyObject *cb;
};  

static void
py_aclnt_t_call_cb (ptr<aclntbun_t> bun, clnt_stat e)
{
  if (bun->cb) {
    PyObject *res = bun->res->safe_get_obj ( e != 0 );
    PyObject *arglist = Py_BuildValue ("(iO)", (int )e, res);
    PyObject *pres = PyObject_CallObject (bun->cb, arglist);

    // will cause a stack trace and exit if error
    py_throwup (pres);

    Py_XDECREF (pres);
    Py_DECREF (arglist);
    Py_DECREF (res);  // Py_BuildValue raises the refcount on res

  } 
}

static PyObject *
py_aclnt_t_call (py_aclnt_t *self, PyObject *args)
{
  int procno = 0;
  PyObject *rpc_args_in = NULL;
  pyw_base_t *rpc_args_out = NULL;
  PyObject *cb = NULL ;
  pyw_base_t *res = NULL;
  callbase *b = NULL;
  const py_rpcgen_table_t *pyt;
  const rpcgen_table *tbl;
  ptr<aclntbun_t> bundle;

  if (!PyArg_ParseTuple (args, "i|OO", &procno, &rpc_args_in, &cb))
    return NULL;

  // now check all arguments
  if (procno >= self->py_prog->prog->nproc) {
    PyErr_SetString (AsyncRPC_Exception, "given RPC proc is out of range");
    return NULL;
  }

  pyt = &self->py_prog->pytab[procno];
  tbl = &self->py_prog->prog->tbl[procno];

  if (cb && !PyCallable_Check (cb)) {
    PyErr_SetString (PyExc_TypeError, "expected a callable object as 3rd arg");
    goto fail;
  }

  if (!rpc_args_in) {
    if (!(rpc_args_out = static_cast<pyw_base_t *> (tbl->alloc_arg ()))) {
      PyErr_SetString (PyExc_MemoryError, "out of memory in alloc_arg");
      goto fail;
    }
  } else {
    if (!(rpc_args_out = pyt->wrap_arg (rpc_args_in, AsyncRPC_Exception)))
      goto fail;
  }

  if (!(res = static_cast<pyw_base_t *> (tbl->alloc_res ()))) {
    PyErr_SetString (PyExc_MemoryError, "out of memory in alloc_res");
    goto fail;
  }

  bundle = New refcounted<aclntbun_t> (rpc_args_out, res, cb);

  b = self->cli->call (procno, rpc_args_out, res, 
		       wrap (py_aclnt_t_call_cb, bundle));
  if (!b) 
    goto fail;

  // XXX eventually return the callbase to cancel a call
  Py_INCREF (Py_None);
  return Py_None;
 fail:
  if (rpc_args_out) delete rpc_args_out ;
  if (res) delete res;
  return NULL;
}

static PyMethodDef py_aclnt_t_methods[] = {
  { "call", (PyCFunction)py_aclnt_t_call, METH_VARARGS,
    PyDoc_STR ("aclnt::call function from libasync") },
  {NULL}
};

PY_CLASS_DEF(py_aclnt_t, "arpc.aclnt", 1, dealloc, -1, 
	     "aclnt object wrapper", methods, 0, 0, init, new, 0);

//
//-----------------------------------------------------------------------


static PyMethodDef module_methods[] = {
  { NULL }
};

static int
all_ins (PyObject *m)
{
  int rc;
  INS (RPC_SUCCESS);
  INS (RPC_CANTENCODEARGS);
  INS (RPC_CANTDECODERES);
  INS (RPC_CANTSEND);
  INS (RPC_CANTRECV);
  INS (RPC_TIMEDOUT);
  INS (RPC_VERSMISMATCH);
  INS (RPC_AUTHERROR);
  INS (RPC_PROGUNAVAIL);
  INS (RPC_PROGVERSMISMATCH);
  INS (RPC_PROCUNAVAIL);
  INS (RPC_CANTDECODEARGS);
  INS (RPC_SYSTEMERROR);
  INS (RPC_UNKNOWNHOST);
  INS (RPC_UNKNOWNPROTO);
  INS (RPC_PROGNOTREGISTERED);
  INS (RPC_FAILED);
  INS (SUCCESS);
  INS (PROG_UNAVAIL);
  INS (PROG_MISMATCH);
  INS (PROC_UNAVAIL);
  INS (GARBAGE_ARGS);
  INS (SYSTEM_ERR);
  INS (AUTH_OK);
  INS (AUTH_BADCRED);
  INS (AUTH_REJECTEDCRED);
  INS (AUTH_BADVERF);
  INS (AUTH_REJECTEDVERF);
  INS (AUTH_TOOWEAK);
  INS (AUTH_INVALIDRESP);
  INS (AUTH_FAILED);
  return 1;

  // XXX the below seem nonstandard
#if 0
  INS (RPC_NOBROADCAST);
  INS (RPC_UNKNOWNADDR);
  INS (RPC_RPCBFAILURE);
  INS (RPC_N2AXLATEFAILURE);
  INS (RPC_INTR);
  INS (RPC_TLIERROR);
  INS (RPC_UDERROR);
  INS (RPC_INPROGRESS);
  INS (RPC_STALERACHANDLE);
#endif


}

#ifndef PyMODINIT_FUNC	/* declarations for DLL import/export */
#define PyMODINIT_FUNC void
#endif
PyMODINIT_FUNC
initarpc (void)
{
  PyObject* m;
  if (PyType_Ready (&py_axprt_t_Type) < 0 ||
      PyType_Ready (&py_axprt_stream_t_Type) < 0 ||
      PyType_Ready (&py_rpc_program_t_Type) < 0 ||
      PyType_Ready (&py_aclnt_t_Type) < 0 ||
      PyType_Ready (&py_asrv_t_Type) < 0 ||
      PyType_Ready (&py_svccb_t_Type) < 0)
    return;

  if (!import_async_exceptions (&AsyncXDR_Exception, 
				&AsyncRPC_Exception,
				NULL))
    return;

  m = Py_InitModule3 ("async.arpc", module_methods,
                      "arpc wrappers for libasync");

  if (m == NULL)
    return;

  if (all_ins (m) < 0)
    return;

  Py_INCREF (&py_axprt_t_Type);
  Py_INCREF (&py_axprt_stream_t_Type);
  Py_INCREF (&py_rpc_program_t_Type);
  Py_INCREF (&py_aclnt_t_Type);
  PyModule_AddObject (m, "axprt", (PyObject *)&py_axprt_t_Type);
  PyModule_AddObject (m, "axprt_stream", (PyObject *)&py_axprt_stream_t_Type);
  PyModule_AddObject (m, "rpc_program", (PyObject *)&py_rpc_program_t_Type);
  PyModule_AddObject (m, "aclnt", (PyObject *)&py_aclnt_t_Type);
  PyModule_AddObject (m, "asrv", (PyObject *)&py_asrv_t_Type);
  PyModule_AddObject (m, "svccb", (PyObject *)&py_svccb_t_Type);
}

