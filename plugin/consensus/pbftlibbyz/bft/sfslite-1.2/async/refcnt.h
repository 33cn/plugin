// -*-c++-*-
/* $Id: refcnt.h 3758 2008-11-13 00:36:00Z max $ */

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

/* This is a simple reference counted garbage collector.  The usage is
 * as follows:
 *
 *   class foo : public bar { ... };
 *
 *      ...
 *      ref<foo> f = new refcounted<foo> ( ... );
 *      ptr<bar> b = f;
 *      f = new refcounted<foo> ( ... );
 *      b = NULL;
 *
 * A refcounted<foo> takes the same constructor arguments as foo,
 * except that constructors with more than 7 arguments cannot be
 * called.  (This is because there are no varargs templates.  You can
 * raise the limit from 7 to an arbitrary number if you wish, however,
 * by editting vatmpl.h.)
 *
 * A ptr<foo> behaves like a foo *, except that it is reference
 * counted.  The foo will be deleted when all pointers go away.  Also,
 * array subscripts will not work on a ptr<foo>.  You can only
 * allocate one reference counted object at a time.
 *
 * A ref<foo> is like a ptr<foo>, except that a ref<foo> can never be
 * NULL.  If you try to assign a NULL ptr<foo> to a ref<foo> you will
 * get an immediate core dump.  The statement "ref<foo> = NULL" will
 * generate a compile time error.
 *
 * A "const ref<foo>" cannot change what it is pointing to, but the
 * foo pointed to can be modified.  A "ref <const foo>" points to a
 * foo you cannot change.  A ref<foo> can be converted to a ref<const
 * foo>.  In general, you can implicitly convert a ref<A> to a ref<B>
 * if you can implicitly convert an A* to a B*.
 *
 * You can also implicitly convert a ref<foo> or ptr<foo> to a foo *.
 * Many functions can get away with taking a foo * instead of a
 * ptr<foo> if they don't eliminate any existing references (or the
 * foo's address around after returning).
 *
 * On both the Pentium and Pentium Pro, a function taking a ref<foo>
 * argument usually seems to take 10-15 more cycles the same function
 * with a foo * argument.  With some versions of g++, though, this
 * number can go as high as 50 cycles unless you compile with
 * '-fno-exceptions'.
 *
 * Sometimes you want to do something other than simply free an object
 * when its reference count goes to 0.  This can usually be
 * accomplished by the reference counted object's destructor.
 * However, after a destructor is run, the memory associated with an
 * object is freed.  If you don't want the object to be deleted, you
 * can define a finalize method that gets invoked once the reference
 * count goes to 0.  Any class with a finalize method must declare a
 * virtual base class of refcount.  For example:
 *
 *   class foo : public virtual refcount {
 *     ...
 *     void finalize () { recycle (this); }
 *   };
 *
 * Occasionally you may want to generate a reference counted ref or
 * ptr from an ordinary pointer.  This might, for instance, be used by
 * the recycle function above.
 *
 * You can do this with the function mkref, but again only if the
 * underlying type has a virtual base class of refcount.  Given the
 * above definition, recycle might do this:
 *
 *   void
 *   recycle (foo *fp)
 *   {
 *     ref<foo> = mkref (fp);
 *     ...
 *   }
 *
 * Note that unlike in Java, an objects finalize method will be called
 * every time the reference count reaches 0, not just the first time.
 * Thus, there is nothing morally wrong with "resurrecting" objects as
 * they are being garbage collected.
 *
 * Use of mkref is potentially dangerous, however.  You can disallow
 * its use on a per-class basis by simply not giving your object a
 * public virtual base class of refcount.
 *
 *   class foo {
 *     // fine, no mkref or finalize allowed
 *   };
 *
 *   class foo : private virtual refcount {
 *     void finalize () { ... }
 *     // finalize will work, but not mkref
 *   };
 *
 * If you like to live dangerously, there are a few more things you
 * can do (but probably shouldn't).  You can keep the original
 * "refcounted<foo> *" around and use it to generate more references
 * from a finalize method (or elsewhere).  If foo has a virtual base
 * class of refcount, it will also inherit the methods refcount_inc()
 * and refcount_dec().  You can use these to create memory leaks and
 * crash your program, respectively.
 *
 * A note on virtual base classes:  refcounted<foo> is derived from
 * type foo.  Thus, if foo has any virtual base classes, only the
 * default constructor will be called when a refcounted<foo> is
 * allocated--regardless of whatever constructor foo's constructor
 * invokes.  As an alternative, you can allocate a
 *
 *    refcounted<foo, vbase>
 *
 * However, you can only do this if foo does not have a virtual base
 * class of recounted and foo does not have a finalize class.
 */

#ifndef _REFCNT_H_INCLUDED_
#define _REFCNT_H_INCLUDED_ 1

#if VERBOSE_REFCNT
#include <typeinfo>
void refcnt_warn (const char *op, const type_info &type, void *addr, int cnt);
#endif /* VERBOSE_REFCNT */

#include "opnew.h"
#include "vatmpl.h"

class __globaldestruction_t {
  static bool started;
public:
  ~__globaldestruction_t () { started = true; }
  operator bool () { return started; }
};
static __globaldestruction_t globaldestruction;
/* The following is for the end of files that use bssptr's */
#define GLOBALDESTRUCT static __globaldestruction_t __gd__ ## __LINE__

template<class T> class ref;
template<class T> class ptr;
enum reftype { scalar, vsize, vbase };
template<class T, reftype = scalar> class refcounted;
template<class T> ref<T> mkref (T *);

class refcount {
  u_int refcount_cnt;
  virtual void refcount_call_finalize () = 0;
  friend class refpriv;
protected:
  refcount () : refcount_cnt (0) {}
  virtual ~refcount () {}
  void finalize () { delete this; }
  void refcount_inc () {
#if VERBOSE_REFCNT
    refcnt_warn ("INC", typeid (*this), this, refcount_cnt + 1);
#endif /* VERBOSE_REFCNT */
    refcount_cnt++;
  }
  void refcount_dec () {
#if VERBOSE_REFCNT
    refcnt_warn ("DEC", typeid (*this), this, refcount_cnt - 1);
#endif /* VERBOSE_REFCNT */
    if (!--refcount_cnt)
      refcount_call_finalize ();
  }
  u_int refcount_getcnt () const { return refcount_cnt; }
};

class refpriv {
protected:
  /* We introduce a private type "privtype," that users won't have
   * floating around.  The idea is that NULL can implicitly be
   * converted to a privtype *.  Thus, when passing NULL as an
   * argument to a function taking a ptr<T>, a null ptr will be
   * constructed.  Likewise, if f is of type ptr<T>, the assignment f
   * = NULL will make f null.  By omitting this from ref, we can
   * ensure that ref<foo> f = NULL results in a compile-time error. */
  class privtype {};

private:
#ifndef NO_TEMPLATE_FRIENDS
  template<class U> friend class ref;
  template<class U> friend class ptr;
#else /* NO_TEMPLATE_FRIENDS */
protected:
#endif /* NO_TEMPLATE_FRIENDS */

  static void rdec (refcount *c) { c->refcount_dec (); }
  static void rinc (refcount *c) { c->refcount_inc (); }
  template<class T> static void rinc (const ::ref<T> &r) { r.inc (); }
  template<class T> static void rinc (const ::ptr<T> &r) { r.inc (); }
  template<class T, reftype v> static void rinc (refcounted<T, v> *pp)
    { pp->refcount_inc (); }
  template<class T> static T *rp (const ::ref<T> &r) { return r.p; }
  template<class T> static T *rp (const ::ptr<T> &r) { return r.p; }
  template<class T, reftype v> static T *rp (refcounted<T, v> *pp)
    { return *pp; }
  static refcount *rc (refcount *c) { return c; } // Make gcc happy ???
  template<class T> static refcount *rc (const ::ref<T> &r) { return r.c; }
  template<class T> static refcount *rc (const ::ptr<T> &r) { return r.c; }
  template<class T, reftype v> static refcount *rc (refcounted<T, v> *pp)
    { return pp; }

  refcount *c;
  explicit refpriv (refcount *cc) : c (cc) {}
  refpriv () {}

public:
#if 0
  template<class T> bool operator== (const ::ref<T> &r) const
    { return c == r.c; }
  template<class T> bool operator== (const ::ptr<T> &r) const
    { return c == r.c; }
  template<class T> bool operator!= (const ::ref<T> &r) const
    { return c != r.c; }
  template<class T> bool operator!= (const ::ptr<T> &r) const
    { return c != r.c; }
#else
  bool operator== (const refpriv &r) const { return c == r.c; }
  bool operator!= (const refpriv &r) const { return c != r.c; }
#endif

  void *Xleak () const { rinc (c); return c; }
  static void Xplug (void *c) { rdec ((refcount *) c); }
};

template<class T> struct type2struct {
  typedef T type;
};
#define TYPE2STRUCT(t, T)			\
template<t> struct type2struct<T> {		\
  struct type {					\
    T v;					\
    operator T &() { return v; }		\
    operator const T &() const { return v; }	\
    type () {}					\
    type (const T &vv) : v (vv) {}		\
  };						\
}
TYPE2STRUCT(, bool);
TYPE2STRUCT(, char);
TYPE2STRUCT(, signed char);
TYPE2STRUCT(, unsigned char);
TYPE2STRUCT(, int);
TYPE2STRUCT(, unsigned int);
TYPE2STRUCT(, long);
TYPE2STRUCT(, unsigned long);
TYPE2STRUCT(class U, U *);

template<class T>
class refcounted<T, scalar>
  : virtual private refcount, private type2struct<T>::type
{
  friend class refpriv;

  virtual void XXX_gcc_repo_workaround () {} // XXX - egcs bug

  operator T *() { return &static_cast<T &> (*this); }
  /* When the reference count on an object goes to 0, the object is
   * deleted by default.  However, some classes may not wish to
   * deallocate their memory at the time the reference count goes to
   * 0.  (For example, a network stream class may wish to finish
   * writing buffered data to the network asynchronously, and delete
   * itself at a later point.)
   *
   * Classes can therefore specify a (non-virtual) "void finalize ()"
   * method to be called when the reference count goes to 0.  Any
   * class with a finalize method must also have refcount as a virtual
   * base class.
   *
   * So what is refcount_call_finalize all about?  Here is a more
   * obvious implementation:
   *
   *   class refcount {
   *     virtual void finalize () { delete this; }
   *     virtual ~refcount () {}
   *     ...
   *     void refcount_dec () { if (!--refcount_cnt) finalize (); }
   *   };
   *
   *   class myclass : public virtual refcount {
   *     void finalize () { ... }
   *   };
   *
   * But there are inconveniences.  If the user forgets to give
   * myclass a virtual refcount supertype, the code will still
   * compile--it will just behave incorrectly at run time.
   *
   * This code solves the problem by calling finalize from
   * refcounted<T> rather than refcount.  If the user forgets to give
   * myclass a virtual refcount subtype, the call to finalize from
   * refount_call_finalize is ambiguous and flags an error.
   *
   * An added benefit of this scheme is that refcount is now a pure
   * virtual class (call_finalize is an abstract method).  This means
   * myclass also becomes a pure virtual class, and cannot be
   * allocated on its own (only as part of a refcounted<T>).  Of
   * course, if you really want to circumvent this restriction, you
   * can do so by giving myclass a virtual call_finalize() method.
   * The common case will probably be that a class with a finalize
   * method always expects to be refcounted.  This scheme makes it
   * hard to violate the requirement accidentally. */
  void refcount_call_finalize () {
    /* An error on the following line probably means you forgot to
     * give T a virtual base class of refcount.  Alternatively, T
     * already has a method called finalize unrelated to the reference
     * counting.  In that case you will have to rename finalize to use
     * a refcounted<T>. */
    finalize ();
  }

  ~refcounted () {}

public:
  VA_TEMPLATE (explicit refcounted, : type2struct<T>::type, {})
};

template<class T>
class refcounted<T, vbase>
  : virtual private refcount
{
  friend class refpriv;
  typedef typename type2struct<T>::type obj_t;
  obj_t obj;

  // virtual void XXX_gcc_repo_workaround () {} // XXX - egcs bug

  operator T *() { return &static_cast<T &> (obj); }
  void refcount_call_finalize () { delete this; }
  ~refcounted () {}

public:
  VA_TEMPLATE (explicit refcounted, : obj, {})
};


template<class T>
class refcounted<T, vsize>
  : virtual private refcount
{
  friend class refpriv;
  typedef refcounted<T, vsize> rc_t;

  virtual void XXX_gcc_repo_workaround () {} // XXX - egcs bug

  operator T *() { return tptr (this); }
  refcounted () { new ((void *) tptr (this)) T; }
  void refcount_call_finalize () {
    tptr (this)->~T ();
    delete this;
  }
  
  ~refcounted () {}

public:
  static rc_t *alloc (size_t n)
    { return new (opnew (n + (size_t) tptr (NULL))) rc_t; }
  static T *tptr (rc_t *rcp)
    { return (T *) ((char *) rcp + ((sizeof (*rcp) + 7) & ~7)); }
};

#define REFOPS_DEFAULT(T)			\
protected:					\
  T *p;						\
  refops () {}					\
						\
public:						\
  T *get () const { return p; }			\
  operator T *() const { return p; }		\
  T *operator-> () const { return p; }		\
  T &operator* () const { return *p; }

template<class T>
class refops {
  REFOPS_DEFAULT (T)
};

template<>
class refops<void> {
protected:
  void *p;
  refops () {}

public:
  operator void *() const { return p; }
  void *get () const { return p; }
};

template<class T> class mkcref;

template<class T>
class ref : public refpriv, public refops<T> {
  friend class refpriv;
  using refops<T>::p;

  friend ref<T> mkref<T> (T *);
  friend class mkcref<T>;
  ref (T *pp, refcount *cc) : refpriv (cc) { p = pp; inc (); }

  void inc () const { rinc (c); }
  void dec () const { rdec (c); }

public:
  typedef T type;
  typedef struct ptr<T> ptr;

  template<class U, reftype v>
  ref (refcounted<U, v> *pp)
    : refpriv (rc (pp)) { p = refpriv::rp (pp); inc (); }
  /* At least with gcc, the copy constructor must be explicitly
   * defined (though it would appear to be redundant given the
   * template constructor bellow). */
  ref (const ref<T> &r) : refpriv (r.c) { p = r.p; inc (); }
  template<class U>
  ref (const ref<U> &r)
    : refpriv (rc (r)) { p = refpriv::rp (r); inc (); }
  template<class U>
  ref (const ::ptr<U> &r)
    : refpriv (rc (r)) { p = refpriv::rp (r); inc (); }

  ~ref () { dec (); }

  template<class U, reftype v> ref<T> &operator= (refcounted<U, v> *pp)
    { rinc (pp); dec (); p = refpriv::rp (pp); c = rc (pp); return *this; }

  /* The copy assignment operator must also explicitly be defined,
   * despite a redundant template. */
  ref<T> &operator= (const ref<T> &r)
    { r.inc (); dec (); p = r.p; c = r.c; return *this; }
  template<class U> ref<T> &operator= (const ref<U> &r)
    { rinc (r); dec (); p = refpriv::rp (r); c = rc (r); return *this; }
  /* Self asignment not possible.  Use ref::inc to cause segfauls on NULL. */
  template<class U> ref<T> &operator= (const ::ptr<U> &r)
    { dec (); p = refpriv::rp (r); c = rc (r); inc (); return *this; }
};

/* To skip initialization of ptr's in BSS */
struct __bss_init {};

template<class T>
class ptr : public refpriv, public refops <T> {
  friend class refpriv;
  using refops<T>::p;

  void inc () const { if (c) (rinc (c)); }
  void dec () const { if (c) (rdec (c)); }

  template<class U, reftype v>
  void set (refcounted<U, v> *pp, bool decme) {
    if (pp) {
      rinc (pp);
      if (decme)
	dec ();
      p = refpriv::rp (pp);
      c = rc (pp);
    }
    else {
      if (decme)
	dec ();
      p = NULL;
      c = NULL;
    }
  }

public:
  typedef T type;
  typedef struct ref<T> ref;

  explicit ptr (__bss_init) {}
  ptr () : refpriv (NULL) { p = NULL; }
  ptr (privtype *) : refpriv (NULL) { p = NULL; }
  template<class U, reftype v>
  ptr (refcounted<U, v> *pp) { set (pp, false); }
  ptr (const ptr<T> &r) : refpriv (r.c) { p = r.p; inc (); }
  template<class U>
  ptr (const ptr<U> &r)
    : refpriv (rc (r)) { p = refpriv::rp (r); inc (); }
  template<class U>
  ptr (const ::ref<U> &r)
    : refpriv (rc (r)) { p = refpriv::rp (r); inc (); }

  ~ptr () { dec (); }

  ptr<T> &operator= (privtype *)
    { dec (); p = NULL; c = NULL; return *this; }
  template<class U, reftype v> ptr<T> &operator= (refcounted<U, v> *pp)
    { set (pp, true); return *this; }

  ptr<T> &operator= (const ptr<T> &r)
    { r.inc (); dec (); p = r.p; c = r.c; return *this; }
  template<class U> ptr<T> &operator= (const ptr<U> &r)
    { rinc (r); dec (); p = refpriv::rp (r); c = rc (r); return *this; }
  template<class U> ptr<T> &operator= (const ::ref<U> &r)
    { rinc (r); dec (); p = refpriv::rp (r); c = rc (r); return *this; }
};

template<class T>
struct bssptr : ptr<T> {
  using ptr<T>::Xleak;
  // Don't initialize (assume we were 0 initialized in the BSS)
  bssptr () : ptr<T> (__bss_init ()) {}
  // Override the effects of destruction
  ~bssptr () { assert (globaldestruction); if (*this != NULL) Xleak (); }
  ptr<T> &operator= (refpriv::privtype *p) { return ptr<T>::operator= (p); }
  template<class U> ptr<T> &operator= (const ptr<U> &r)
    { return ptr<T>::operator= (r); }
  template<class U> ptr<T> &operator= (const ::ref<U> &r)
    { return ptr<T>::operator= (r); }
};

template<class T> inline ref<T>
mkref (T *p)
{
  return ref<T> (p, p);
}

template<class T>
struct mkcref {
  static ref<T> mkref (T *p)
  {
    return ref<T> (p, const_cast<refcount *> (static_cast<const refcount *> (p)));
  }
};

template<class T> ref<const T>
mkref (const T *p)
{
  return mkcref<const T>::mkref (p);
}


#endif /* !_REFCNT_H_INCLUDED_ */
