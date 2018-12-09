
// -*- c++ -*-
#include "safeptr.h"
#include "vatmpl.h"

namespace sp {

  //=======================================================================

  template<class T> class referee;

  //=======================================================================

  template<class T>
  class wkref : public base_ptr<T> {
  public:
    wkref (T *o) : _obj (o) { linkup (); }
    wkref (T &o) : _obj (&o) { linkup (); }
    wkref (const wkref<T> &r) : _obj (r._obj) { linkup (); }

    wkref () : _obj (NULL) {}
    ~wkref () { clear (); }

    wkref<T> &operator= (const wkref<T> &w2) {
      clear ();
      _obj = w2._obj;
      linkup ();
      return *this;
    }
    
    wkref<T> &operator= (T* o) {
      clear ();
      _obj = o;
      linkup ();
      return *this;
    }

    wkref<T> &operator= (T &o) {
      clear ();
      _obj = o;
      linkup ();
      return *this;
    }

    friend class referee<T>;

  protected:
    void clear () { if (_obj) _obj->rm (this); _obj = NULL; }
    void linkup () { if (_obj) { _obj->add (this); } }

    const T *obj () const { return _obj; }
    T *obj () { return _obj; }
    list_entry<wkref<T> > _lnk;


    T *_obj;
  };
  

  //=======================================================================

  template<class T>
  class referee {
  public:
    referee () {}
    ~referee ()
    {
      while (_lst.first) {
	wkref<T> *i = _lst.first;
	assert (*i);
	i->clear (); // will remove i from list!
      }
    }
    void add (wkref<T> *i) { _lst.insert_head (i); } 
    void rm (wkref<T> *i) { _lst.remove (i); } 
  private:
    list<wkref<T>, &wkref<T>::_lnk> _lst;
  };

  //=======================================================================

  //
  // A 'manual' pointer class.  If you hold a pointer, you can manually
  // dealloc the object it points to, killing the object and resetting
  // all references to the pointer.  Thus, one cannot double-deallocate
  // it, and one cannot access the object it points to after deallocation.
  //
  template<class T>
  class man_ptr : public wkref<T> {
  public:
    man_ptr (T *o) : wkref<T> (o) {}

    void dealloc () 
    { 
      if (this->obj ()) 
	delete this->obj (); 
      assert (!this->obj ());
    }

  };

  //=======================================================================

#define OP (
#define CP )

  //
  // An allocator class for man_ptr's:
  //
  //  sp::man_ptr<foo_t> = sp::man_alloc<foo_t> (a, b, x);
  //
  template<class T>
  class man_alloc {
  public:

    //
    // Note, need to use OP instead of '(' to satisfy the preprocessor.
    // Same goes for CP instead of ')'
    //
    VA_TEMPLATE( explicit man_alloc ,		\
		 : _p OP man_ptr<T> OP New T ,	\
				   CP CP {} )
    
    operator man_ptr<T>&() { return _p; }
    operator const man_ptr<T> &() const { return _p; }
  private:
    man_ptr<T> _p;
  };

#undef CP
#undef OP


  //=======================================================================


};
