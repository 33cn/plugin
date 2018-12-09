
#include <Python.h>
#include <stdlib.h>
#include "async.h"
#include "py_util.h"
#include "py_gen.h"
#include "py_debug.h"

//
// $Id: core.C 867 2005-05-16 19:37:57Z max $
//

static PyObject *
Core_amain (PyObject *self, PyObject *args)
{
  // of course, does not return
  amain ();

  Py_INCREF (Py_None);
  return Py_None;
}

static void
Core_cb (ptr<pop_t> handler)
{
  PyObject *pres = PyEval_CallObject (handler->obj (), NULL);
  Py_XDECREF (pres);
}

static PyObject *
Core_nowcb (PyObject *self, PyObject *args)
{
  PyObject *cb;
  if (!PyArg_ParseTuple (args, "O", &cb))
    return NULL;

  if (!PyCallable_Check (cb)) {
    PyErr_SetString (PyExc_TypeError, "expected a callable object");
    return NULL;
  }
  delaycb (0, 0, wrap (Core_cb, pop_t::alloc (cb)));
  
  Py_INCREF (Py_None);
  return Py_None;
}


static PyObject *
Core_delaycb (PyObject *self, PyObject *args)
{
  PyObject *handler;
  int s, us;

  if (!PyArg_ParseTuple (args, "iiO", &s, &us, &handler))
    return NULL;

  if (!PyCallable_Check (handler)) {
    PyErr_SetString (PyExc_TypeError, "expected a callable object as 3rd arg");
    return NULL;
  }

  delaycb  (s, us, wrap (Core_cb, pop_t::alloc (handler)));

  Py_INCREF (Py_None);
  return Py_None;
}

static PyObject *
Core_make_async (PyObject *self, PyObject *args)
{
  int fd = 0;

  if (!PyArg_ParseTuple (args, "i", &fd))
    return NULL;

  make_async (fd);

  Py_INCREF (Py_None);
  return Py_None;
}

static PyObject *
Core_fdcb (PyObject *self, PyObject *args)
{
  int fd = 0;
  int i_op = 0;
  PyObject *cb = NULL;

  if (!PyArg_ParseTuple (args, "ii|O", &fd, &i_op, &cb))
    return NULL;

  selop op = selop (i_op);

  if (op != selread && op != selwrite) {
    PyErr_SetString (PyExc_TypeError, "unknown select option");
    return NULL;
  }

  if (!cb) {
    fdcb (fd, op, NULL);
  } else {
    if (!PyCallable_Check (cb)) {
      PyErr_SetString (PyExc_TypeError, "callable object expected");
      return NULL;
    }
    ptr<pop_t> pop = pop_t::alloc (cb);
    make_async (fd);
    fdcb (fd, op, wrap (Core_cb, pop));
  }

  Py_INCREF (Py_None);
  return Py_None;
}

static PyObject *
Core_sigcb (PyObject *self, PyObject *args)
{
  int sig = 0;
  int flags = 0;
  PyObject *cb = NULL;

  if (!PyArg_ParseTuple (args, "i|Oi", &sig, &cb, &flags))
    return NULL;

  if (!cb) {
    sigcb (sig, NULL, flags);
  } else {
    if (!PyCallable_Check (cb)) {
      PyErr_SetString (PyExc_TypeError, "callable object expected");
      return NULL;
    }
    sigcb (sig, wrap (Core_cb, pop_t::alloc (cb)), flags);
  }

  Py_INCREF (Py_None);
  return Py_None;
}

static PyObject *
Core_exit (PyObject *self, PyObject *args)
{
  int i = 0;
  if (!PyArg_ParseTuple (args, "|i", &i))
    return NULL;

  // report debug information to find memory leaks in pysfs
  PYDEBUG_MEMREPORT(warnx);

  exit (i);
  return NULL;
}

static PyObject *
Core_setprogname (PyObject *self, PyObject *args)
{
  char *s;
  if (!PyArg_ParseTuple (args, "s", &s))
    return NULL;
  if (s)
    setprogname (s);
  Py_INCREF (Py_None);
  return Py_None;
}

static struct PyMethodDef core_methods[] = {
  { "amain",  Core_amain , METH_NOARGS,
    PyDoc_STR ("amain () from libasync") },
  { "delaycb", Core_delaycb, METH_VARARGS,
    PyDoc_STR ("delaycb from libasync") },
  { "nowcb", Core_nowcb, METH_VARARGS,
    PyDoc_STR ("delaycb (0,0,cb) from libasync") },
  { "fdcb", Core_fdcb, METH_VARARGS,
    PyDoc_STR ("fdcb from libasync") },
  { "sigcb", Core_sigcb, METH_VARARGS,
    PyDoc_STR ("sigcb from libasync") },
  { "make_async", Core_make_async, METH_VARARGS,
    PyDoc_STR ("make_async from libasync") },
  { "exit", Core_exit, METH_VARARGS,
    PyDoc_STR ("async exit function") },
  { "setprogname" , Core_setprogname, METH_VARARGS,
    PyDoc_STR ("setprogname from libasync") },
  { NULL, NULL, NULL, NULL }
};

static int
all_ins (PyObject *m)
{
  int rc;
  INS (selread);
  INS (selwrite);
  return 1;
}

extern "C" {
void initcore()
{
  PyObject *m;
  m = Py_InitModule3 ("async.core", core_methods,
		      "wrappers for SFS core event loop");
  if (m == NULL)
    return;

  if (all_ins (m) < 0)
    return;
    
  (void) Py_InitModule ("core", core_methods);
}

}
