
#include "py_util.h"


bool
assure_callable (PyObject *obj)
{
  if (! PyCallable_Check (obj)) {
    PyErr_SetString (PyExc_TypeError, "expected callable object");
    return false;
  }
  return true;
}

void
py_throwup (PyObject *obj)
{
  bool flag = PyErr_Occurred ();
  if (!obj && !flag) {
    warn << "python call returned NULL, but no error set!\n";
    exit (2);
  }
  if (flag) {
    if (obj) {
      warn << "python error set, but valid object returned from call!\n";
    }
    PyErr_Print ();
    exit (2);
  }
}
