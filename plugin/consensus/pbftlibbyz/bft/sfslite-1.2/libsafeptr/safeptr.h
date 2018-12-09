
// -*- c++ -*-

//
// Safe pointers!
//

#ifndef __LIBSAFEPTR_SAFEPTR_H__
#define __LIBSAFEPTR_SAFEPTR_H__

namespace sp {

  // 
  // Class base_ptr<T>
  //
  //   A base class for all safe/smart pointers used on the system.
  //   These are the only safe operations that should be used.  Note
  //   we also assume the either the pointer has a valid value, or
  //   is empty.  I.e., no dangling pointers allowed!
  //
  template<class T>
  class base_ptr {
  public:
    virtual ~base_ptr () {}

#define DEREF(cnst, star) \
    cnst T *o = obj ();	  \
    assert (o);		  \
    return star o;

    T &operator* () { DEREF(,*); }
    const T &operator* () const { DEREF(const,*); }
    T *operator-> () { DEREF(,); }
    const T *operator-> () const { DEREF(const,); }
    
    const T *nonnull_volatile_ptr () const { DEREF(const,); }
    T *nonnull_volatile_ptr () { DEREF(,); }

#undef DEREF

    virtual operator bool() const { return obj(); }

    virtual bool operator== (const base_ptr<T> &p2) const 
    { return obj() == p2.obj (); }
    virtual bool operator!= (const base_ptr<T> &p2) const 
    { return obj() != p2.obj (); }

  protected:
    base_ptr () {}
    virtual const T *obj () const = 0;
    virtual T *obj () = 0;
  };
};


#endif /* __LIBSAFEPTR_SAFEPTR_H__ */
