/* $Id: genpyc.C 3758 2008-11-13 00:36:00Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */

#include "rpcc.h"
#include "rxx.h"

str module;
str dotted_m;

typedef enum 
{ 
  PASS_ONE = 1, PASS_TWO = 2, PASS_THREE = 3, N_PASSES = 4
} pass_num_t;

qhash<str,str> enum_tab;
qhash<str,u_int32_t> proc_tab;
qhash<str,str> literal_tab;

static str
pyc_type (const str &s)
{
  if (s == "void")
    return s;

  strbuf b;
  b << "py_" << s;
  return b;
}

static str
pyw_type (const str &s)
{
  strbuf b;
  b << "pyw_" << s;
  return b;
}

static str
py_type_obj (const str &in)
{
  strbuf b;
  b << in << "_Type";
  return pyc_type (b);
}
static str
enum_cast (const str &i, const str &v)
{
  strbuf b;
  b << "(" << i << " )" << v;
  return b;
}

static str
union_tag_cast (const rpc_union *u, const str &v)
{
  return enum_cast (u->tagtype, v);
}


static void
pmshl (str id)
{
  aout <<
    "void *" << id << "_alloc ();\n"
    XDR_RETURN " xdr_" << id << " (XDR *, void *);\n";
}

static str
rpc_decltype (const rpc_decl *d)
{
  str wt = pyw_type (d->type);
  strbuf b;
  if (d->type == "string") {
    b << pyw_type ("rpc_str") << "<" << d->bound << ">";
  } else if (d->type == "opaque")
    switch (d->qual) {
    case rpc_decl::ARRAY:
      b << pyw_type ("rpc_opaque") << "<" << d->bound << ">";
      break;
    case rpc_decl::VEC:
      b << pyw_type ("rpc_bytes") << "<" << d->bound << ">";
      break;
    default:
      panic ("bad rpc_decl qual for opaque (%d)\n", d->qual);
      break;
    }
  else
    switch (d->qual) {
    case rpc_decl::SCALAR:
      b << pyw_type (d->type) ;
      break;
    case rpc_decl::PTR:
      b << pyw_type ("rpc_ptr") << "<" << pyc_type (d->type) 
	<< ", &" << py_type_obj (d->type) << ">";
      break;
    case rpc_decl::ARRAY:
      b << pyw_type ("rpc_array") << "<" << wt << ", " << d->bound << ">";
      break;
    case rpc_decl::VEC:
      b << pyw_type ("rpc_vec") << "<" << wt << ", " << d->bound << ">";
      break;
    default:
      panic ("bad rpc_decl qual (%d)\n", d->qual);
    }
  return b;
}

static void
pdecl_py (str prefix, const rpc_decl *d)
{
  str name = d->id;
  aout << prefix << rpc_decltype (d) << " " << pyw_type (name) << ";\n";
}

static void
pdecl (str prefix, const rpc_decl *d)
{
  str name = d->id;
  aout << prefix << rpc_decltype (d) << " " << name << ";\n";
}

static str
obj_in_class (const rpc_decl *d)
{
  strbuf b;
  b << d->id << ".obj ()";
  return b;
}

static void
dump_init_frag (str prfx, const rpc_decl *d)
{
  aout << prfx << "if (!self->" << d->id << ".init ()) {\n"
       << prfx << "  Py_DECREF (self);\n"
       << prfx << "  return NULL;\n"
       << prfx << "}\n"
       << "\n";
}

static str
get_inner_obj (const str &vn, const str &ct)
{
  strbuf bout;
  bout << "  if (!_obj) {\n"
       << "    PyErr_SetString (PyExc_RuntimeError,\n"
       << "                     \"null wrapped obj unexpected\");\n"
       << "    return false;\n"
       << "  }\n"
       << "  " << ct << " *" << vn 
       << " = casted_obj ();\n" ;

  return bout;
}

static void
dump_w_class_init_func (const str &wt, const str &pt)
{
  aout << "inline bool\n"
       << wt << "::init ()\n"
       << "{\n"
       << "  _typ = &" << pt << ";\n"
       << "  return alloc ();\n"
       << "}\n\n";
}


static void
dump_union_init_and_clear (const rpc_union *u)
{
  str ct = pyc_type (u->id);
  str wt = pyw_type (u->id);
  str pt = py_type_obj (u->id);

  dump_w_class_init_func (wt, pt);
  
  aout << "inline bool\n"
       << wt << "::clear ()\n"
       << "{\n"
       << get_inner_obj ("o", ct)
       << "  o->_base.destroy ();\n"
       << "  Py_DECREF (_obj);\n"
       << "  _obj = NULL;\n"
       << "  return true;\n"
       << "}\n\n" ;
}

static void
dump_w_enum_clear_func (const str &id)
{
  const str wt = pyw_type (id);

  aout << "inline bool\n"
       << wt << "::clear ()\n"
       << "{\n"
       << "  Py_XDECREF (_obj);\n"
       << "  _obj = NULL;\n"
       << "  return true;\n"
       << "}\n\n" ;

}

static void
dump_w_class_clear_func (const rpc_struct *rs)
{
  str ct = pyc_type (rs->id);
  str wt = pyw_type (rs->id);
  aout << "inline bool\n"
       << wt << "::clear ()\n"
       << "{\n"
       << "  PyObject *tmp = _obj;\n"
       << "  _obj = NULL;\n"
       << "  Py_XDECREF (tmp);\n"
       << "  return true;\n"
       << "}\n\n";
}

static void
dump_class_uncasted_new_func (const rpc_struct *rs)
{
  str ct = pyc_type (rs->id);
  aout << "static " << ct << " *\n"
       << ct
       << "_uncasted_new (PyTypeObject *type, "
       <<                 "PyObject *args, PyObject *kwds)\n"
       << "{\n"
       << "  " << ct << " *self;\n"
       << "  PYDEBUG_PYALLOC (" << ct << ");\n"
       << "  if (!(self = (" << ct << " *)type->tp_alloc (type, 0)))\n"
       << "    return NULL;\n";
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++) {
    dump_init_frag ("  ", rd);
  }
  aout << "  return self;\n"
       << "}\n\n";
}

static void
dump_class_new_func (const str &id)
{
  str ct = pyc_type (id);
  aout << "static PyObject *\n"
       << ct 
       << "_new (PyTypeObject *type, PyObject *args, PyObject *kwds)\n"
       << "{\n"
       << "  " << ct << " *o = " << ct << "_uncasted_new (type, args, kwds);\n"
       << "  return reinterpret_cast<PyObject *> (o);\n"
       << "}\n\n";
}

static void
dump_enum_new_func (const str &id)
{
  str ct = pyc_type (id);
  aout << "static " << ct << " *\n"
       << ct << "_new (PyTypeObject *type, PyObject *args, PyObject *kwds)\n"
       << "{\n"
       << "  " << ct << " *self;\n"
       << "  PYDEBUG_PYALLOC (" << ct << ");\n"
       << "  if (!(self = (" << ct << " *)type->tp_alloc (type, 0)))\n"
       << "    return NULL;\n"
       << "  return self;\n"
       << "}\n\n";
}

static void
dump_class_init_func (const rpc_struct *rs)
{
  str ct = pyc_type (rs->id);
  aout << "static int\n"
       << ct << "_init (" << ct << " *self, PyObject *args, PyObject *kwds)\n"
       << "{\n";
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++) {
    aout << "  PyObject *" << rd->id << " = NULL;\n";
  }
  aout << "\n";
  aout << "  static char *kwlist[] = { ";
  bool first = true;
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++) {
    if (first)
      first = false;
    else
      aout << ",\n                            ";
    aout << "\"" << rd->id << "\"";
  }
  aout << ", NULL};\n\n";

  aout << "  if (!PyArg_ParseTupleAndKeywords (args, kwds, \"|";
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++) 
    aout << "O";
  aout << "\",\n"
       << "                                    kwlist,\n"
       << "                                    ";
  first = true;
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++) {
    if (first)
      first = false;
    else {
      aout << ",\n                                    ";
    }
    aout << "&" << rd->id;
  }
  aout << "))\n"
       << "    return -1;\n\n";

  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++) {
    aout << "  if (" << rd->id << " && "
	 << " !self->" << rd->id << ".safe_set_obj (" << rd->id << ")) {\n"
	 << "    py_prepend_err_msg (\"in field '" << rd->id << "': \");\n"
	 << "    return -1;\n"
	 << "  }\n";
  }
  
  
  aout << "  return 0;\n"
       << "}\n\n";
}

static void
dump_str_func (const str &id)
{
  str ct = pyc_type (id);
  aout << "static PyObject *\n"
       << ct << "_str (" << ct << " *obj)\n"
       << "{\n"
       << "  return pretty_print_to_py_str_obj (obj);\n"
       << "}\n\n";
}

static void
dump_print_func (const str &id)
{
  str ct = pyc_type (id);
  aout << "static int\n"
       << ct << "_print (" << ct << " *obj, FILE *fp, int flags)\n"
       << "{\n"
       << "  return pretty_print_to_fd (obj, fp);\n"
       << "}\n\n";

}

static void
dump_union_init_func (const rpc_union *u)
{
  str ct = pyc_type (u->id);
  aout << "static int\n"
       << ct << "_init (" << ct << " *self, PyObject *args, PyObject *kwds)\n"
       << "{\n"
       << "  int state = 0;\n"
       << "  if (!PyArg_ParseTuple (args, \"|i\", &state)) {\n"
       << "    return -1;\n"
       << "  }\n"
       << "  self->set_" << u->tagid << " ("
       << union_tag_cast (u, "state") << ");\n"
       << "  return 0;\n"
       << "}\n\n";
}

static void
dump_class_dealloc_func (const rpc_struct *rs)
{
  str ct = pyc_type (rs->id);
  aout << "static void\n"
       << ct << "_dealloc (" << ct << " *self)\n"
       << "{\n";
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++) {
    aout << "  self->" << rd->id << ".clear ();\n";
  }
  aout << "  PYDEBUG_PYFREE (" << ct << ");\n"
       << "  self->ob_type->tp_free ((PyObject *) self);\n"
       << "}\n\n";
}

static void
dump_allocator (const str &id)
{
  str wt = pyw_type (id);
  str ppt = py_type_obj (id);
  aout << "void *\n"
       << wt << "_alloc ()\n"
       << "{\n"
       << "  return (void *)New " << wt << ";\n"
       << "}\n\n";
}

static void
dump_enum_in_range_decl (const rpc_enum *e)
{
  const str id = e->id;
  aout << "bool " << id << "_is_in_range (" << id << " t);\n"
       << "bool " << id << "_is_in_range_set_error (" << id << " t);\n\n";
}

static void
dump_enum_in_range (const rpc_enum *e)
{
  const str id = e->id;
  aout << "bool\n"
       << id << "_is_in_range (" << id << " t)\n"
       << "{\n"
       << "  return ";
  bool first = true;
  for (const rpc_const *rc = e->tags.base (); rc < e->tags.lim (); rc++) {
    if (first) {
      first = false;
    } else {
      aout << " ||\n         ";
    }
    aout << "t == " << rc->id;
  }

  aout << ";\n"
       << "}\n\n"
       << "bool\n"
       << id << "_is_in_range_set_error (" << id << " t)\n"
       << "{\n"
       << "  bool rc = " << id << "_is_in_range (t);\n"
       << "  if (!rc)\n"
       << "    PyErr_SetString (PyExc_OverflowError, "
       <<                       "\"Enum value is not in range\");\n"
       << "  return rc;\n"
       << "}\n\n";
}

static void
dump_w_enum_convert (const str &id)
{
  const str wt = pyw_type (id);

  aout << "template<> struct converter_t<" << wt << ">\n"
       << "{\n"
       << "  static bool convert (PyObject *obj, " << id << " *res)\n"
       << "  {\n"
       << "    if (obj == NULL) {\n"
       << "      PyErr_SetString(PyExc_RuntimeError, "
       << "\"Unexpected NULL arg to setter\");\n"
       << "      return false;\n"
       << "    }\n"
       << "    " << id << " t = "
       << enum_cast (id, "0") << ";\n"
       << "    if (PyInt_Check (obj)) {\n"
       << "      t = " << enum_cast (id, "PyInt_AsLong (obj)") << ";\n"
       << "    } else if (PyLong_Check (obj)) {\n"
       << "      t = " << enum_cast (id, "PyLong_AsLong (obj)") << ";\n"
       << "    } else {\n"
       << "      PyErr_SetString (PyExc_TypeError,\n"
       << "                       \"Non-integral type given as switch tag\")"
       <<                         ";\n"
       << "      return false;\n"
       << "    }\n"
       << "    if (!" << id << "_is_in_range_set_error (t)) {\n"
       << "       return false;\n"
       << "    }\n"
       << "    *res = t;\n"
       << "    return true;\n"
       << "  }\n\n"

       << "  static PyObject *convert (PyObject *obj)\n"
       << "  {\n"
       << "    " << id << " t;\n"
       << "    if (!convert (obj, &t))\n"
       << "      return NULL;\n"
       << "    return Py_BuildValue (\"i\", int (t));\n"
       << "  }\n"
       << "};\n\n";
}

static void
dump_convert (const str &id)
{
  str ptt = py_type_obj (id);
  str pct = pyc_type (id);
  str wt = pyw_type (id);
  
  aout << "template<> struct converter_t<" << wt << ">\n"
       << "{\n"
       << "  static PyObject *convert (PyObject *obj)\n"
       << "  {\n"
       << "    if (!PyObject_IsInstance (obj, (PyObject *)&" 
       <<                                           ptt <<  ")) {\n"
       << "       PyErr_SetString (PyExc_TypeError, \"expected object of type "
       <<                                           id << "\");\n"
       << "       return NULL;\n"
       << "    }\n"
       << "    Py_INCREF (obj);\n"
       << "    return obj;\n"
       << "  }\n"
       << "};\n\n";
}

static void
dump_xdr_func (const str &id)
{
  str ct = pyc_type (id);
  str wt = pyw_type (id);
  aout << XDR_RETURN << "\n"
       << "xdr_" << wt << " (XDR *xdrs, void *objp)\n"
       << "{\n"
       << "  return xdr_doit<" << wt << "> (xdrs, objp);\n"
       << "}\n\n";
}

static void
dump_w_rpc_traverse (const str &id)
{
  str ct = pyc_type (id);
  str wt = pyw_type (id);

  aout << "template<class T> bool\n"
       << "rpc_traverse (T &t, " << wt << " &wo)\n"
       << "{\n"
       << "  " << ct << " *io = wo.casted_obj ();\n"
       << "  if (!io) {\n"
       << "    PyErr_SetString (PyExc_UnboundLocalError,\n"
       << "                     \"uninitialized XDR field\");\n"
       << "    return false;\n"
       << "  }\n"
       << "  return rpc_traverse (t, *io);\n"
       << "}\n\n" ;
}

static void
dump_rpc_traverse (const rpc_struct *rs)
{

  aout << "template<class T> "
       << (rs->decls.size () > 1 ? "" : "inline ") << "bool\n"
       << "rpc_traverse (T &t, " << pyc_type (rs->id) << " &obj)\n"
       << "{\n";
  str fnc;
  const rpc_decl *rd = rs->decls.base ();
  if (rd < rs->decls.lim ()) {
    aout << "  return rpc_traverse (t, obj." << rd->id << ")";
    rd++;
    while (rd < rs->decls.lim ()) { 
      aout << "\n         && rpc_traverse (t, obj." << rd->id << ")";
      rd++;
    }
    aout << ";\n";
  }
  else
    aout << "  return true;\n";
  aout << "}\n\n";
}

static void
dump_class_members (const str &t)
{
  aout << "static PyMemberDef " << t << "_members[] = {\n"
       << "  {NULL}\n"
       << "};\n\n" ;
}

static void
dump_class_method_decls (const str &ct)
{
  aout << "PyObject * " << ct << "_str2xdr (" << ct 
       << " *self, PyObject *args);\n"
       << "PyObject * " << ct << "_xdr2str (" << ct << " *self);\n"
       << "PY_XDR_OBJ_WARN_DECL (" << ct << ", warn);\n"
       << "PY_XDR_OBJ_WARN_DECL (" << ct << ", warnx);\n\n";
}

static void
dump_class_methods (const str &id)
{
  str ct = pyc_type (id);

  aout << "PyObject *\n"
       << ct << "_xdr2str (" << ct << " *self)\n"
       << "{\n"
       << "  str ret = xdr2str (*self);\n"
       << "  if (ret) {\n"
       << "    return PyString_FromStringAndSize (ret.cstr (), ret.len ());\n"
       << "  } else {\n"
       << "    Py_INCREF (Py_None);\n"
       << "    return Py_None;\n"
       << "  }\n"
       << "}\n\n"
       << "PyObject *\n"
       << ct << "_str2xdr (" << ct << " *self, PyObject *args)\n"
       << "{\n"
       << "  PyObject *input;\n" 
       << "  if (!PyArg_Parse (args, \"(O)\", &input))\n"
       << "    return NULL;\n"
       << "\n"
       << "  if (!PyString_Check (input)) {\n"
       << "    PyErr_SetString (PyExc_TypeError,\n"
       << "                     \"Expected a string object\");\n"
       << "    return NULL;\n"
       << "  }\n"
       << "\n"
       << "  char *dat;\n"
       << "  int len;\n"
       << "  if (PyString_AsStringAndSize (input, &dat, &len) < 0) {\n"
       << "    PyErr_SetString (PyExc_Exception,\n"
       << "                     \"Garbled / corrupted string passed\");\n"
       << "    return NULL;\n"
       << "  }\n"
       << "  if (len == 0 || !dat) {\n"
       << "    goto succ;\n"
       << "  }\n"
       << "  if (!str2xdr (*self, str (dat, len))) {\n"
       << "    PyErr_SetString (AsyncXDR_Exception,\n"
       << "                     \"Failed to demarshal XDR data\");\n"
       << "    return NULL;\n"
       << "  }\n"
       << " succ:\n"
       << "  Py_INCREF (Py_None);\n"
       << "  return Py_None;\n"
       << "}\n\n"
       << "PY_XDR_OBJ_WARN (" << ct << ", warn);\n"
       << "PY_XDR_OBJ_WARN (" << ct << ", warnx);\n\n";
}

static void
dump_class_methods_struct (const str &ct)
{
  aout << "static PyMethodDef " << ct << "_methods[] = {\n"
       << "  {\"warn\", (PyCFunction)" << ct << "_warn, "
       <<        " METH_NOARGS,\n"
       << "   \"RPC-pretty print this method to stderr via async::warn\"},\n"
       << "  {\"warnx\", (PyCFunction)" << ct << "_warnx, "
       <<        " METH_NOARGS,\n"
       << "   \"RPC-pretty print this method to stderr via async::warnx\"},\n"
       << "  {\"xdr2str\", (PyCFunction)" << ct << "_xdr2str, "
       <<        " METH_NOARGS,\n"
       << "   \"Export RPC structure to a regular string buffer\"},\n"
       << "  {\"str2xdr\", (PyCFunction)" << ct << "_str2xdr, "
       <<        " METH_VARARGS,\n"
       << "   \"Import RPC structure from a regular string buffer\"},\n"
       << " {NULL}\n"
       << "};\n\n";
}

static void
dump_type_object_decl (const str &id)
{
  str t = py_type_obj (id);
  str ct = pyc_type (id);
  str pt = id;
  aout << "PY_CLASS_DECL(" << ct << ");\n\n";
}

static void
dump_type_object (const str &id)
{
  str t = py_type_obj (id);
  str ct = pyc_type (id);
  str pt = id;

  aout << "PY_CLASS_DEF2(" << ct << ", \"" << dotted_m << "." << pt 
       <<                "\", 1, dealloc, "
       <<                "-1, \"" << pt << " object\",\n"
       << "              methods, members, getsetters, init, new, 0,\n"
       << "              str, 0, print);\n"
       << "\n";
}

static str
py_rpc_procno_method (const str &ct, const str &id)
{
  strbuf b;
  b << ct << "_RPC_PROCNO_" << id;
  return b;
}

static str
py_rpcprog_type (const rpc_program *prog, const rpc_vers *rv)
{
  return pyc_type (rpcprog (prog, rv));
}

static str
py_rpcprog_extension (const rpc_program *prog, const rpc_vers *rv)
{
  strbuf sb;
  sb << "py_" << rpcprog (prog, rv) << "_tbl";
  return sb;
}

static void
dump_prog_py_obj (const rpc_program *prog)
{
  for (const rpc_vers *rv = prog->vers.base (); rv < prog->vers.lim (); rv++) {
    str type = py_rpcprog_type (prog, rv);
    str objType = rpcprog (prog, rv);
    aout << "struct " << type << " : public py_rpc_program_t {\n"
	 << "};\n\n"
	 << "PY_CLASS_NEW(" << type << ")\n\n"
	 << "static int\n"
	 << type << "_init (" << type << " *self, PyObject *args, "
	 << "PyObject *kwds)\n"
	 << "{\n"
	 << "  if (!PyArg_ParseTuple (args, \"\"))\n"
	 << "    return -1;\n"
	 << "  self->prog = &" << rpcprog (prog, rv) << ";\n" 
	 << "  self->pytab = " << py_rpcprog_extension (prog, rv) << ";\n"
	 << "  return 0;\n"
	 << "}\n\n";

    for (const rpc_proc *rp = rv->procs.base (); rp < rv->procs.lim (); rp++) {
      aout << "static PyObject *\n"
	   << py_rpc_procno_method (type, rp->id) 
	   <<  " (" <<  type << " *self)"
	   << "{\n"
	   << "  return Py_BuildValue (\"i\", " << rp->id << ");\n"
	   << "}\n\n";
    }

    aout << "static PyMethodDef " << type << "_methods[] = {\n";
    for (const rpc_proc *rp = rv->procs.base (); rp < rv->procs.lim (); rp++) {
      aout << "  {\"" << rp->id <<  "\", (PyCFunction)" 
	   << py_rpc_procno_method (type, rp->id)
	   << ", METH_NOARGS,\n"
	   << "   \"RPC Procno for " << rp->id << "\"},\n";
    }
    aout << "  {NULL}\n"
	 << "};\n\n";


	 
    aout << "PY_CLASS_DEF(" << type << ", \"" << dotted_m << "." 
	 << objType << "\", "
	 <<               "1, 0, -1,\n"
	 << "             \"" << objType << " RPC program wrapper\", "
	 <<               "methods, 0, 0, init, new, 0);\n\n";
  }
  
}

static void
dump_getter_decl (const str &cl, const str &id)
{
  aout << "PyObject * " << cl << "_get" << id 
       << " (" << cl << " *self, void *closure);\n";
}

static void
dump_getter (const str &cl, const rpc_decl *d)
{
  strbuf b;
  b << "self->" << obj_in_class (d);
  str obj = b;
  aout << "PyObject * " << cl << "_get" << d->id 
       << " (" << cl << " *self, void *closure)\n"
       << "{\n"
       << "  PyObject *obj = " << obj << ";\n"
       << "  if (!obj) {\n"
       << "    PyErr_SetString (PyExc_UnboundLocalError,\n"
       << "                     \"undefined class variable\");\n"
       << "    return NULL;\n"
       << "  }\n"
       << "  Py_XINCREF (obj);\n"
       << "  return obj;\n"
       << "}\n\n";
}


static void
dump_setter (const str &cl, const rpc_decl *d)
{
  aout << "int " << cl << "_set" << d->id 
       << " (" << cl << " *self, PyObject *value, void *closure)\n"
       << "{\n"
       << "  if (value == NULL) {\n"
       << "    PyErr_SetString(PyExc_RuntimeError, "
       << "\"Unexpected NULL arg to setter\");\n"
       << "    return -1;\n"
       << "  }\n"
       << "  return self->" << d->id << ".safe_set_obj (value) ? 0 : -1;\n"
       << "};\n\n";
}



static void
dump_setter_decl (const str &cl, const str &id)
{
  aout << "int " << cl << "_set" << id
       << " (" << cl << " *self, PyObject *value, void *closure);\n";
}

static void
dump_getsetter_decl (const str &cl, const str &id)
{
  if (id) {
    dump_getter_decl (cl, id);
    dump_setter_decl (cl, id);
  }
}

static void
dump_getsetter (const str &cl, const rpc_decl *d)
{
  if (d->id) {
    dump_getter (cl, d);
    dump_setter (cl, d);
  }
}

static void
dump_getsetter_decls (const rpc_struct *rs)
{
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++)
    dump_getsetter_decl (pyc_type (rs->id), rd->id);
  aout << "\n\n";
}

static void
dump_union_getsetter_decls (const rpc_union *u)
{
  str ct = pyc_type (u->id);
  dump_getsetter_decl (ct, u->tagid);
  for (const rpc_utag *rd = u->cases.base (); rd < u->cases.lim (); rd ++)
    dump_getsetter_decl (ct, rd->tag.id);
  aout << "\n\n";
}

static void
dump_union_tag_getter (const rpc_union *u)
{
  str cl = pyc_type (u->id);
  aout << "PyObject *\n" 
       << cl << "_get" << u->tagid
       << " (" << cl << " *self, void *closure)\n"
       << "{\n"
       << "  PyObject *obj = Py_BuildValue (\"i\", int(self->"
       <<                                 u->tagid << "));\n"
       << "  return obj;\n"
       << "}\n\n";

}

static void
dump_union_tag_setter (const rpc_union *u)
{
  str cl = pyc_type (u->id);
  const str tt = u->tagtype;
  const str twt = pyw_type (tt);
  aout << "int\n" 
       << cl << "_set" << u->tagid 
       << " (" << cl << " *self, PyObject *value, void *closure)\n"
       << "{\n"
       << "  " << tt  << " t;\n"
       << "  if (!converter_t<" << twt << ">::convert (value, &t))\n"
       << "    return -1;\n"
       << "  self->set_" << u->tagid << " (t);\n"
       << "  return 0;\n"
       << "}\n\n";
}

static void
dump_union_setter (const rpc_union *u, const rpc_utag *t)
{
  str cl = pyc_type (u->id);
  aout << "int\n" << cl << "_set" << t->tag.id 
       << " (" << cl << " *self, PyObject *value, void *closure)\n"
       << "{\n"
       << "  if (value == NULL) {\n"
       << "    PyErr_SetString (PyExc_RuntimeError, "
       << "\"Unexpected NULL arg to setter\");\n"
       << "    return -1;\n"
       << "  }\n";
  if (t->swval) {
    aout << "  self->set_" << u->tagid << " (" << t->swval << ");\n";
  } else {
    aout << "  if (!" << cl << "_is_def_case (self->" << u->tagid << ")) {\n"
	 << "    PyErr_SetString (AsyncUnion_Exception,\n"
	 << "                     \"implicit switch to default case\");\n"
	 << "    return -1;\n"
	 << "  }\n";
  }
  aout << "  return self->" << t->tag.id 
       << "->safe_set_obj (value) ? 0 : -1;\n"
       << "}\n\n";
}

static void
dump_union_getter (const rpc_union *u, const rpc_utag *t)
{
  str cl = pyc_type (u->id);
  aout << "PyObject *\n"
       << cl << "_get" << t->tag.id << " (" << cl << " *self, void *closure)\n"
       << "{\n";
  if (t->swval) 
    aout << "  if (self->" << u->tagid << " != " << t->swval << ") {\n"
	 << "    PyErr_SetString (AsyncUnion_Exception,\n"
	 << "                    \"union does not have switch type: "
	 <<                       t->swval << "\");\n";
  else
    aout << "  if (!" << cl << "_is_def_case (self->" << u->tagid << ")) {\n"
	 << "    PyErr_SetString (AsyncUnion_Exception,\n"
	 << "                     \"union not set to default case\");\n";
  
  aout << "    return NULL;\n"
       << "  }\n"
       << "  PyObject *obj = self->" << t->tag.id << "->obj ();\n"
       << "  if (!obj) {\n"
       << "    PyErr_SetString (PyExc_UnboundLocalError,\n"
       << "                     \"unbound XDR variable\");\n"
       << "    return NULL;\n"
       << "  }\n"
       << "  Py_XINCREF (obj);\n"
       << "  return obj;\n"
       << "}\n\n";
}

static void
dump_union_is_valid_case (const rpc_union *u)
{

}

static void
dump_union_is_def_case (const rpc_union *u)
{
  str cl = pyc_type (u->id);
  aout << "bool\n"
       << cl << "_is_def_case (const " 
       << u->tagtype << " &tag)\n"
       << "{\n"
       << "  return ";
  bool first = true;
  for (const rpc_utag *rd = u->cases.base (); rd < u->cases.lim (); rd++) {
    if (!rd->swval)
      continue;
    if (!first)
      aout << "\n         && ";
    else
      first = false;
    aout << "tag != " << rd->swval;
  }
  if (first)
    aout << "true";
  aout << ";\n"
       << "}\n\n";
}

static void
dump_union_tag_getsetter (const rpc_union *u)
{
  dump_union_tag_getter (u);
  dump_union_tag_setter (u);
}

static void
dump_union_getsetter (const rpc_union *u, const rpc_utag *t)
{
  if (t->tag.id) {
    dump_union_getter (u, t);
    dump_union_setter (u, t);
  }
}

static void
dump_union_getsetters (const rpc_union *u)
{
  dump_union_is_valid_case (u);
  dump_union_tag_getsetter (u);
  dump_union_is_def_case (u);
  for (const rpc_utag *rd = u->cases.base (); rd < u->cases.lim (); rd++) {
    dump_union_getsetter (u, rd);
  }
}

static void
dump_getsetters (const rpc_struct *rs)
{
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++)
    dump_getsetter (pyc_type (rs->id), rd);
}

static void
dump_getsetter_table_row (const str &prfx, const str &typ, const str &id,
			  const str &txt)
{
  aout << prfx << "{\"" << id << "\", "
       <<            "(getter)" << typ << "_get" << id << ", "
       <<            "(setter)" << typ << "_set" << id << ",\n"
       << prfx << " \"" << txt << ": " << id << "\", "
       <<            "NULL },\n" ;
}

static void
dump_getsetter_table (const rpc_struct *rs)
{
  str ct = pyc_type (rs->id);
  aout << "static PyGetSetDef " << ct << "_getsetters[] = {\n";
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++)
    dump_getsetter_table_row ("  ", ct, rd->id, "class variable");
  aout << "  {NULL}\n"
       << "};\n\n";
}

static void
dump_union_getsetter_table (const rpc_union *u)
{
  str ct = pyc_type (u->id);
  aout << "static PyGetSetDef " << ct << "_getsetters[] = {\n";
  dump_getsetter_table_row ("  ", ct, u->tagid, "union switch element");
  for (const rpc_utag *rd = u->cases.base (); rd < u->cases.lim (); rd ++)
    if (rd->tag.id)
      dump_getsetter_table_row ("  ", ct, rd->tag.id, "union case");
  aout << "  {NULL}\n"
       << "};\n\n";
}

static void
print_print (str type)
{
  str pref (strbuf ("%*s", int (8 + type.len ()), ""));
  aout << "void\n"
    "print_" << type << " (const void *_objp, const strbuf *_sbp, "
    "int _recdepth,\n" <<
    pref << "const char *_name, const char *_prefix)\n"
    "{\n"
    "  rpc_print (_sbp ? *_sbp : warnx, *static_cast<const " << type
       << " *> (_objp),\n"
    "             _recdepth, _name, _prefix);\n"
    "}\n";
}


static void
print_w_struct (const str &id)
{
  str wt = pyw_type (id);
  str ct = pyc_type (id);
  aout <<
    "const strbuf &\n"
    "rpc_print (const strbuf &sb, const " << wt << " &w, "
    "int recdepth,\n"
    "           const char *name, const char *prefix)\n"
    "{\n"
    "  if (w.print_err (sb, recdepth, name, prefix)) return sb;\n"
    "  const " << ct << " *o = w.const_casted_obj ();\n"
    "  return o ? rpc_print (sb, *o, recdepth, name, prefix) : sb;\n"
    "}\n\n";
}

static void
print_struct (const rpc_struct *s)
{
  str ct = pyc_type (s->id);
  aout <<
    "const strbuf &\n"
    "rpc_print (const strbuf &sb, const " << ct << " &obj, "
    "int recdepth,\n"
    "           const char *name, const char *prefix)\n"
    "{\n"
    "  if (name) {\n"
    "    if (prefix)\n"
    "      sb << prefix;\n"
    "    sb << \"" << s->id << " \" << name << \" = \";\n"
    "  };\n"
    "  const char *sep;\n"
    "  str npref;\n"
    "  if (prefix) {\n"
    "    npref = strbuf (\"%s  \", prefix);\n"
    "    sep = \"\";\n"
    "    sb << \"{\\n\";\n"
    "  }\n"
    "  else {\n"
    "    sep = \", \";\n"
    "    sb << \"{ \";\n"
    "  }\n";
  const rpc_decl *dp = s->decls.base (), *ep = s->decls.lim ();
  if (dp < ep)
    aout <<
      "  rpc_print (sb, obj." << dp->id << ", recdepth, "
      "\"" << dp->id << "\", npref);\n";
  while (++dp < ep)
    aout <<
      "  sb << sep;\n"
      "  rpc_print (sb, obj." << dp->id << ", recdepth, "
      "\"" << dp->id << "\", npref);\n";
  aout <<
    "  if (prefix)\n"
    "    sb << prefix << \"};\\n\";\n"
    "  else\n"
    "    sb << \" }\";\n"
    "  return sb;\n"
    "}\n";
  print_print (ct);
  print_print (pyw_type (s->id));
}

static void
print_case (str prefix, const rpc_union *rs, const rpc_utag *rt)
{
  if (rt->tag.type != "void")
    aout
      << prefix << "sb << sep;\n"
      << prefix << "rpc_print (sb, *obj." << rt->tag.id << ", "
      " recdepth, \"" << rt->tag.id << "\", npref);\n";
  aout << prefix << "break;\n";
}

static void
print_break (str prefix, const rpc_union *rs)
{
  aout << prefix << "break;\n";
}


static void
print_union (const rpc_union *s)
{
  str ct = pyc_type (s->id);
  aout <<
    "const strbuf &\n"
    "rpc_print (const strbuf &sb, const " << ct << " &obj, "
    "int recdepth,\n"
    "           const char *name, const char *prefix)\n"
    "{\n"
    "  if (name) {\n"
    "    if (prefix)\n"
    "      sb << prefix;\n"
    "    sb << \"" << s->id << " \" << name << \" = \";\n"
    "  };\n"
    "  const char *sep;\n"
    "  str npref;\n"
    "  if (prefix) {\n"
    "    npref = strbuf (\"%s  \", prefix);\n"
    "    sep = \"\";\n"
    "    sb << \"{\\n\";\n"
    "  }\n"
    "  else {\n"
    "    sep = \", \";\n"
    "    sb << \"{ \";\n"
    "  }\n"
    "  rpc_print (sb, obj." << s->tagid << ", recdepth, "
    "\"" << s->tagid << "\", npref);\n";
  pswitch ("  ", s, "obj." << s->tagid, print_case, "\n", print_break);
  aout <<
    "  if (prefix)\n"
    "    sb << prefix << \"};\\n\";\n"
    "  else\n"
    "    sb << \" }\";\n"
    "  return sb;\n"
    "}\n";
  print_print (ct);
  print_print (pyw_type (s->id));
}

static void
print_enum (const rpc_enum *s)
{
  str ct = s->id ; 
  aout <<
    "const strbuf &\n"
    "rpc_print (const strbuf &sb, const " << ct << " &obj, "
    "int recdepth,\n"
    "           const char *name, const char *prefix)\n"
    "{\n"
    "  char *p;\n"
    "  switch (obj) {\n";
  for (const rpc_const *cp = s->tags.base (),
	 *ep = s->tags.lim (); cp < ep; cp++)
    aout <<
      "  case " << cp->id << ":\n"
      "    p = \"" << cp->id << "\";\n"
      "    break;\n";
  aout <<
    "  default:\n"
    "    p = NULL;\n"
    "    break;\n"
    "  }\n"
    "  if (name) {\n"
    "    if (prefix)\n"
    "      sb << prefix;\n"
    "    sb << \"" << s->id << " \" << name << \" = \";\n"
    "  };\n"
    "  if (p)\n"
    "    sb << p;\n"
    "  else\n"
    "    sb << int (obj);\n"
    "  if (prefix)\n"
    "    sb << \";\\n\";\n"
    "  return sb;\n"
    "};\n\n"
    "const strbuf &\n"
    "rpc_print (const strbuf &sb, const " << pyc_type (ct) << " &obj, "
    "int recdepth,\n"
    "           const char *name, const char *prefix)\n"
    "{\n"
    "  return rpc_print (sb, obj.value, recdepth, name, prefix);\n"
    "}\n\n";

  print_print (ct);
  print_print (pyw_type (s->id));
}

static void
dumpprint (const rpc_sym *s)
{
  switch (s->type) {
  case rpc_sym::STRUCT:
    print_struct (s->sstruct.addr ());
    print_w_struct (s->sstruct.addr ()->id);
    break;
  case rpc_sym::UNION:
    print_union (s->sunion.addr ());
    print_w_struct (s->sunion.addr ()->id);
    break;
  case rpc_sym::ENUM:
    print_enum (s->senum.addr ());
    print_w_struct (s->senum.addr ()->id);
    break;
  case rpc_sym::TYPEDEF:
    print_print (pyw_type (s->stypedef->id));
  default:
    break;
  }
}

static void
dumpstruct_mthds (const rpc_sym *s)
{
  const rpc_struct *rs = s->sstruct.addr ();
  dump_getsetters (rs);
  dump_allocator (rs->id);
  dump_class_methods (rs->id);
}

static void
dumpunion_mthds (const rpc_sym *s)
{
  const rpc_union *u = s->sunion.addr ();
  dump_union_getsetters (u);
  dump_allocator (u->id);
  dump_class_methods (u->id);
}

static void
dump_py_struct (const rpc_struct *rs)
{
  str ct = pyc_type (rs->id);
  aout << "struct " << ct << " {\n"
       << "  PyObject_HEAD\n";
  for (const rpc_decl *rd = rs->decls.base (); rd < rs->decls.lim (); rd++)
    pdecl ("  ", rd);
  aout << "};\n\n";
  aout << "RPC_STRUCT_DECL (" << ct << ")\n\n";
}



static void
dump_w_class (const str &ct, const str &wt, const str &pt, const str &id)
{
  aout << "struct " << wt << " : public pyw_tmpl_t<" << wt << ", "
       << ct << " >\n"
       << "{\n"
       << "  " << wt << " () : pyw_tmpl_t<" << wt << ", " 
       <<                             ct << " > (&" << pt << ")\n"
       << "    { alloc (); }\n"
       << "  " << wt << " (PyObject *o) :\n" 
       << "    pyw_tmpl_t<" << wt << ", " << ct << " > (o, &" 
       <<                  pt << ") {}\n"
       << "  " << wt << " (pyw_err_t e) :\n"
       << "    pyw_tmpl_t<" << wt << ", " << ct << " > (e, &" << pt << ") {}\n"
       << "  " << wt << " (const " << wt << " &p) :\n"
       << "    pyw_tmpl_t<" << wt << ", " << ct << " > (p) {}\n"
       << "  bool init ();\n"
       << "  bool clear ();\n"
       << "};\n\n";
}

static void
dump_type2str_decl (const str &prfx, const str &id)
{
  aout << prfx << "RPC_TYPE2STR_DECL (" << id << ");\n";
}

static void
dump_py_type2str_decl (const str &id) 
{
  dump_type2str_decl ("PY_", id);
}

static void
dump_type2str_decl (const str &id)
{
  dump_type2str_decl ("", id);
}

static void
dump_rpc_print_decl (const str &id)
{
  aout << "RPC_PRINT_DECL (" << id << ");\n";
}

static void
dump_w_class (const str &id)
{
  str ct = pyc_type (id);
  str wt = pyw_type (id);
  str pt = py_type_obj (id);
  dump_w_class (ct, wt, pt, id);
}

static void
dump_print_decls (const str &id)
{
  dump_py_type2str_decl (id);
  aout << "\n";
}

static void
dumpstruct_hdr (const rpc_sym *s)
{
  const rpc_struct *rs = s->sstruct.addr ();
  str wt = pyw_type (rs->id);

  dump_py_struct (rs);
  dump_type_object_decl (rs->id);
  dump_w_class (rs->id);
  dump_convert (rs->id);

  dump_print_decls (rs->id);

  dump_rpc_traverse (rs);
  dump_w_rpc_traverse (rs->id);

  dump_rpc_print_decl (wt);

  // moved to header for time being
  dump_w_class_init_func (pyw_type (rs->id), py_type_obj (rs->id));
  dump_w_class_clear_func (rs);
}

static void
dumpstruct (const rpc_sym *s)
{
  const rpc_struct *rs = s->sstruct.addr ();
  str ct = pyc_type (rs->id);

  dump_class_dealloc_func (rs);
  dump_class_uncasted_new_func (rs);
  dump_class_new_func (rs->id);
  dump_class_members (ct);
  dump_class_method_decls (ct);
  dump_class_methods_struct (ct);
  dump_getsetter_decls (rs);
  dump_getsetter_table (rs);
  dump_class_init_func (rs);
  dump_str_func (rs->id);
  dump_print_func (rs->id);
  dump_type_object (rs->id);

  dump_xdr_func (rs->id);
}

static void
punionmacro (str prefix, const rpc_union *rs, const rpc_utag *rt)
{
  if (rt->tag.type == "void")
    aout << prefix << "voidaction; \\\n";
  else
    aout << prefix << "action (" << pyw_type (rt->tag.type) << ", "
	 << rt->tag.id << "); \\\n";
  aout << prefix << "break; \\\n";
}

static void
punionmacrodefault (str prefix, const rpc_union *rs)
{
  aout << prefix << "defaction; \\\n";
  aout << prefix << "break; \\\n";
}

static void
dump_c_union (const rpc_sym *s)
{
  bool hasdefault = false;

  const rpc_union *rs = s->sunion.addr ();
  str ct = pyc_type (rs->id);
  str tct = rs->tagtype ; // pyc_type (rs->tagtype);
  aout << "\nstruct " << ct << " {\n"
       << "  PyObject_HEAD\n"
       << "  const " << tct << " " << rs->tagid << ";\n"
       << "  union {\n"
       << "    union_entry_base _base;\n";
  for (const rpc_utag *rt = rs->cases.base (); rt < rs->cases.lim (); rt++) {
    if (!rt->swval)
      hasdefault = true;
    if (rt->tagvalid && rt->tag.type != "void") {
      str type = rpc_decltype (&rt->tag);
      if (type[type.len ()-1] == '>')
	type = type << " ";
      aout << "    union_entry<" << type << "> "
	   << rt->tag.id << ";\n";
    }
  }
  aout << "  };\n\n";

  aout << "#define rpcunion_tag_" << ct << " " << rs->tagid << "\n";
  aout << "#define rpcunion_switch_" << ct 
       << "(swarg, action, voidaction, defaction) \\\n";
  pswitch ("  ", rs, "swarg", punionmacro, " \\\n", punionmacrodefault);

  aout << "\n"
       << "  " << ct << " (" << tct << " _tag = ("
       << tct << ") 0) : " << rs->tagid << " (_tag)\n"
       << "    { _base.init (); set_" << rs->tagid << " (_tag); }\n"

       << "  " << ct << " (" << "const " << ct << " &_s)\n"
       << "    : " << rs->tagid << " (_s." << rs->tagid << ")\n"
       << "    { _base.init (_s._base); }\n"
       << "  ~" << ct << " () { _base.destroy (); }\n"
       << "  " << ct << " &operator= (const " << ct << " &_s) {\n"
       << "    const_cast<" << tct << " &> ("
       << rs->tagid << ") = _s." << rs->tagid << ";\n"
       << "    _base.assign (_s._base);\n"
       << "    return *this;\n"
       << "  }\n\n";

  aout << "  void set_" << rs->tagid << " (" << tct << " _tag) {\n"
       << "    const_cast<" << tct << " &> (" << rs->tagid
       << ") = _tag;\n"
       << "    rpcunion_switch_" << ct << "\n"
       << "      (_tag, RPCUNION_SET, _base.destroy (), _base.destroy ());\n"
       << "  }\n";

  aout << "};\n";

  aout << "\ntemplate<class T> bool\n"
       << "rpc_traverse (T &t, " << ct << " &obj)\n"
       << "{\n"
       << "  " << tct << " tag = obj." << rs->tagid << ";\n"
       << "  if (!rpc_traverse (t, tag))\n"
       << "    return false;\n"
       << "  if (tag != obj." << rs->tagid << ")\n"
       << "    obj.set_" << rs->tagid << " (tag);\n\n"
       << "  rpcunion_switch_" << ct << "\n"
       << "    (obj." << rs->tagid << ", RPCUNION_TRAVERSE, "
       << "return true, return false);\n"
       << "}\n"
    ;

  // No Stompcasting for now
#if 0
       << "inline bool\n"
       << "rpc_traverse (const stompcast_t &s, " << ct << " &obj)\n"
       << "{\n"
       << "  rpcunion_switch_" << ct << "\n"
       << "    (obj." << rs->tagid << ", RPCUNION_REC_STOMPCAST,\n"
       << "     obj._base.destroy (); return true, "
       << "obj._base.destroy (); return true;);\n"
       << "}\n";
#endif

  aout << "RPC_UNION_DECL (" << ct << ")\n";

  aout << "\n";
}

static void
dump_union_dealloc_func (const rpc_union *u)
{
  str ct = pyc_type (u->id);
  aout << "static void\n"
       << ct << "_dealloc (" << ct << " *self)\n"
       << "{\n"
       << "  self->_base.destroy ();\n"
       << "  PYDEBUG_PYFREE (" << ct << ");\n"
       << "  self->ob_type->tp_free ((PyObject *) self);\n"
       << "}\n\n";
}

static void
dump_union_uncasted_new_func (const rpc_union *u)
{
  
  str ct = pyc_type (u->id);
  aout << "static " << ct << " *\n"
       << ct << "_uncasted_new (PyTypeObject *type, "
       <<                       "PyObject *args, PyObject *kwds)\n"
       << "{\n"
       << "  " << ct << " *self;\n"
       << "  PYDEBUG_PYALLOC (" << ct << ");\n"
       << "  if (!(self = (" << ct << " *)type->tp_alloc (type, 0)))\n"
       << "    return NULL;\n"
       << "  int arg = 0;\n"
       << "  if (args && !PyArg_ParseTuple (args, \"|i\", &arg))\n"
       << "    return NULL;\n"
       << "  self->_base.init ();\n"
       << "  self->set_" << u->tagid
       << " (" << union_tag_cast (u, "arg") << ");\n"
       << "  return self;\n"
       << "}\n\n";

}

static void
dumpunion_hdr (const rpc_sym *s)
{
  const rpc_union *u = s->sunion.addr ();
  str wt = pyw_type (u->id);

  dump_c_union (s);
  dump_type_object_decl (u->id);
  dump_w_class (u->id);
  dump_print_decls (u->id);
  dump_convert (u->id);
  dump_w_rpc_traverse (u->id);

  // moved to hdr for time being
  dump_union_init_and_clear (u);

  dump_rpc_print_decl (wt);
}


static void
dumpunion (const rpc_sym *s)
{
  const rpc_union *u = s->sunion.addr ();
  str ct = pyc_type (u->id);
  dump_union_dealloc_func (u);
  dump_union_uncasted_new_func (u);
  dump_class_new_func (u->id);
  dump_class_members (ct);
  dump_class_method_decls (ct);
  dump_class_methods_struct (ct);
  dump_union_getsetter_decls (u);
  dump_union_getsetter_table (u);
  dump_union_init_func (u);
  dump_str_func (u->id);
  dump_print_func (u->id);
  dump_type_object (u->id);

  dump_xdr_func (u->id);
}


static void
dump_c_enum (const rpc_enum *rs, bool d)
{
  int ctr = 0;
  str lastval;
  str ct = rs->id;

  if (d)
    aout << "enum " << ct << " {\n";

  for (const rpc_const *rc = rs->tags.base (); rc < rs->tags.lim (); rc++) {
    if (enum_tab[rc->id]) {
      warn << "duplicate enum key: " << rc->id << "\n";
    }
    str ev;
    strbuf b;
    if (rc->val) {
      lastval = rc->val;
      ctr = 1;
      b << rc->val;
      ev = b;
      if (d)
	aout << "  " << rc->id << " = " << rc->val << ",\n";
    }
    else if (lastval && (isdigit (lastval[0]) || lastval[0] == '-'
			 || lastval[0] == '+')) {
     
      long l = strtol (lastval, NULL, 0) + ctr ++;
      b << l;
      ev = b;
      if (d)
	aout << "  " << rc->id << " = " << ev << ",\n";
    } else if (lastval) {
      b << lastval << " + " << ctr ++;
      ev = b;
      if (d)
	aout << "  " << rc->id << " = " << ev << ",\n";
    } else {
      b << ctr ++ ;
      ev = b;
      if (d)
	aout << "  " << rc->id << " = " << ev << ",\n";
    }
    enum_tab.insert (rc->id, ev);
  }
  if (d) { 
    aout << "};\n";
    
    aout << "\ntemplate<class T> inline bool\n"
	 << "rpc_traverse (T &t, " << ct << " &obj)\n"
	 << "{\n"
	 << "  u_int32_t val = obj;\n"
	 << "  if (!rpc_traverse (t, val))\n"
	 << "    return false;\n"
	 << "  obj = " << ct << " (val);\n"
	 << "  return true;\n"
	 << "}\n\n";
  }
    
}

static void
dump_py_enum (const rpc_enum *e)
{
  str ct = pyc_type (e->id);
  aout << "struct " << ct << " {\n"
       << "  PyObject_HEAD\n"
       << "  " << e->id << " value;\n"
       << "};\n"
       << "RPC_ENUM_DECL (" << ct << ")\n"
       << "\n";
}

static void
dump_enum_dealloc_func (const rpc_enum *e)
{
  str ct = pyc_type (e->id);
  aout << "static void\n"
       << ct << "_dealloc (" << ct << " *self)\n"
       << "{\n"
       << "  PYDEBUG_PYFREE (" << ct << ");\n"
       << "  self->ob_type->tp_free ((PyObject *) self);\n"
       << "}\n\n";
}

static void
dump_int_to_enum_converter (const rpc_enum *e)
{
  str et = e->id;
  aout << "static bool\n"
       << et << "_int2enum (int in, " << et << " *out)\n"
       << "{\n"
       << "  " << et << " tmp = " << enum_cast (et, "in") << ";\n"
       << "  if (!" << et << "_is_in_range_set_error (tmp))\n"
       << "    return false;\n"
       << "  *out = tmp;\n"
       << "  return true;\n"
       << "}\n\n";
}

static void
dump_enum_init_func (const rpc_enum *e)
{
  str ct = pyc_type (e->id);
  aout << "static int\n"
       << ct << "_init (" << ct << " *self, PyObject *args, PyObject *k)\n"
       << "{\n"
       << "  int i = 0;\n"
       << "  if (!PyArg_ParseTuple (args, \"i\", &i))\n"
       << "    return -1;\n"
       << "  if (!" << e->id << "_int2enum (i, &self->value))\n"
       << "     return -1;\n"
       << "  return 0;\n"
       << "}\n\n";
}

static void
dump_enum_type_object (const str &id)
{
  const str ct = pyc_type (id);
  const str pt = id;
  aout << "PY_CLASS_DEF2 (" << ct << ", \"" << dotted_m << "." << pt << "\",\n"
       << "               1, dealloc, -1, \"" << pt << " object\",\n"
       << "               methods, members, 0, init, new, 0,\n"
       << "               str, 0, print);\n"
       << "\n";
}

static void
dump_enum_rpc_traverse (const str &id)
{
  aout << "template<class T> bool\n"
       << "rpc_traverse (T &t, " << pyc_type (id) << " &o)\n"
       << "{\n"
       << "  return rpc_traverse (t, o.value);\n"
       << "}\n\n";
}

static void
dumpenum_mthds (const rpc_enum *rs)
{
  dump_allocator (rs->id);
  dump_class_methods (rs->id);
}

static void
dumpenum_hdr (const rpc_sym *s)
{
  const rpc_enum *rs = s->senum.addr ();
  const str wt = pyw_type (rs->id);
  const str ct = pyc_type (rs->id);

  dump_c_enum (rs, true);
  dump_type_object_decl (rs->id);
  dump_py_enum (rs);
  dump_class_method_decls (ct);

  // dump the object wrapper
  dump_w_class (rs->id);
  dump_print_decls (rs->id);
  dump_type2str_decl (rs->id);

  // allow for pretty-printing when cross-linking XDR modules
  dump_rpc_print_decl (rs->id);
  dump_rpc_print_decl (wt);

  dump_enum_in_range_decl (rs);
  dump_w_enum_convert (rs->id);
  dump_w_rpc_traverse (rs->id);
  dump_enum_rpc_traverse (rs->id);

  // moved to header for time being
  dump_w_class_init_func (wt, py_type_obj (rs->id));
  dump_w_enum_clear_func (rs->id);
}


static void
dumpenum (const rpc_sym *s)
{
  const rpc_enum *rs = s->senum.addr ();
  const str wt = pyw_type (rs->id);
  const str ct = pyc_type (rs->id);

  dump_c_enum (rs, false);

  // convert used throughout
  dump_enum_in_range (rs);
  dump_int_to_enum_converter (rs);
  dump_enum_dealloc_func (rs);

  dump_enum_new_func (rs->id);
  // dump_enum_dealloc_func (rs);
  dump_class_members (ct);
  dump_class_methods_struct (ct);
  dump_enum_init_func (rs);
  dump_str_func (rs->id);
  dump_print_func (rs->id);
  dump_enum_type_object (rs->id);


  dump_xdr_func (rs->id);
  dumpenum_mthds (rs);
}

static void
dumptypedef_hdr (const rpc_sym *s)
{
  const rpc_decl *rd = s->stypedef.addr ();
  pdecl_py ("typedef ", rd);
  pmshl (rd->id);
  aout << "RPC_TYPEDEF_DECL (" << rd->id << ")\n";
}

static void
dumptypedef (const rpc_sym *s)
{
  const rpc_decl *rs = s->stypedef.addr ();
  dump_allocator (rs->id);
  dump_xdr_func (rs->id);
}

static void
mktbl (const rpc_program *rs)
{
  for (const rpc_vers *rv = rs->vers.base (); rv < rs->vers.lim (); rv++) {
    str name = rpcprog (rs, rv);
    aout << "static const rpcgen_table " << name << "_tbl[] = {\n"
	 << "  " << rs->id << "_" << rv->val << "_APPLY (XDRTBL_DECL)\n"
	 << "};\n"
	 << "const rpc_program " << name << " = {\n"
	 << "  " << rs->id << ", " << rv->id << ", " << name << "_tbl,\n"
	 << "  sizeof (" << name << "_tbl" << ") / sizeof ("
	 << name << "_tbl[0]),\n"
	 << "  \"" << name << "\"\n"
	 << "};\n\n";
  }
  aout << "\n";
}


static void 
dump_constants_ins1 (const str &k, const str &v)
{
  aout << "  if ((rc = PyModule_AddIntConstant (mod, \""
       << k << "\", (long ) (" << v << "))) < 0)\n"
       << "    return rc;\n";
}

static void 
dump_constants_ins1 (const str &k, u_int32_t v)
{
  strbuf b;
  b << v;
  dump_constants_ins1 (k, b);
}

static void
dump_constants_trav (const str &s, str *v)
{
  dump_constants_ins1 (s, *v);
}

static void
dump_constants_trav_i (const str &s, u_int32_t *v)
{
  dump_constants_ins1 (s, *v);
}

static void
dump_constants ()
{
  aout << "static int\n"
       << "py_module_all_ins (PyObject *mod)\n"
       << "{\n"
       << "  int rc = 0;\n";
  proc_tab.traverse (wrap (dump_constants_trav_i));
  enum_tab.traverse (wrap (dump_constants_trav));
  literal_tab.traverse (wrap (dump_constants_trav));
  aout << "  return rc;\n"
       << "}\n\n";
}

static void
collect_procnos (const rpc_program *rs)
{
  for (const rpc_vers *rv = rs->vers.base (); rv < rs->vers.lim (); rv++) {
    for (const rpc_proc *rp = rv->procs.base (); rp < rv->procs.lim (); rp++) {
      proc_tab.insert (rp->id, rp->val);
    }
  }
}

static void
dumpprog_hdr (const rpc_sym *s)
{
  aout << "\n";
  const rpc_program *rs = s->sprogram.addr ();
  for (const rpc_vers *rv = rs->vers.base (); rv < rs->vers.lim (); rv++) {
    str type = py_rpcprog_type (rs, rv);
    aout << "PY_CLASS_DECL (" << type << ");\n";
  }
  aout << "\n";
}

static void
dumpprog (const rpc_sym *s)
{
  const rpc_program *rs = s->sprogram.addr ();
  // aout << "\nenum { " << rs->id << " = " << rs->val << " };\n";
  aout << "#ifndef " << rs->id << "\n"
       << "#define " << rs->id << " " << rs->val << "\n"
       << "#endif /* !" << rs->id << " */\n";
  for (const rpc_vers *rv = rs->vers.base (); rv < rs->vers.lim (); rv++) {
    //aout << "extern const rpc_program " << rpcprog (rs, rv) << ";\n";
    aout << "enum { " << rv->id << " = " << rv->val << " };\n";
    aout << "enum {\n";
    for (const rpc_proc *rp = rv->procs.base (); rp < rv->procs.lim (); rp++)
      aout << "  " << rp->id << " = " << rp->val << ",\n";
    aout << "};\n";
    aout << "#define " << rs->id << "_" << rv->val
	 << "_APPLY_NOVOID(macro," << pyw_type ("void") << ")";
    u_int n = 0;
    for (const rpc_proc *rp = rv->procs.base (); rp < rv->procs.lim (); rp++) {
      while (n++ < rp->val)
	aout << " \\\n  macro (" << n-1 << ", false, false)";
      aout << " \\\n  macro (" << rp->id << ", " << pyw_type (rp->arg)
	   << ", " << pyw_type (rp->res) << ")";
    }
    aout << "\n";
    aout << "#define " << rs->id << "_" << rv->val << "_APPLY(macro) \\\n  "
	 << rs->id << "_" << rv->val << "_APPLY_NOVOID(macro, "
	 << pyw_type ("void") << ")\n";
  }

  // dump typechecker methods
  aout << "\n";
  for (const rpc_vers *rv = rs->vers.base (); rv < rs->vers.lim (); rv++) {
    u_int n = 0;
    str name = rpcprog (rs, rv);
    aout << "static const py_rpcgen_table_t " 
	 << py_rpcprog_extension (rs, rv) << "[] = {\n";
    for (const rpc_proc *rp = rv->procs.base (); rp < rv->procs.lim (); rp++) {
      while (n++ < rp->val) {
	aout << " py_rpcgen_error,\n";
	//aout << "  { convert_error, convert_error, wrap_error, "
	//    << "dealloc_error },\n";
      }
      aout << "  { py_wrap<" << pyw_type (rp->arg) << ">, "
	   << " py_wrap<" << pyw_type (rp->res) << "> },\n";
    }
    aout << "};\n\n";
  }

  mktbl (rs);
  collect_procnos (rs);
  dump_prog_py_obj (rs);
}

static void
dumphdr (const rpc_sym *s)
{
  switch (s->type) {
  case rpc_sym::CONST:
    aout << "enum { " << s->sconst->id
	 << " = " << s->sconst->val << " };\n";
    break;
  case rpc_sym::STRUCT:
    dumpstruct_hdr (s);
    break;
  case rpc_sym::UNION:
    dumpunion_hdr (s);
    break;
  case rpc_sym::ENUM:
    dumpenum_hdr (s);
    break;
  case rpc_sym::TYPEDEF:
    dumptypedef_hdr (s);
    break;
  case rpc_sym::PROGRAM:
    dumpprog_hdr (s);
    break;
  case rpc_sym::LITERAL:
    aout << *s->sliteral << "\n";
    break;
  default:
    break;
  }
}

static void
dumpsym (const rpc_sym *s, pass_num_t pass)
{
  switch (s->type) {
  case rpc_sym::STRUCT:
    if (pass == PASS_ONE)
      dumpstruct (s);
    else if (pass == PASS_THREE) 
      dumpstruct_mthds (s);
    break;
  case rpc_sym::UNION:
    if (pass == PASS_ONE)
      dumpunion (s);
    else if (pass == PASS_THREE)
      dumpunion_mthds (s);
    break;
  case rpc_sym::ENUM:
    if (pass == PASS_ONE)
      dumpenum (s);
    break;
  case rpc_sym::TYPEDEF:
    if (pass == PASS_ONE)
      dumptypedef (s);
    break;
  case rpc_sym::PROGRAM:
    if (pass == PASS_THREE)
      dumpprog (s);
    break;
  default:
    break;
  }
}


static str
makemodulename (str fname)
{
  strbuf guard;
  const char *p;
  static rxx x1 ("(.*?)_(so|lib).C");
  static rxx x2 ("(.*?).[a-zA-Z]+");
  static rxx x3 ("([a-zA-Z0-9_]+)");

  if ((p = strrchr (fname, '/')))
    p++;
  else p = fname;

  str r;

  if (x1.match (p))
    r = x1[1];
  else if (x2.match (p))
    r = x2[1];
  else if (x3.match (p))
    r = x3[1];
  else {
    warn << "Invalid output file name given; cannot convert it: " << p << "\n";
    exit (-1);
  }
  return r;
}

static vec<str>
get_c_classes (const rpc_sym *s)
{
  vec<str> ret;
  switch (s->type) {
  case rpc_sym::STRUCT:
    ret.push_back (s->sstruct.addr ()->id);
    break;
  case rpc_sym::UNION:
    ret.push_back (s->sunion.addr ()->id);
    break;
  case rpc_sym::PROGRAM:
    {
      const rpc_program *p = s->sprogram.addr ();
      for (const rpc_vers *rv = p->vers.base (); rv < p->vers.lim (); rv++) {
	ret.push_back (rpcprog (p, rv));
      }
    }
    break;
  case rpc_sym::ENUM:
    ret.push_back (s->senum.addr ()->id);
    break;
  default:
    break;
  }
  return ret;
}

static void
add_baseclass_to_rpc_programs (const symlist_t &lst)
{
  aout << "  // after import, we can fix up the rpc_program types..\n";
  for (const rpc_sym *s = lst.base (); s < lst.lim () ; s++) {
    if (s->type != rpc_sym::PROGRAM) 
      continue;
    vec<str> clss = get_c_classes (s);
    str cls;
    while (clss.size ()) {
      cls = clss.pop_back ();
      aout <<  "  " << py_type_obj (cls) << ".tp_base = py_rpc_program;\n";
    }
  }
}

static void
init_python_type_structures (const symlist_t &lst)
{
  aout << "\n"
       << "  if (";
  bool first = true;
  str cls;
  for (const rpc_sym *s = lst.base (); s < lst.lim () ; s++) {
    vec<str> clss = get_c_classes (s);
    str cls;
    while (clss.size ()) {
      cls = clss.pop_back ();
      if (first)
	first = false;
      else {
	aout << " ||\n      ";
      }
      aout << "PyType_Ready (&" << py_type_obj (cls) << ") < 0";
    }
  }
  aout << ")\n"
       << "    return;\n"
       << "\n"
    ;
}

static void
dump_init_func (const symlist_t &lst)
{
  strbuf b ("static_init_module_");
  b << module << "()";
  str func = b;

  aout << "INITFN(static_init_module_" << module << ");\n"
       << "static void\n"
       << func << "\n"
       << "{\n"
       << "  if (!import_sfs_exceptions (&AsyncXDR_Exception, NULL,\n"
       << "                              &AsyncUnion_Exception))\n"
       << "    return;\n"
       << "\n"
       << "  py_rpc_program = import_type (\"sfs.arpc\", \"rpc_program\");\n"
       << "  if (!py_rpc_program)\n"
       << "    return;\n"
       << "\n";

  add_baseclass_to_rpc_programs (lst);
  aout << "\n";
  init_python_type_structures (lst);

  for (const rpc_sym *s = lst.base (); s < lst.lim () ; s++) {
    vec<str> clss = get_c_classes (s);
    str cls;
    while (clss.size ()) {
      cls = clss.pop_back ();
      aout << "  Py_INCREF (&" << py_type_obj (cls) << ");\n";
    }
  }

  aout << "}\n"
       << "\n";
}

static void
dumpmodule (const symlist_t &lst)
{
  aout << "static PyMethodDef module_methods[] = {\n"
       << "  {NULL}\n"
       << "};\n\n"
       << "#ifndef PyMODINIT_FUNC	"
       << "/* declarations for DLL import/export */\n"
       << "#define PyMODINIT_FUNC void\n"
       << "#endif\n"
       << "PyMODINIT_FUNC\n"
       << "init" << module << " (void)\n"
       << "{\n"
       << "  PyObject* m;\n"
       << "\n";



  aout << "\n"
       << "  m = Py_InitModule3 (\"" << module << "\", module_methods,\n"
       << "                      \"Python/rpc/XDR module for " 
       << module << ".\");\n"
       << "\n"
       << "  if (m == NULL)\n"
       << "    return;\n"
       << "  if (py_module_all_ins (m) < 0)\n" 
       << "    return;\n"
       << "\n";
  
  for (const rpc_sym *s = lst.base (); s < lst.lim () ; s++) {
    vec<str> clss = get_c_classes (s);
    str cls;
    while (clss.size ()) {
      cls = clss.pop_back ();
      aout << "  PyModule_AddObject (m, \"" << cls
	   << "\", (PyObject *)&" << py_type_obj (cls) << ");\n";
    }
  }
  
  aout << "}\n"
       << "\n";
}
static str
makehdrname (str fname)
{
  static rxx x ("(.*?)([^/]+)_(lib|so).C");
  strbuf hdr;
  const char *p;

  if (!x.match (fname)) {
    // old-style translation
    if ((p = strrchr (fname, '/')))
      p++;
    else p = fname;
    
    hdr.buf (p, strlen (p) - 1);
    hdr.cat ("h");
  } else {
    hdr << x[2] << ".h";
  }
  return hdr;

} 

static void
init_globals (const str &fname)
{
  module = makemodulename (fname);
  dotted_m = python_module_name ? python_module_name : module;
}


void
genpyc_lib (str fname)
{
  init_globals (fname);

  aout << "// -*-c++-*-\n"
       << "/* This file was automatically generated by rpcc. */\n\n"
       << "\n"
       << "#include \"" << makehdrname (fname) << "\"\n\n"
       << "static PyObject *AsyncXDR_Exception;\n"
       << "static PyObject *AsyncUnion_Exception;\n"
       << "static PyTypeObject *py_rpc_program;\n"
       << "\n";

  int last = rpc_sym::LITERAL;
  
  // 3 - pass system to get dependencies / orders right
  for (pass_num_t i = PASS_ONE; i < N_PASSES ; 
       i = (pass_num_t) ((int )i + 1)) {
    for (const rpc_sym *s = symlist.base (); s < symlist.lim (); s++) {
      if (last != s->type
	  || last == rpc_sym::PROGRAM
	  || last == rpc_sym::TYPEDEF
	  || last == rpc_sym::STRUCT
	  || last == rpc_sym::UNION
	  || last == rpc_sym::ENUM)
	aout << "\n";
      last = s->type;
      switch (i) {
      case PASS_ONE:
      case PASS_THREE:
	dumpsym (s, i);
	break;
      case PASS_TWO:
	dumpprint (s);
	break;
      default:
	break;
      }
    }
  }
  dump_init_func (symlist);
}

void
genpyc_so (str fname)
{
  init_globals (fname);
  static rxx define_rxx ("#\\s*define\\s*([a-zA-Z0-9_]+)\\s*([a-zA-Z0-9_]+)");

  aout << "// -*-c++-*-\n"
       << "/* This file was automatically generated by rpcc. */\n\n"
       << "\n"
       << "#include \"" << makehdrname (fname) << "\"\n\n"
    ;

  // collect all of the needed symbols
  for (const rpc_sym *s = symlist.base (); s < symlist.lim (); s++) {
    switch (s->type) {
    case rpc_sym::ENUM:
      dump_c_enum (s->senum.addr (), false);
      break;
    case rpc_sym::PROGRAM:
      collect_procnos (s->sprogram.addr ());
      break;
    case rpc_sym::LITERAL:
      if (define_rxx.match (*s->sliteral)) {
	literal_tab.insert (define_rxx[1], define_rxx[2]);
      }
      break;
    default:
      break;
    }
  }

  dump_constants ();
  dumpmodule (symlist);
}

static str
makeguard (str fname)
{
  strbuf guard;
  const char *p;

  if ((p = strrchr (fname, '/')))
    p++;
  else p = fname;

  guard << "__PY_RPCC_";
  while (char c = *p++) {
    if (isalnum (c))
      c = toupper (c);
    else
      c = '_';
    guard << c;
  }
  guard << "_INCLUDED__";

  return guard;
}

void
genpyh (str fname)
{
  str guard = makeguard (fname);

  aout << "// -*-c++-*-\n"
       << "/* This file was automatically generated by rpcc. */\n\n"
       << "#include \"py_rpctypes.h\"\n"
       << "#include \"py_gen.h\"\n"
       << "#include \"py_util.h\"\n"
       << "#include \"xdrmisc.h\"\n"
       << "#include \"crypt.h\"\n"
       << "#ifndef " << guard << "\n"
       << "#define " << guard << " 1\n\n";

  int last = rpc_sym::LITERAL;
  for (const rpc_sym *s = symlist.base (); s < symlist.lim (); s++) {
    if (last != s->type
	|| last == rpc_sym::PROGRAM
	|| last == rpc_sym::TYPEDEF
	|| last == rpc_sym::STRUCT
	|| last == rpc_sym::UNION
	|| last == rpc_sym::ENUM)
      aout << "\n";
    last = s->type;
    dumphdr (s);
  }

  aout << "#endif /* !" << guard << " */\n";
}


