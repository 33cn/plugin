/* $Id: malloc.C 3769 2008-11-13 20:21:34Z max $ */

#include "amisc.h"

#ifdef PYMALLOC
# ifdef _POSIX_C_SOURCE
#  undef _POSIX_C_SOURCE
# endif
# include <Python.h>
#endif /* PYMALLOC */

#ifdef DMALLOC
bool dmalloc_init::initialized;
void
dmalloc_init::init ()
{
  if (suidsafe () < 0
      && (getenv ("DMALLOC_OPTIONS") || getenv ("STKTRACE"))) {
    setgid (getgid ());
    setuid (getuid ());
    const char msg[] = "setuid disabled for malloc debugging\n";
    write (2, msg, sizeof (msg) - 1);
  }
  initialized = true;
}
#endif /* DMALLOC */

/* The xmalloc handler will have no effect when dmalloc is used, but
 * we leave it in for compatibility. */
static void
default_xmalloc_handler (int size)
{
  const char msg[] = "malloc failed\n";
  v_write (errfd, msg, sizeof (msg) - 1);
  myabort ();
}
void (*xmalloc_handler) (int) = default_xmalloc_handler;

#ifdef DMALLOC

#undef new
dmalloc_t dmalloc;

#ifndef CHECK_ARRAY_DELETE
# define CHECK_ARRAY_DELETE 1
#endif /* !CHECK_ARRAY_DELETE */
#if CHECK_ARRAY_DELETE
/* Want to catch array new followed by non-array delete */
enum { new_array_shift = 8 };
static char array_marker[new_array_shift] = { 
  0x79, 0x46, 0x55, 0x93, 0x12, 0x69, 0xaa, 0x7f
};
#endif /* !CHECK_ARRAY_DELETE */

void *
operator new (size_t size, dmalloc_t, const char *file, int line)
{
  /* As per the C++ standard, allocating 0 bytes must not return NULL,
   * and must return a different pointer each time... */
  if (!size)
    size = 1;
  return _xmalloc_leap (file, line, size);
}

void *
operator new[] (size_t size, dmalloc_t, const char *file, int line)
{
#if CHECK_ARRAY_DELETE
  size = size + new_array_shift;
#endif /* CHECK_ARRAY_DELETE */
  if (!size)
    size = 1;
#if CHECK_ARRAY_DELETE
  char *ret = static_cast <char *> (_xmalloc_leap (file, line, size));
  memcpy (ret, array_marker, new_array_shift);
  return ret + new_array_shift;
#else /* !CHECK_ARRAY_DELETE */
  return _xmalloc_leap (file, line, size);
#endif /* !CHECK_ARRAY_DELETE */
}

void *
operator new (size_t size, nothrow_t, const char *file, int line) throw ()
{
  if (!size)
    size = 1;
  return _malloc_leap (file, line, size);
}

void *
operator new[] (size_t size, nothrow_t, const char *file, int line) throw ()
{
#if CHECK_ARRAY_DELETE
  size = size + new_array_shift;
#endif /* CHECK_ARRAY_DELETE */
  if (!size)
    size = 1;
#if CHECK_ARRAY_DELETE
  char *ret = static_cast <char *> (_xmalloc_leap (file, line, size));
  memcpy (ret, array_marker, new_array_shift);
  return ret + new_array_shift;
#else /* !CHECK_ARRAY_DELETE */
  return _xmalloc_leap (file, line, size);
#endif /* !CHECK_ARRAY_DELETE */
}

#else /* !DMALLOC */

void *
xmalloc (size_t size)
{
  void *p;
#ifdef PYMALLOC
  if (!(p = PyMem_Malloc (size)))
#else /* ! PYMALLOC */
  if (!(p = malloc (size)))
#endif /* PYMALLOC */
    default_xmalloc_handler (size);
  return p;
}

void *
xrealloc (void *o, size_t size)
{
  void *p;
#ifdef PYMALLOC
  if (!(p = PyMem_Realloc (o, size)))
#else /* ! PYMALLOC */
  if (!(p = realloc (o, size)))
#endif /* PYMALLOC */
    default_xmalloc_handler (size);
  return p;
}

char *
xstrdup (const char *s)
{
  char *d;
  d = (char *) xmalloc (strlen (s) + 1);
  strcpy (d, s);
  return d;
}

#endif /* !DMALLOC */

using std::bad_alloc;
using std::nothrow_t;

void *
operator new (size_t size) throw (bad_alloc)
{
  if (!size)
    size = 1;
  return txmalloc (size);
}

void *
operator new (size_t size, nothrow_t) throw ()
{
  if (!size)
    size = 1;
#if PYMALLOC
  return PyMem_Malloc (size);
#else /* !PYMALLOC */
  return malloc (size);
#endif
}

#ifndef delete_throw
# define delete_throw throw()
#endif /* !delete_throw */

#ifndef DMALLOC

void
operator delete (void *ptr) delete_throw
{
  xfree (ptr);
}

void *
operator new[] (size_t size) throw (bad_alloc)
{
  if (!size)
    size = 1;
  return txmalloc (size);
}

void *
operator new[] (size_t size, nothrow_t) throw ()
{
  if (!size)
    size = 1;
#if PYMALLOC
  return PyMem_Malloc (size);
#else /* !PYMALLOC */
  return malloc (size);
#endif
}

void
operator delete[] (void *ptr) delete_throw
{
  xfree (ptr);
}

#else /* DMALLOC */

#include <ihash.h>

struct hashptr {
  hashptr () {}
  hash_t operator() (const void *obj) const
    { return reinterpret_cast<u_long> (obj); }
};
struct objref {
  const void *obj;
  const char *refline;
  int *const flagp;
  ihash_entry<objref> hlink;
  
  objref (const void *o, const char *fl, int *fp);
  ~objref ();
};
static ihash<const void *, objref, &objref::obj,
	     &objref::hlink, hashptr> objreftab;

inline
objref::objref (const void *o, const char *fl, int *fp)
  : obj (o), refline (fl), flagp (fp)
{
  objreftab.insert (this);
}
inline
objref::~objref ()
{
  objreftab.remove (this);
}

int nodelete_ignore_count;

static int do_nodelete_flag;
inline bool
do_nodelete ()
{
  return do_nodelete_flag > 0 && !nodelete_ignore_count
    && !globaldestruction && objreftab.constructed ();
}
void
nodelete_addptr (const void *obj, const char *fl, int *fp)
{
  if (!do_nodelete_flag) {
    u_long dmalloc_flags = dmalloc_debug_current ();
    do_nodelete_flag = (dmalloc_flags && 0x800) ? 1 : -1;
  }
  if (do_nodelete ())
    vNew objref (obj, fl, fp);
}
void
nodelete_remptr (const void *obj, const char *fl, int *fp)
{
  if (do_nodelete ())
    for (objref *oref = objreftab[obj]; oref; oref = objreftab.nextkeq (oref))
      if (oref->refline == fl && oref->flagp == fp) {
	delete oref;
	return;
      }
}
inline void
nodelete_check (const void *ptr)
{
  if (do_nodelete ())
    for (objref *oref = objreftab[ptr]; oref;
	 oref = objreftab.nextkeq (oref)) {
      if (oref->flagp)
	(*oref->flagp)++;
      else
	panic ("deleting ptr %p still referenced from %s\n",
	       ptr, oref->refline);
    }
}

void
operator delete (void *ptr) delete_throw
{
  nodelete_check (ptr);
  if (stktrace_record > 0)
    dmalloc_free (__backtrace (__FILE__, 2), __LINE__, ptr, DMALLOC_FUNC_FREE);
  else
    xfree (ptr);
}

void *
operator new[] (size_t size) throw (bad_alloc)
{
#if CHECK_ARRAY_DELETE
  size = size + new_array_shift;
#endif /* CHECK_ARRAY_DELETE */
  if (!size)
    size = 1;
#if CHECK_ARRAY_DELETE
  char *ret = static_cast <char *> (txmalloc (size));
  memcpy (ret, array_marker, new_array_shift);
  return ret + new_array_shift;
#else /* !CHECK_ARRAY_DELETE */
  return txmalloc (size);
#endif /* !CHECK_ARRAY_DELETE */
}

void *
operator new[] (size_t size, nothrow_t) throw ()
{
#if CHECK_ARRAY_DELETE
  size = size + new_array_shift;
#endif /* CHECK_ARRAY_DELETE */
  if (!size)
    size = 1;
#if CHECK_ARRAY_DELETE
  char *ret = static_cast <char *> (txmalloc (size));
  memcpy (ret, array_marker, new_array_shift);
  return ret + new_array_shift;
#else /* !CHECK_ARRAY_DELETE */
  return txmalloc (size);
#endif /* !CHECK_ARRAY_DELETE */
}

void
operator delete[] (void *_ptr) delete_throw
{
#if CHECK_ARRAY_DELETE
  char *ptr = static_cast<char *> (_ptr) - new_array_shift;
  if (memcmp (ptr, array_marker, new_array_shift)) {
    char msg[] = "non-array delete of array (or fencepost error)\n";
    write (errfd, msg, sizeof (msg) - 1);
  }
  xfree (ptr);
#else /* !CHECK_ARRAY_DELETE */
  xfree (_ptr);
#endif /* !CHECK_ARRAY_DELETE */
}

#endif /* DMALLOC */
