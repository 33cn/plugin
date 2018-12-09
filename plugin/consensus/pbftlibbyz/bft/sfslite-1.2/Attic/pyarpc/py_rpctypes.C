
#include "py_rpctypes.h"
#include "py_gen.h"


#define INT_RPC_TRAVERSE(T)                                       \
bool                                                              \
rpc_traverse (XDR *xdrs, pyw_##T &obj)                            \
{                                                                 \
  T tmp;                                                          \
  switch (xdrs->x_op) {                                           \
  case XDR_ENCODE:                                                \
    if (!obj.get (&tmp)) return false;                            \
  default:                                                        \
    break;                                                        \
  }                                                               \
  bool rc = rpc_traverse (xdrs, tmp);                             \
  if (rc) obj.set (tmp);                                          \
  return rc;                                                      \
}

#define ALLOC_DEFN(T)                                           \
void *                                                          \
T##_alloc ()                                                    \
{                                                               \
   return New T;                                                \
}

#define XDR_DEFN(T)						\
BOOL								\
xdr_##T (XDR *xdrs, void *objp)				        \
{								\
  return xdr_doit<T> (xdrs, objp);                              \
}								

#define INT_DO_ALL_C(T, s, P)                                   \
ALLOC_DEFN (pyw_##T);                                           \
XDR_DEFN (pyw_##T);                                             \
PY_RPC_PRINT_GEN (pyw_##T, sb.fmt ("0x%" s , obj.get ()));      \
RPC_PRINT_DEFINE (pyw_##T);                                     \
INT_RPC_TRAVERSE(T);

INT_DO_ALL_C (u_int32_t, "x", UnsignedLong);
INT_DO_ALL_C (int32_t, "d", Long);

#if SIZEOF_LONG != 8
INT_DO_ALL_C (u_int64_t, "qx", UnsignedLong );
INT_DO_ALL_C (int64_t, "qd", Long);
# else /* SIZEOF_LONG == 8 */
INT_DO_ALL_C (u_int64_t, "lx", UnsignedLong);
INT_DO_ALL_C (int64_t, "ld", Long );
#endif /* SIZEOF_LONG == 8 */

//
// Allocat pyw_void type
//
ALLOC_DEFN(pyw_void);
XDR_DEFN(pyw_void);
RPC_PRINT_GEN (pyw_void, sb.fmt ("<void>"));
RPC_PRINT_DEFINE (pyw_void);

bool
rpc_traverse (XDR *xdrs, pyw_void &v)
{
  return true;
}


pyw_base_t *
wrap_error (PyObject *in, PyObject *rpc_exception)
{
  PyErr_SetString (rpc_exception, "undefined RPC procedure called");
  return NULL;
}


py_rpcgen_table_t py_rpcgen_error = 
{
  wrap_error, wrap_error
};


bool
pyw_base_t::clear ()
{
  if (_obj) {
    Py_XDECREF (_obj);
    _obj = NULL;
  }
  return true;
}

bool
pyw_base_t::set_obj (PyObject *o)
{
  PyObject *tmp = _obj;
  Py_INCREF (o);
  _obj = o;
  Py_XDECREF (tmp);
  return true;
}

PyObject *
pyw_base_t::safe_get_obj (bool fail)
{
  PyObject *o = _obj;
  if (!o || fail) 
    o = Py_None;
  Py_INCREF (o);
  return o;
}

PyObject *
pyw_base_t::get_obj ()
{
  Py_XINCREF (_obj);
  return _obj;
}


PyObject *
pyw_base_t::unwrap ()
{
  PyObject *ret = get_obj ();
  delete this;
  return ret;
}

bool
pyw_base_t::init ()
{
  _obj = NULL;
  _typ = NULL;
  return true;
}


bool
pyw_base_t::alloc ()
{
  if (_typ) {
    _obj = _typ->tp_new (_typ, NULL, NULL);
    if (!_obj) {
      PyErr_SetString (PyExc_MemoryError, "out of memory in allocator");
      return false;
    }
  } else {
    _obj = Py_None;
    Py_INCREF (Py_None);
  }
  return true;
}


bool
pyw_base_err_t::print_err (const strbuf &b, int recdepth, 
			   const char *name, const str &prfx) const
{
  str s;
  switch (_err) {
  case PYW_ERR_NONE:
    return false;
  case PYW_ERR_TYPE:
    s = "TYPE";
    break;
  case PYW_ERR_BOUNDS:
    s = "BOUNDS";
    break;
  default:
    s = "unspeficied";
    break;
  }
  b << "<" << s << "-ERROR>;\n";
  return true;

}

PyObject *
py_rpc_ptr_t::alloc ()
{
  if (!p && !(p = typ->tp_new (typ, NULL, NULL))) 
    return PyErr_NoMemory ();
  return p;
}

void
py_rpc_ptr_t::clear ()
{
  PyObject *tmp = p;
  p = NULL;
  Py_XDECREF (tmp);
}

void
py_rpc_ptr_t::set_typ (PyTypeObject *o)
{
  if (typ) {
    assert (o == typ);
  } else {
    Py_INCREF (reinterpret_cast<PyObject *> (o));
    typ = o;
  }
}


//----------------------------------------------------------------------- 
// RPC PTR implementation
//

static void
py_rpc_ptr_t_dealloc (py_rpc_ptr_t *self)
{
  Py_XDECREF (self->typ); // need XDECREF in case of initialization failure
  Py_XDECREF (self->p);
  self->ob_type->tp_free (reinterpret_cast<PyObject *> (self));
}

PY_CLASS_NEW(py_rpc_ptr_t);

static bool
py_rpc_ptr_t_set_internal (py_rpc_ptr_t *self, PyObject *o)
{
  if (!PyObject_IsInstance (o, (PyObject *)self->typ)) {
    PyErr_SetString (PyExc_TypeError, "incorrect type for RPC ptr");
    return false;
  }

  PyObject *old = self->p;
  self->p = o;
  Py_XINCREF (o);
  Py_XDECREF (old);
  
  return true;
}

static int
py_rpc_ptr_t_init (py_rpc_ptr_t *self, PyObject *args, PyObject *kwds)
{
  PyObject *t = NULL;
  PyObject *o = NULL;
  static char *kwlist[] = { "typ", "obj", NULL };
  if (!PyArg_ParseTupleAndKeywords (args, kwds, "O|O", kwlist, &t, &o))
    return -1;

  if (!PyType_Check (t)) {
    PyErr_SetString (PyExc_TypeError, "expected a python type object");
    return -1;
  }
  self->typ = reinterpret_cast<PyTypeObject *> (t);
  Py_INCREF (self->typ);

  if (!py_rpc_ptr_t_set_internal (self, o))
    return -1;
  return 0;
}

static PyObject *
py_rpc_ptr_t_set (py_rpc_ptr_t *self, PyObject *args)
{
  PyObject *o = NULL;
  if (!PyArg_ParseTuple (args, "O",  &o))
    return NULL;

  if (!py_rpc_ptr_t_set_internal (self, o))
    return NULL;
  Py_INCREF (Py_None);
  return Py_None;
}

static PyObject *
py_rpc_ptr_t_get (py_rpc_ptr_t *self)
{
  PyObject *o = self->p ? self->p : Py_None;
  Py_INCREF (o);
  return o;
}

static PyObject *
py_rpc_ptr_t_alloc (py_rpc_ptr_t *self)
{
  if (!self->alloc ())
    return NULL;
  
  Py_INCREF (Py_None);
  return Py_None;
}

static PyObject *
py_rpc_ptr_t_clear (py_rpc_ptr_t *self)
{
  self->clear ();
  Py_INCREF (Py_None);
  return Py_None;
}

static PyMethodDef py_rpc_ptr_t_methods[] = {
  { "get", (PyCFunction)py_rpc_ptr_t_get,  METH_NOARGS,
    PyDoc_STR ("get the objet that an RPC pointer points to") },
  { "alloc", (PyCFunction)py_rpc_ptr_t_alloc, METH_NOARGS,
    PyDoc_STR ("allocate the object that the pointer points to") },
  { "set", (PyCFunction)py_rpc_ptr_t_set, METH_VARARGS,
    PyDoc_STR ("set the object that the pointer points to") },
  { "clear", (PyCFunction)py_rpc_ptr_t_clear, METH_NOARGS,
    PyDoc_STR ("clear out the pointer") },
 {NULL}
};

static PyObject *
py_rpc_ptr_t_getp (py_rpc_ptr_t *self, void *closure)
{
  return py_rpc_ptr_t_get (self);
}

static int
py_rpc_ptr_t_setp (py_rpc_ptr_t *self, PyObject *value, void *closure)
{
  return py_rpc_ptr_t_set_internal (self, value) ? 0 : -1;
}

static PyGetSetDef py_rpc_ptr_t_getsetters[] = {
  {"p", (getter)py_rpc_ptr_t_getp, (setter)py_rpc_ptr_t_setp,
   "magic pointer shortcut", NULL },
  {NULL}
};


PY_CLASS_DEF3(py_rpc_ptr_t, "async.rpc_ptr", 1, dealloc, -1, 
	      "RPC ptr object", methods, 0, getsetters, init, new, 0, 
	      0,0,0);

//
//-----------------------------------------------------------------------



//-----------------------------------------------------------------------
// 
// rpc_byte temp var
//

const strbuf &
rpc_print (const strbuf &sb, const pyw_rpc_byte_t &obj,
	   int recdepth, const char *name, const char *prefix)
{
  if (obj.print_err (sb, recdepth, name, prefix)) return sb;
  return rpc_print (sb, obj.get_byte (), recdepth, name, prefix);
}

//
//-----------------------------------------------------------------------
