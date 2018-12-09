
// -*-c++-*-

#ifndef __PY_VEC_H_INCLUDED__
#define __PY_VEC_H_INCLUDED__

#include <Python.h>
#include "structmember.h"
#include "arpc.h"
#include "rpctypes.h"

extern PyTypeObject py_rpc_vec_generic_type;

template<class T, size_t max>
struct py_rpc_vec {
  PyListObject list;

  static void init_rpc_vec_type (PyTypeObject *base, PySequenceMethods *s)
  {
    *base = rpc_vec_generic_type;
    base->tp_basicsize = sizeof (py_rpc_vec<T,max>);
    base->tp_name = py_vec_tp_name<T> ();

    memset (s, 0, sizeof (PySequenceMethods));
    s->list_ass_item ();
  }
    
};



#endif /* PY_VEC_H_INCLUDED */
