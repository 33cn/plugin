// -*-c++-*-

#ifndef __PY_RPCTYPES_H_INCLUDED__
#define __PY_RPCTYPES_H_INCLUDED__

#include <Python.h>
#include "structmember.h"
#include "arpc.h"
#include "rpctypes.h"
#include "qhash.h"
#include "async.h"
#include "py_debug.h"


//-----------------------------------------------------------------------
//
//  Native Object types for RPC ptrs
//

struct py_rpc_ptr_t {
  PyObject_HEAD
  PyTypeObject *typ;
  PyObject *p;
  PyObject *alloc ();
  void clear ();
  operator bool () const { return p ; }
  void set_typ (PyTypeObject *o);
};
extern PyTypeObject py_rpc_ptr_t_Type;

//
//-----------------------------------------------------------------------

// 
// Here are the classes that we are going to be using to make Python
// wrappers for C++/Async style XDR structures:
//
//   py_u_int32_t 
//   py_int32_t
//   py_u_int64_t
//   py_int64_t
//
//   - note these classes are normal C++ classes that wrap
//     a python PyObject *, in which the data is actually
//     stored.  We need this wrapper ability for when we
//     reassign new values to the underlying C++ data; we
//     cannot afford to reassign the whole object then, and
//     therefore need to wrap them.
// 
//   py_rpc_str<n>
//
//    - as above, but parameterized by n, a max size parameter.
//
// for each one of these types, and also the custom types, we
// define the following functions, as required by the XDR 
// interface.
//
//   void * T_alloc ()
//    - allocate an object of the given variety
//   bool xdr_T (XDR *x, void *objp)
//   template<class X> bool rpc_traverse (X &x, T &t)
//    - standard from DM's XDR library.
//

//-----------------------------------------------------------------------
// Wrapper Class Definitions
//

typedef enum { PYW_ERR_NONE = 0,
	       PYW_ERR_TYPE = 1,
	       PYW_ERR_BOUNDS = 2,
	       PYW_ERR_UNBOUND = 3 } pyw_err_t;

class pyw_base_err_t {
public:
  pyw_base_err_t () : _err (PYW_ERR_NONE) {}
  pyw_base_err_t (pyw_err_t e) : _err (e) {}
  bool print_err (const strbuf &b, int recdepth,
		  const char *name, const str &prfx) const ;
  void set_err (pyw_err_t e) { _err = e; }
protected:
  pyw_err_t _err;
};

class pyw_base_t : public pyw_base_err_t {
public:
  
  pyw_base_t () : 
    pyw_base_err_t (), _obj (NULL), _typ (NULL) {}
  pyw_base_t (PyTypeObject *t) : 
    pyw_base_err_t (), _obj (NULL), _typ (t) {}
  pyw_base_t (PyObject *o, PyTypeObject *t) : 
    pyw_base_err_t (), _obj (o), _typ (t) {}
  pyw_base_t (pyw_err_t e, PyTypeObject *t) : 
    pyw_base_err_t (e), _obj (NULL), _typ (t) {}
  pyw_base_t (const pyw_base_t &p) :
    pyw_base_err_t (p._err), _obj (p._obj), _typ (p._typ) 
  { Py_XINCREF (_obj); }
  
  PYDEBUG_VIRTUAL_DESTRUCTOR ~pyw_base_t () { Py_XDECREF (_obj); }

  // sometimes we allocate room for objects via Python's memory
  // allocation routines.  Doing so does not do "smart" initialization,
  // as C++ would do when you call "new."  We can mimic this behavior,
  // though, with a call to tp_alloc (), and then by calling the init()
  // functions.  Note that the init functions should not at all access
  // memory, since memory will be garbage.
  bool init (); 

  bool alloc ();
  bool clear ();
  PyObject *obj () { return _obj; }

  // takes a borrowed reference, and INCrefs
  bool set_obj (PyObject *o);

  PyObject *get_obj ();
  PyObject *safe_get_obj (bool fail = false);
  const PyObject *const_obj () const { return _obj; }
  PyObject *unwrap () ;

protected:
  PyObject *_obj;
  PyTypeObject *_typ;

};

template<class W, class P>
class pyw_tmpl_t : public pyw_base_t {
public:
  pyw_tmpl_t (PyTypeObject *t) : pyw_base_t (t) 
  { PYDEBUG_NEW (W) }
  pyw_tmpl_t (PyObject *o, PyTypeObject *t) : pyw_base_t (o, t) 
  { PYDEBUG_NEW (W) }
  pyw_tmpl_t (const pyw_tmpl_t<W,P> &t) : pyw_base_t (t) 
  { PYDEBUG_NEW (W) }
  pyw_tmpl_t (pyw_err_t e, PyTypeObject *t) : pyw_base_t (e, t) 
  { PYDEBUG_NEW (W) }

  ~pyw_tmpl_t () { PYDEBUG_DEL (W) }
  
  P *casted_obj () { return reinterpret_cast<P *> (_obj); }
  const P *const_casted_obj () const 
  { return reinterpret_cast<const P *> (_obj); }

  // aborted, but we'll keep it in here anyways.
  bool copy_from (const pyw_tmpl_t<W,P> &t);

  bool safe_set_obj (PyObject *in);

};

template<class W, size_t m = RPC_INFINITY>
class pyw_tmpl_str_t : public pyw_tmpl_t<W, PyStringObject> {
public:
  pyw_tmpl_str_t () :
    pyw_tmpl_t<W, PyStringObject> (&PyString_Type)  {}
  pyw_tmpl_str_t (PyObject *o) :
    pyw_tmpl_t<W, PyStringObject> (o, &PyString_Type) {}
  pyw_tmpl_str_t (pyw_err_t e) :
    pyw_tmpl_t<W, PyStringObject> (e, &PyString_Type) {}
  pyw_tmpl_str_t (const pyw_tmpl_str_t<W, m> &o) :
    pyw_tmpl_t<W, PyStringObject> (o) {}

  char *get (size_t *sz) const;
  const char * get () const ;
  bool set (char *buf, size_t len);
  bool init ();
  enum { maxsize = m };
};

class pyw_rpc_byte_t : public pyw_base_err_t 
{
public:
  pyw_rpc_byte_t (pyw_err_t e) : pyw_base_err_t (e), _byt (0) {}
  inline void set_byte (char c) { _byt = c; }
  inline char get_byte () const { return _byt; }
private:
  char _byt;
};

template<class W, size_t m = RPC_INFINITY>
class pyw_tmpl_opq_t : public pyw_tmpl_str_t<W,m> 
{
public:
  pyw_tmpl_opq_t () : pyw_tmpl_str_t<W,m> () {}
  pyw_tmpl_opq_t (PyObject *o) : pyw_tmpl_str_t<W,m> (o) {}
  pyw_tmpl_opq_t (pyw_err_t e) : pyw_tmpl_str_t<W,m> (e) {}
  pyw_tmpl_opq_t (const pyw_tmpl_opq_t<W,m> &o) : pyw_tmpl_str_t<W,m> (o){}

  bool init_trav () const;
  bool init ();
  size_t size () const { return _sz; }
protected:
  mutable const char *_str;
  mutable size_t _sz;
};

template<size_t M = RPC_INFINITY>
class pyw_rpc_str : public pyw_tmpl_str_t<pyw_rpc_str<M>, M >
{
public:
  pyw_rpc_str () : pyw_tmpl_str_t<pyw_rpc_str<M>, M > () {}
  pyw_rpc_str (PyObject *o) : pyw_tmpl_str_t<pyw_rpc_str<M>, M > (o) {}
  pyw_rpc_str (pyw_err_t e) : pyw_tmpl_str_t<pyw_rpc_str<M>, M > (e) {}
  pyw_rpc_str (const pyw_rpc_str<M> &o) 
    : pyw_tmpl_str_t<pyw_rpc_str<M>, M > (o) {}
};

template<size_t M = RPC_INFINITY>
class pyw_rpc_bytes : public pyw_tmpl_opq_t<pyw_rpc_bytes<M>, M>
{
public:
  pyw_rpc_bytes () : pyw_tmpl_opq_t<pyw_rpc_bytes<M>, M > () {}
  pyw_rpc_bytes (PyObject *o) : pyw_tmpl_opq_t<pyw_rpc_bytes<M>, M > (o) {}
  pyw_rpc_bytes (pyw_err_t e) : pyw_tmpl_opq_t<pyw_rpc_bytes<M>, M > (e) {}
  pyw_rpc_bytes (const pyw_rpc_bytes<M> &o) 
    : pyw_tmpl_opq_t<pyw_rpc_bytes<M>, M > (o) {}

  bool get_char (u_int i, pyw_rpc_byte_t *c) const;
  pyw_rpc_byte_t operator[] (u_int i) const;
};

template<size_t M> 
class pyw_rpc_opaque : public pyw_tmpl_opq_t<pyw_rpc_opaque<M>, M>
{
public:
  pyw_rpc_opaque () : pyw_tmpl_opq_t<pyw_rpc_opaque<M>, M > () {}
  pyw_rpc_opaque (PyObject *o) : pyw_tmpl_opq_t<pyw_rpc_opaque<M>, M > (o) {}
  pyw_rpc_opaque (pyw_err_t e) : pyw_tmpl_opq_t<pyw_rpc_opaque<M>, M > (e) {}
  pyw_rpc_opaque (const pyw_rpc_opaque<M> &o) 
    : pyw_tmpl_opq_t<pyw_rpc_opaque<M>, M > (o) {}
  bool get_char (u_int i, pyw_rpc_byte_t *c) const;
  pyw_rpc_byte_t operator[] (u_int i) const;
};

template<class T, size_t sz>
class pyw_rpc_array : public pyw_tmpl_t<pyw_rpc_array<T,sz>, PyListObject >
{
public:
  pyw_rpc_array () :
    pyw_tmpl_t<pyw_rpc_array<T,sz>, PyListObject > (&PyList_Type) {}
  pyw_rpc_array (PyObject *o) :
    pyw_tmpl_t<pyw_rpc_array<T,sz>, PyListObject > (o) {}
  pyw_rpc_array (pyw_err_t e) :
    pyw_tmpl_t<pyw_rpc_array<T,sz>, PyListObject > (e, &PyList_Type) {}
  pyw_rpc_array (const pyw_rpc_array<T,sz> &p) :
    pyw_tmpl_t<pyw_rpc_array<T,sz>, PyListObject > (p) {}

  T operator[] (u_int i) const;
  u_int size () const { return n_elem; }
  bool get_slot (u_int i, T *out) const;
  bool set_slot (PyObject *el, u_int i);
  bool init ();

  enum { n_elem = sz } ;

};

template<class T, size_t max>
class pyw_rpc_vec : public pyw_tmpl_t<pyw_rpc_vec<T, max>, PyListObject > 
{
public:
  pyw_rpc_vec () : 
    pyw_tmpl_t<pyw_rpc_vec<T, max>, PyListObject > (&PyList_Type), _sz (0) {}
  pyw_rpc_vec (PyObject *o) : 
    pyw_tmpl_t<pyw_rpc_vec<T, max>, PyListObject > (o, &PyList_Type), _sz (0) 
  {}
  pyw_rpc_vec (pyw_err_t e) : 
    pyw_tmpl_t<pyw_rpc_vec<T, max>, PyListObject > (e, &PyList_Type), _sz (0) 
  {}
  pyw_rpc_vec (const pyw_rpc_vec<T,max> &p) : 
    pyw_tmpl_t<pyw_rpc_vec<T, max>, PyListObject > (p), _sz (0) {}
    
  T operator[] (u_int i) const;
  u_int size () const;
  bool get_slot (u_int i, T *out) const;
  bool set_slot (PyObject *el, u_int i);
  bool shrink (size_t sz);
  bool init ();
  enum { maxsize = max };
private:
  mutable size_t _sz;
};

class pyw_void : public pyw_tmpl_t<pyw_void, PyObject>
{
public:
  pyw_void () : pyw_tmpl_t<pyw_void, PyObject> (NULL)
  {
    _obj = Py_None;
    Py_INCREF (_obj);
  }
  pyw_void (PyObject *o) : pyw_tmpl_t<pyw_void, PyObject> (o, NULL) {}
  bool init () { return true; }

};

template<class T, T *typ>
class pyw_enum_t : public pyw_tmpl_t<pyw_enum_t<T, typ>, T>
{
public:
  pyw_enum_t () : pyw_tmpl_t<pyw_enum_t<T, typ>, T> (typ) {}
  pyw_enum_t (PyObject *o) : pyw_tmpl_t<pyw_enum_t<T, typ>, T> (o, typ) {}
  pyw_enum_t (const pyw_enum_t<T, typ> &p) :
    pyw_tmpl_t<pyw_enum_t<T, typ>, T> (p) {}
};

template<class T, PyTypeObject *t>
class pyw_rpc_ptr : public pyw_tmpl_t<pyw_rpc_ptr<T,t>, py_rpc_ptr_t>
{
public:
  pyw_rpc_ptr () 
    : pyw_tmpl_t<pyw_rpc_ptr<T,t>, py_rpc_ptr_t> (&py_rpc_ptr_t_Type) {}
  pyw_rpc_ptr (PyObject *o) 
    : pyw_tmpl_t<pyw_rpc_ptr<T,t>, py_rpc_ptr_t> (o, &py_rpc_ptr_t_Type) {}
  pyw_rpc_ptr (const pyw_rpc_ptr<T,t> &p) :
    pyw_tmpl_t<pyw_rpc_ptr<T,t>, py_rpc_ptr_t> (p) {}
  bool init ();
  T * palloc ();
  operator bool() const 
  { return const_casted_obj () && const_casted_obj ()->p ; }
};



//-----------------------------------------------------------------------
// py_rpcgen_table_t
//
//   we need to add just one more method to the basic rpcgen_table,
//   which is for wrapping generic Python objects in their appropriate
//   typed C++ wrappers.
//  

typedef pyw_base_t * (*wrap_fn_t) (PyObject *, PyObject *);

struct py_rpcgen_table_t {
  wrap_fn_t wrap_arg;
  wrap_fn_t wrap_res;
};

extern py_rpcgen_table_t py_rpcgen_error;

//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
// 
// Allocate / Deallocate / Decref / init / clear
//
// 

// for native types, we just call into the class, but compiled
// types don't have methods in their structs, so they will simply
// specialize this template
template<class T> inline bool py_init (T &t) { return t.init (); }
template<class T> inline bool py_clear (T &t) { return t.clear (); }

inline void *
pyw_rpc_str_alloc ()
{
  return New pyw_rpc_str<RPC_INFINITY> ();
}

inline void *
pyw_rpc_bytes_alloc ()
{
  return New pyw_rpc_bytes<RPC_INFINITY> ();
}

template<class W, size_t m> bool
pyw_tmpl_str_t<W,m>::init ()
{
  if (!pyw_tmpl_t<W, PyStringObject >::init ())
    return false;
  _typ = &PyString_Type;
  return true;
}

template<class W, size_t m> bool
pyw_tmpl_opq_t<W,m>::init ()
{
  if (!pyw_tmpl_str_t<W, m>::init ())
    return false;
  _str = NULL;
  _sz = 0;
  return true;
}

template<class T, size_t sz> bool
pyw_rpc_array<T,sz>::init ()
{
  if (!pyw_tmpl_t<pyw_rpc_array<T,sz >, PyListObject >::init ())
    return false;
  _typ = &PyList_Type;
  PyObject *l = PyList_New (n_elem);
  if (!l) {
    PyErr_NoMemory ();
    return false;
  }
  for (u_int i = 0; i < size () ; i++) {
    T *t = New T ;
    PyList_SET_ITEM (l, i, t->unwrap ());
  }
  bool rc = set_obj (l); 
  Py_DECREF (l);
  return rc;
}

template<class T, size_t M> bool
pyw_rpc_vec<T,M>::init ()
{
  if (!pyw_tmpl_t<pyw_rpc_vec<T,M >, PyListObject >::init ())
    return false;
  _typ = &PyList_Type;
  PyObject *l = PyList_New (0);
  if (!l) {
    PyErr_SetString (PyExc_MemoryError, "allocation of new list failed");
    return false;
  }

  bool rc = set_obj (l);
  // calling set_obj will increase the reference count on l, so we
  // should DECREF is again here
  Py_DECREF (l);
  return rc;
}

template<class T, PyTypeObject *t> T *
pyw_rpc_ptr<T,t>::palloc ()
{
  assert (_obj);
  return reinterpret_cast<T *> (casted_obj ()->alloc ());
}

template<class T, PyTypeObject *t> bool
pyw_rpc_ptr<T,t>::init ()
{
  if (!pyw_tmpl_t<pyw_rpc_ptr<T,t>, py_rpc_ptr_t>::init ())
    return false;
  _typ = &py_rpc_ptr_t_Type; // note, not type t!!!
  if (!alloc ())
    return false;
  casted_obj ()->set_typ (t);   // this is where we use t !
  return true;
}

#define ALLOC_DECL(T)                                           \
extern void * T##_alloc ();

ALLOC_DECL(pyw_void);

//
//
//-----------------------------------------------------------------------


//-----------------------------------------------------------------------
// XDR wrapper functions
// 

#define XDR_DECL(T)                                            \
extern BOOL xdr_##T (XDR *xdrs, void *objp);

template <class T> bool
xdr_doit (XDR *xdrs, void *objp)
{
  switch (xdrs->x_op) {
  case XDR_DECODE:
  case XDR_ENCODE:
    return rpc_traverse (xdrs, *static_cast<T *> (objp));
  case XDR_FREE:
    rpc_destruct (static_cast<T *> (objp));
    return true;
  default:
    panic ("unexpected XDR op: %d\n", xdrs->x_op);
  }
}

template<size_t n> inline bool
xdr_pyw_rpc_str (XDR *xdrs, void *objp)
{
  return xdr_doit<pyw_rpc_str<n> > (xdrs, pbj);
}

template<class W, size_t n> inline bool
xdr_pyw_rpc_bytes (XDR *xdrs, void *objp)
{
  return xdr_doit<pyw_tmpl_opq_t<W,n> *> (xdrs, objp);
}

template<class T, size_t n> inline bool
xdr_pyw_rpc_vec (XDR *xdrs, void *objp)
{
  return xdr_doit<pyw_rpc_vec<T,n> *> (xdrs, objp);
}

template<class T, size_t n> inline bool
xdr_pyw_rpc_array (XDR *xdrs, void *objp)
{
  return xdr_doit<pyw_rpc_array<T,n> *> (xdrs, obj);
}

template<class T, PyTypeObject *t> inline bool
xdr_pyw_rpc_ptr (XDR *xdr, void *objp)
{
  return xdr_doit<pyw_rpc_ptr<T,t> *> (objp);
}

//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
// RPC Printing 
//

#define PY_RPC_TYPE2STR_DECL(T)			\
template<> struct rpc_type2str<pyw_##T> {	\
  static const char *type () { return #T; }	\
};

#define NOT_A_STRUCT(T)                         \
inline bool rpc_isstruct (T) { return false; }

NOT_A_STRUCT(pyw_rpc_byte_t);

template<size_t m> struct rpc_type2str<pyw_rpc_bytes<m> > {
  static const char *type () { return "opaque"; }
};

template<size_t n> struct rpc_namedecl<pyw_rpc_bytes<n> > {
  static str decl (const char *name) {
    return rpc_namedecl<rpc_bytes<n> >::decl (name);
  }
};
template<size_t n> struct rpc_namedecl<pyw_rpc_opaque<n> > {
  static str decl (const char *name) {
    return rpc_namedecl<rpc_opaque<n> >::decl (name);
  }
};


template<size_t n> const strbuf &
rpc_print (const strbuf &sb, const pyw_rpc_str<n> &pyobj,
	   int recdepth = RPC_INFINITY,
	   const char *name = NULL, const char *prefix = NULL)
{
  const char *obj = pyobj.get ();
  if (!obj) obj = "<NULL>";
  if (pyobj.print_err (sb, recdepth, name, prefix)) 
    return sb;

  if (prefix)
    sb << prefix;

  if (name)
    sb << rpc_namedecl<rpc_str<n> >::decl (name) << " = ";
  if (obj)
    sb << "\"" << obj << "\"";	// XXX should map " to \" in string
  else
    sb << "NULL";
  if (prefix)
    sb << ";\n";
  return sb;
}


template<class T, PyTypeObject *t> const strbuf &
rpc_print (const strbuf &sb, const pyw_rpc_ptr<T,t> &obj,
	   int recdepth = RPC_INFINITY,
	   const char *name = NULL, const char *prefix = NULL)
{
  if (name) {
    if (prefix)
      sb << prefix;
    sb << rpc_namedecl<pyw_rpc_ptr<T,t> >::decl (name) << " = ";
  }
  if (!obj)
    sb << "NULL;\n";
  else if (!recdepth)
    sb << "...\n";
  else {
    sb << "&";
    rpc_print (sb, * reinterpret_cast<T *> (obj.const_casted_obj ()->p), 
	       recdepth - 1, NULL, prefix);
  }
  return sb;
}


template<class L, class E> E
obj_at (const L &lst, u_int i)
{
  if (i >= lst.size ())
    return E (PYW_ERR_BOUNDS);
  E ret (PYW_ERR_NONE);
  if (!lst.get_slot (i, &ret))
    return E (PYW_ERR_TYPE);
  return ret;
}

#define OBJ_AT(typ)                                              \
template<class T, size_t m> T                                    \
pyw_rpc_##typ <T,m>::operator[] (u_int i) const                  \
{                                                                \
  return obj_at<pyw_rpc_##typ <T,m>, T> (*this, i);             \
}

OBJ_AT (vec);
OBJ_AT (array);

template<class W> pyw_rpc_byte_t 
char_at (const W &obj, u_int i) 
{
  if (i >= obj.size ())
    return pyw_rpc_byte_t (PYW_ERR_BOUNDS);
  pyw_rpc_byte_t ret (PYW_ERR_NONE);
  obj.get_char (i, &ret);
  return ret;
}
const strbuf &
rpc_print (const strbuf &sb, const pyw_rpc_byte_t &obj,
	   int recdepth = RPC_INFINITY,
	   const char *name = NULL, const char *prefix = NULL);
 
template<size_t m> pyw_rpc_byte_t
pyw_rpc_bytes<m>::operator[] (u_int i) const { return char_at (*this, i); }
template<size_t m> pyw_rpc_byte_t
pyw_rpc_opaque<m>::operator[] (u_int i) const { return char_at (*this, i); }


#define RPC_ARRAYVEC_DECL(TEMP)                                 \
template<class T, size_t n> const strbuf &			\
rpc_print (const strbuf &sb, const TEMP<T, n> &obj,		\
	   int recdepth = RPC_INFINITY,				\
	   const char *name = NULL, const char *prefix = NULL)	\
{								\
  if (obj.print_err (sb, recdepth, name, prefix))               \
    return sb;                                                  \
  return rpc_print_array_vec (sb, obj, recdepth, name, prefix);	\
}

RPC_ARRAYVEC_DECL(pyw_rpc_vec);
RPC_ARRAYVEC_DECL(pyw_rpc_array);

#define RPC_ARRAYVEC_OPQ_DECL(T)                                \
template<size_t n> const strbuf &                               \
rpc_print (const strbuf &sb, const T<n> &obj,                   \
	   int recdepth = RPC_INFINITY,                         \
	   const char *name = NULL, const char *prefix = NULL)  \
{                                                               \
  if (obj.print_err (sb, recdepth, name, prefix)) return sb;    \
  if (!obj.init_trav ()) return sb;                             \
  return rpc_print_array_vec (sb, obj, recdepth, name, prefix); \
}

RPC_ARRAYVEC_OPQ_DECL(pyw_rpc_bytes);
RPC_ARRAYVEC_OPQ_DECL(pyw_rpc_opaque);

template<class T, size_t n> struct rpc_namedecl<pyw_rpc_vec<T, n> > {
  static str decl (const char *name) {
    return rpc_namedecl<rpc_vec<T, n> >::decl (name);
  }
};
template<class T, size_t n> struct rpc_namedecl<pyw_rpc_array<T,n> > {
  static str decl (const char *name) {
    return rpc_namedecl<array<T, n> >::decl (name);
  }
};
template<class T, PyTypeObject *t> struct rpc_namedecl<pyw_rpc_ptr<T,t> > {
  static str decl (const char *name) {
    return rpc_namedecl<T>::decl (str (strbuf () << "*" << name));
  }
};

#define PY_RPC_PRINT_GEN(T, expr)				\
const strbuf &							\
rpc_print (const strbuf &sb, const T &obj, int recdepth,	\
	   const char *name, const char *prefix)		\
{								\
  if (obj.print_err (sb, recdepth, name, prefix))               \
    return sb;                                                  \
  if (name) {							\
    if (prefix)							\
      sb << prefix;						\
    sb << rpc_namedecl<T >::decl (name) << " = ";		\
  }								\
  expr;								\
  if (prefix)							\
    sb << ";\n";						\
  return sb;							\
}

#define PY_XDR_OBJ_WARN(T,w)                                    \
PyObject *                                                      \
T##_##w (T *self)                                               \
{                                                               \
   dump_to (self, w);                                           \
   Py_INCREF (Py_None);                                         \
   PyErr_Clear ();                                              \
   return Py_None;                                              \
}

#define PY_XDR_OBJ_WARN_DECL(T,w)                               \
PyObject * T##_##w (T *self) ;

template<class T, class S> void
dump_to (const T *obj, S s)
{
  rpc_print (s, *obj);
}


//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
// operators
//   - multipurpose classes the hold lots of specialized template calls
//     for the particular class
//


//-----------------------------------------------------------------------
// convert
//   - a python type into its C equivalent.  For compiled, complex
//     types, this will do a type check, and return the arg passed.
//     for wrapped objects, it will allocate the wrapper, and 
//     add the object into the wrapper
//

template<class W> struct converter_t {};

template<class W> pyw_base_t *
py_wrap (PyObject *in, PyObject *dummy)
{
  PyObject *out = converter_t<W>::convert (in);
  if (!out) return NULL;
  W * ret = New W (out);
  if (!ret) {
    PyErr_SetString (PyExc_MemoryError, "out of memory in py_wrap");
    return NULL;
  }
  return ret;
}

template<size_t m> 
struct converter_t<pyw_rpc_str<m> >
{
  static PyObject * convert (PyObject *in)
  {
    if (!PyString_Check (in)) {
      PyErr_SetString (PyExc_TypeError, "expected string type");
      return NULL;
    }
    char *dat = PyString_AsString (in);
    if (!dat) {
      PyErr_SetString (PyExc_RuntimeError, "AsString failed");
      return NULL;
    }
    size_t len = strlen (dat);
    if (len > m) {
      PyErr_SetString (PyExc_OverflowError, 
		       "string exceeds predeclared limits");
      return NULL;
    }
    PyObject *out = PyString_FromString (dat); // will cut off at first '\0'
    if (!out) 
      return PyErr_NoMemory ();

    return out;
  }
};

template<size_t m>
struct converter_t<pyw_rpc_opaque<m> >
{
  static PyObject * convert (PyObject *in)
  {
    char buf[m];
    memset (buf, 0, m);
    if (!PyString_Check (in)) {
      PyErr_SetString (PyExc_TypeError, "expected string type");
      return NULL;
    }
    char *dat;
    int sz;
    if (PyString_AsStringAndSize (in, &dat, &sz) < 0) {
      PyErr_SetString (PyExc_RuntimeError, "AsStringAndSize failed");
      return NULL;
    }
    if (sz > int (m)) {
      PyErr_SetString (PyExc_OverflowError, 
		       "bytes exceed predeclared limits");
      return NULL;
    }
    memcpy (buf, dat, sz);
    return PyString_FromStringAndSize (buf, m);
  }
};

template<size_t m>
struct converter_t<pyw_rpc_bytes<m> >
{
  static PyObject *convert (PyObject *in)
  {
    if (!PyString_Check (in)) {
      PyErr_SetString (PyExc_TypeError, "expected string type");
      return NULL;
    }
    if (PyString_Size (in) > int (m)) {
      PyErr_SetString (PyExc_OverflowError, 
		       "bytes exceed predeclared limits");
      return NULL;
    }
    Py_INCREF (in);
    return in;
  }

};

template<class T, size_t m> struct converter_t<pyw_rpc_vec<T,m> >
{
  static PyObject * convert (PyObject *in)
  {
    if (!PyList_Check (in)) {
      PyErr_SetString (PyExc_TypeError, "expected a list type");
      return NULL;
    }
    if (PyList_Size (in) > int (m)) {
      PyErr_SetString (PyExc_OverflowError, 
		       "list execeeds predeclared list length");
      return NULL;
    }
    
    // XXX
    // Would be nice to typecheck here; maybe use RPC traverse?

    Py_INCREF (in);
    return in;
  }
};

template<class T, size_t m> struct converter_t<pyw_rpc_array<T,m> >
{
  static PyObject * convert (PyObject *in)
  {
    if (!PyList_Check (in)) {
      PyErr_SetString (PyExc_TypeError, "expected a list type");
      return NULL;
    }

    if (PyList_Size (in) > int (m)) {
      PyErr_SetString (PyExc_OverflowError, 
		       "list execeeds predeclared list length");
      return NULL;
    }

    while (PyList_Size (in) < int (m)) {
      T *n = New T;
      // Note: unwrap will dealloc the wrapper object
      if (PyList_Append (in, n->unwrap ()) < 0) 
	return NULL;
    }

    // might be good to do a recursive type check...
    Py_INCREF (in);
    return in;
  }
};
    


template<class T, PyTypeObject *t> struct converter_t<pyw_rpc_ptr<T, t> >
{
  static PyObject * convert (PyObject *in)
  {
    py_rpc_ptr_t *ret = NULL;
    if (PyObject_IsInstance (in, (PyObject *)t)) {
      if ((ret = PyObject_New (py_rpc_ptr_t, &py_rpc_ptr_t_Type))) {
	Py_INCREF (in);
	ret->set_typ (t);
	ret->p = in;
      }
    } else if (PyObject_IsInstance (in, (PyObject *)&py_rpc_ptr_t_Type)) {
      if (!PyObject_IsInstance (in, (PyObject *)t)) {
	PyErr_SetString (PyExc_TypeError, "rpc_ptr of wrong type passed");
      } else {
	Py_INCREF (in);
	ret = reinterpret_cast<py_rpc_ptr_t *> (in);
      }
    } else {
      PyErr_SetString (PyExc_TypeError, "type mismatch for rpc_ptr");
    }
    return reinterpret_cast<PyObject *> (ret);
  }
};

template<> struct converter_t<pyw_void>
{
  static PyObject * convert (PyObject *in)
  {
    if (in && in != Py_None) {
      PyErr_SetString (PyExc_TypeError, "expected void argument");
      return NULL;
    }
    Py_INCREF (Py_None);
    return Py_None;
  }
};

PyObject *unwrap_error (void *);
pyw_base_t *wrap_error (PyObject *o, PyObject *e);

//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
// Shortcuts for allocating integer types all at once
//


#define INT_XDR_CLASS(ctype,ptype)                               \
class pyw_##ctype : public pyw_tmpl_t<pyw_##ctype, PyLongObject> \
{                                                                \
public:                                                          \
  pyw_##ctype ()                                                 \
    : pyw_tmpl_t<pyw_##ctype, PyLongObject> (&PyLong_Type) {}    \
  pyw_##ctype (pyw_err_t e)                                      \
    : pyw_tmpl_t<pyw_##ctype, PyLongObject> (e, &PyLong_Type) {} \
  pyw_##ctype (const pyw_##ctype &p)                             \
    : pyw_tmpl_t<pyw_##ctype, PyLongObject> (p) {}               \
  pyw_##ctype (PyObject *o)                                      \
    : pyw_tmpl_t<pyw_##ctype, PyLongObject> (o, &PyLong_Type) {} \
  ctype get () const                                             \
  {                                                              \
    ctype c = 0;                                                 \
    (void )get (&c);                                             \
    return c;                                                    \
  }                                                              \
  bool get (ctype *t) const                                      \
  {                                                              \
    if (!_obj) {                                                 \
      PyErr_SetString (PyExc_UnboundLocalError,                  \
                      "unbound int/long");                       \
      return false;                                              \
    }                                                            \
    *t = PyLong_As##ptype (_obj);                                \
    return true;                                                 \
  }                                                              \
  bool set (ctype i) {                                           \
    Py_XDECREF (_obj);                                           \
    return (_obj = PyLong_From##ptype (i));                      \
  }                                                              \
  bool init ()                                                   \
  {                                                              \
    if (!pyw_tmpl_t<pyw_##ctype, PyLongObject>::init ())         \
      return false;                                              \
    _typ = &PyLong_Type;                                         \
    return true;                                                 \
  }                                                              \
};

inline PyObject *
convert_int (PyObject *in)
{
  PyObject *out = NULL;
  assert (in);
  if (PyInt_Check (in))
    out = PyLong_FromLong (PyInt_AsLong (in));
  else if (!PyLong_Check (in)) {
    PyErr_SetString (PyExc_TypeError,
		     "integer or long value expected");
  } else {
    out = in;
    Py_INCREF (out);
  }
  return out;
}

#define INT_CONVERTER(T)                                         \
template<> struct converter_t<T> {                               \
  static PyObject *convert (PyObject *in)                        \
   { return convert_int (in); }                                  \
};

#define RPC_TRAVERSE_DECL(ptype)                                 \
bool rpc_traverse (XDR *xdrs, ptype &obj);

#define INT_DO_ALL_H(ctype,ptype)                                \
XDR_DECL(pyw_##ctype)                                            \
INT_XDR_CLASS(ctype, ptype)                                      \
RPC_TRAVERSE_DECL(pyw_##ctype)                                   \
PY_RPC_TYPE2STR_DECL(ctype)                                      \
RPC_PRINT_TYPE_DECL(pyw_##ctype)                                 \
RPC_PRINT_DECL(pyw_##ctype)                                      \
INT_CONVERTER(pyw_##ctype);                                      \
ALLOC_DECL(pyw_##ctype);                                         \
NOT_A_STRUCT(pyw_##ctype);

INT_DO_ALL_H(u_int32_t, UnsignedLong);
INT_DO_ALL_H(int32_t, Long);
INT_DO_ALL_H(u_int64_t, UnsignedLongLong);
INT_DO_ALL_H(int64_t, LongLong);

RPC_TRAVERSE_DECL(pyw_void);
RPC_PRINT_TYPE_DECL (pyw_void);
RPC_PRINT_DECL(pyw_void);
PY_RPC_TYPE2STR_DECL(void);
XDR_DECL(pyw_void);
ALLOC_DECL(pyw_void);

//
//-----------------------------------------------------------------------




//-----------------------------------------------------------------------
// RPC traversal
//

template<class W, size_t m> inline bool
rpc_encode (XDR *xdrs, pyw_tmpl_str_t<W,m> &obj)
{
  size_t sz;
  char *dat = obj.get (&sz);
  return dat && xdr_putint (xdrs, sz) && xdr_putpadbytes (xdrs, dat, sz);
}

template<size_t n> inline bool
rpc_traverse (XDR *xdrs, pyw_rpc_str<n> &obj)
{
  switch (xdrs->x_op) {
  case XDR_ENCODE:
    {
      return rpc_encode (xdrs, obj);
    }
  case XDR_DECODE:
    {
      u_int32_t size;
      if (!xdr_getint (xdrs, size) || size > n)
	return false;
      char *dp = (char *) XDR_INLINE (xdrs, size + 3 & ~3);
      if (!dp || memchr (dp, '\0', size))
	return false;
      return obj.set (dp, size);
    }
  case XDR_FREE:
    return obj.clear ();
  default:
    return true;
  }
}

template<class T, size_t n> inline bool
rpc_traverse (T &t, pyw_rpc_opaque<n> &obj)
{
  char buf[n];
  memset (buf, 0, n);
  for (u_int i = 0; i < n; i++) {
    buf[i] = obj[i].get_byte ();
    rpc_traverse (t, buf[i]);
  }
  PyObject *pystr = PyString_FromStringAndSize (buf, n);
  if (!pystr) {
    PyErr_NoMemory ();
    return false;
  }
  return obj.set_obj (pystr);
}

template<size_t n> inline bool
rpc_traverse (XDR *xdrs, pyw_rpc_opaque<n> &obj)
{
  char buf[n];
  switch (xdrs->x_op) {
  case XDR_ENCODE: 
    {
      size_t sz;
      char *dat = obj.get (&sz);
      if (!dat) return false;
      if (sz < n) {
	memset (buf, 0, n);
	memcpy (buf, dat, sz);
	dat = buf;
      }
      bool rc = xdr_putpadbytes (xdrs, dat, n);
      return rc;
    }
  case XDR_DECODE:
    {
      bool rc;
      if (!xdr_getpadbytes (xdrs, buf, n)) 
	rc = false;
      else
	rc = obj.set (buf, n);
      return rc;
    }
  case XDR_FREE:
    return obj.clear ();
  default:
    return true;
  }
}

template<size_t n> inline bool
rpc_traverse (XDR *xdrs, pyw_rpc_bytes<n> &obj)
{
  switch (xdrs->x_op) {
  case XDR_ENCODE:
    {
      return rpc_encode (xdrs, obj);
    }
  case XDR_DECODE:
    {
      u_int32_t size;
      if (!xdr_getint (xdrs, size) || size > n)
	return false;
      char *dp = (char *) XDR_INLINE (xdrs, size + 3 & ~3);
      if (!dp)
	return false;
      return obj.set (dp, size);
    }
  case XDR_FREE:
    return obj.clear ();
  default:
    return true;
  }
}


template<class L, class E> inline bool
rpc_traverse_slot (XDR *xdrs, L &v, u_int i)
{
  switch (xdrs->x_op) {
  case XDR_ENCODE:
    {
      E el (PYW_ERR_NONE);
      if (!v.get_slot (i, &el))
	return false;
      return rpc_traverse (xdrs, el);
    }
  case XDR_DECODE:
    {
      E el;
      if (!rpc_traverse (xdrs, el)) {
	return false;
      }
      PyObject *obj = el.get_obj ();
      bool rc = v.set_slot (obj, i);
      Py_DECREF (obj);
      return rc;
    }
  default:
    return false;
  }
}

template<class T, class A, size_t max> inline bool
rpc_traverse (T &t, pyw_rpc_vec<A,max> &obj)
{
  u_int size = obj.size ();

  if (!rpc_traverse (t, size))
    return false;
  if (size > obj.maxsize) {
    PyErr_SetString (PyExc_OverflowError,
		     "predeclared Length of rpc vector exceeded");
    return false;
  }
  if (size < obj.size ())
    if (!obj.shrink (size)) {
      PyErr_SetString (PyExc_RuntimeError,
		       "unexepected failure for array shrink");
      return false;
    }

  for (u_int i = 0; i < size; i++) 
    if (!rpc_traverse_slot<pyw_rpc_vec<A,max>, A> (t, obj, i))
      return false;

  return true;
}

template<class T, class A, size_t max> inline bool
rpc_traverse (T &t, pyw_rpc_array<A,max> &obj)
{
  u_int size = obj.size ();
  assert (size == max);
  for (u_int i = 0; i < size; i++)
    if (!rpc_traverse_slot<pyw_rpc_array<A,max>, A> (t, obj, i))
      return false;
  return true;
}

template<class C, class T, PyTypeObject *t> inline bool
rpc_traverse (C &c, pyw_rpc_ptr<T,t> &obj)
{
  bool nonnil = false;
  py_rpc_ptr_t *p = obj.casted_obj ();
  if (p && *p)
    nonnil = true;
  if (!rpc_traverse (c, nonnil))
    return false;
  if (nonnil) {
    T *x = obj.palloc ();
    if (!x)
      return false;
    return rpc_traverse (c, *x);
  }
  if (p)
    p->clear ();
  return true;
}

//
//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
//  RPC Program
//    - wrapper around the libasync type

struct py_rpc_program_t {
  PyObject_HEAD
  const rpc_program *prog;
  const py_rpcgen_table_t *pytab;
};

//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
// rpc_str and rpc_bytes methods
//

template<class W, size_t M> bool
pyw_tmpl_opq_t<W, M>::init_trav () const
{
  char *c;
  int i;
  if (_obj && PyString_AsStringAndSize (_obj, &c, &i) >= 0) {
    _str = c;
    _sz = i;
    return true;
  } else
    return false;
}

template<size_t m> bool
pyw_rpc_opaque<m>::get_char (u_int i, pyw_rpc_byte_t *c) const
{
  assert (_str);
  if (i >= maxsize) {
    c->set_err (PYW_ERR_BOUNDS);
  } else if (i >= _sz) {
    c->set_byte ('\0');
  } else {
    c->set_byte (_str[i]);
  }
  return true;
}

template<size_t M> bool
pyw_rpc_bytes<M>::get_char (u_int i, pyw_rpc_byte_t *c) const
{
  assert (_str);
  if (i >= _sz) {
    c->set_err (PYW_ERR_BOUNDS);
    return false;
  }
  c->set_byte (_str[i]);
  return true;
}

template<class W, size_t m> const char *
pyw_tmpl_str_t<W,m>::get () const 
{
  size_t dummy;
  return get (&dummy);
}

template<class W, size_t M> char *
pyw_tmpl_str_t<W,M>::get (size_t *sz) const
{
  char *ret;
  int i;

  if (!_obj) {
    PyErr_SetString (PyExc_UnboundLocalError, "undefined string");
    return NULL;
  }

  if ( PyString_AsStringAndSize (_obj, &ret, &i) < 0) {
    PyErr_SetString (PyExc_RuntimeError,
		     "failed to access string data");
    return NULL;
  }
  *sz = i;
  if (*sz > maxsize) {
    PyErr_SetString (PyExc_OverflowError, 
		     "Length of string exceeded\n");
    return NULL;
  }
  return ret;
}

template<class W, size_t m> bool 
pyw_tmpl_str_t<W,m>::set (char *buf, size_t len)
{
  Py_XDECREF (_obj);
  if (len > maxsize) {
    PyErr_SetString (PyExc_TypeError, 
		     "Length of string exceeded; value trunc'ed\n");
    len = maxsize;
  }
  return (_obj = PyString_FromStringAndSize (buf, len));
}

//
//-----------------------------------------------------------------------

//

//-----------------------------------------------------------------------
// py_rpc_vec methods
//

template<class T, size_t m> size_t
pyw_rpc_vec<T,m>::size () const
{
  int l = PyList_Size (_obj);
  if (l < 0) {
    PyErr_SetString (PyExc_RuntimeError,
		     "got negative list length from vector");
    return -1;
  }
  _sz = l;
  return u_int (l);
}

template<class L, class E> bool
vec_get_slot (u_int i, const L &l, E *out) 
{
  assert (i < l.size ());
  if (!l.const_obj ()) {
    PyErr_SetString (PyExc_UnboundLocalError, "unbound vector");
    return NULL;
  }
  PyObject *in = PyList_GetItem (const_cast<PyObject *> (l.const_obj ()), i);
  if (!in)
    return false;
  return out->safe_set_obj (in);
}

template<class T, size_t m> bool
pyw_rpc_vec<T,m>::get_slot (u_int i, T *out)  const
{
  return vec_get_slot (i, *this, out);
}

template<class T, size_t m> bool
pyw_rpc_array<T,m>::get_slot (u_int i, T *out) const
{
  return vec_get_slot (i, *this, out);
}


template<class T, size_t m> bool
pyw_rpc_vec<T,m>::set_slot (PyObject *el, u_int i)
{
  int rc;
  if (!el) {
    PyErr_SetString (PyExc_UnboundLocalError, "NULL array slot assignment");
    return false;
  } else if (i < _sz) {
    // SetItem "steals" a reference from us
    Py_INCREF (el); 
    if ((rc = PyList_SetItem (_obj, i, el)) < 0)
      Py_DECREF (el);
  } else if (i == _sz) {
    rc = PyList_Append (_obj, el);
    _sz ++;
  } else {
    PyErr_SetString (PyExc_IndexError, "out-of-order list insertion");
    return false;
  }
  return (rc >= 0);
}

template<class T, size_t m> bool
pyw_rpc_array<T,m>::set_slot (PyObject *el, u_int i)
{
  int rc;
  assert (i < size ());
  if (!el) {
    PyErr_SetString (PyExc_UnboundLocalError, "NULL array slot assignment");
    return false;
  }

  // Recall that PyList_Set "steals" a reference from us
  Py_INCREF (el);
  if ((rc = PyList_SetItem (_obj, i, el)) < 0)
    Py_DECREF (el);

  return (rc >= 0);
}

template<class T, size_t m> bool
pyw_rpc_vec<T,m>::shrink (size_t n)
{
  assert (n <= _sz);
  _sz -= n;
  int rc = PyList_SetSlice (_obj, _sz, _sz + n, NULL);
  return (rc >= 0);
}

//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
//

template<class W, class P> bool
pyw_tmpl_t<W,P>::safe_set_obj (PyObject *in)
{
  PyObject *out = converter_t<W>::convert (in);
  bool ret = out ? set_obj (out) : false;
  Py_XDECREF (out);
  return ret;
}

//
// XXX - copy_from will most likely be aborted
//
template<class W, class P> bool
pyw_tmpl_t<W,P>::copy_from (const pyw_tmpl_t<W,P> &src)
{
  const P *in = reinterpret_cast<const P *> (src.const_obj ());
  PyObject *n = py_obj_copy (in);
  if (!n)
    return false;
  bool rc = set_obj (n);
  Py_XDECREF (n);
  return rc;
}

//
//-----------------------------------------------------------------------

//-----------------------------------------------------------------------
// 
// debug functions
//

template<class W> str
getkey () 
{
  const char *typ = rpc_type2str<W>::type ();
  if (!typ) {
    warn << "no typ found for wrapper in debug_inc!\n";
    return NULL;
  }
  return typ;
}

//
//-----------------------------------------------------------------------


#endif
