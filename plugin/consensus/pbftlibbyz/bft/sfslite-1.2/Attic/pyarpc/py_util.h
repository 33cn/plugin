
// -*-c++-*-

#ifndef __PY_UTIL_H_INCLUDED__
#define __PY_UTIL_H_INCLUDED__

#include <Python.h>
#include "async.h"

template<class T> str
safe_to_str (const T *in)
{
  strbuf b;
  rpc_print (b, *in);
  PyErr_Clear ();
  return b;
}


template<class T> PyObject *
pretty_print_to_py_str_obj (const T *in)
{
  str s = safe_to_str (in);
  return PyString_FromFormat (s.cstr ());
}

template<class T> int
pretty_print_to_fd (const T *in, FILE *fp)
{
  str s = safe_to_str (in);
  if (s[s.len () - 1] == '\n') {
    mstr m (s.len ());
    strcpy (m.cstr (), s.cstr ());
    m[m.len () - 1] = '\0';
    m.setlen (m.len () - 1);
    s = m;
  }
    
  fputs (s.cstr (), fp);
  return 0;
}

template<class T>
struct pp_t {
  pp_t (T *o) : _p (o) { Py_INCREF (pyobj ()); }
  ~pp_t () { Py_DECREF (pyobj ()); }
  PyObject *pyobj () { return reinterpret_cast<PyObject *> (_p); }
  T * obj () { return _p; }
  static ptr<pp_t<T> > alloc (T *t) 
  { return t ? New refcounted<pp_t<T> > (t) : NULL ; }
private:
  T *const _p;
};

typedef pp_t<PyObject> pop_t;

bool assure_callable (PyObject *obj);
void py_throwup (PyObject *obj) ;

#endif
