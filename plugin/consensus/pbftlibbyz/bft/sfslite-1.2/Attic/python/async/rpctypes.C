
#include "py_rpctypes.h"
#include "py_gen.h"


//
// note that the symbol type for py_rpc_ptr_t_Type is 
// defined externally, in libpyarpc.so
///

static PyMethodDef module_methods[] = {
  {NULL}
};


PyMODINIT_FUNC
initrpctypes (void)
{
  if (PyType_Ready (&py_rpc_ptr_t_Type) < 0)
    return;
  PyObject *m;
  m = Py_InitModule3 ("async.rpctypes", module_methods,
		      "RPC types for Python XDR");
  if ( m == NULL )
    return;

  Py_INCREF (&py_rpc_ptr_t_Type);
  PyModule_AddObject (m, "rpc_ptr", (PyObject *)&py_rpc_ptr_t_Type);
}
