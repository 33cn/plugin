/* Copyright (c) 1995 Andrew C. Myers */

#ifndef _HASH_MAP_H
#define _HASH_MAP_H

#include "basic.h"
#include "generator.h"
#include "bhash.h"
#include "buckets.h"

#include "th_assert.h"
#include <stddef.h>

/*
   \section{High-Performance Dynamically-Resizing Hash Table}

    This section defines three easy-to-use implementations of
    a parameterized map type, all based on the same underlying hash-table
    technology. Which implementation is chosen depends on the choice of
    key and value types.

    "Map<KEY,VALUE>": an efficient implementation using chained buckets.
        Has the weakness that each bucket is long-word aligned, possibly
        wasting space. It also generates new code for every instantiation.
    
    "PtrMap<KEY,VALUE>": on most machines, an efficient implementation for
        the case where KEY and VALUE are both pointer types. The win over
        "BucketMap" is that all "PtrMap"s share the same code.
    
    The interface to all these map classes is the following:

    int size() const;
      // Returns the number of (k, v) pairs in the table

    void add(KEY, VALUE);
      // Checks: the table does not already contain a mapping for the key.
      // Effects: Adds a new mapping to the hash table.

    void add1(KEY, VALUE);
      // "add" followed with allowAutoResize().

    bool find(KEY k, VALUE &v) const;
      // Place the value corresponding to this key in "v",
      //    and return true. Return false if no such mapping exists.

    VALUE& fetch(KEY k) const { return operator[](k); }
    VALUE& operator[](KEY k) const;
      // Checks: there is a mapping for "k".
      // Effects: returns the value corresponding to "k".

    bool contains(KEY k) const;
      // Return whether there is a mapping for "k".

    bool store(KEY, VALUE);
      // If a mapping already exists, store the value
      //     under the specified key, replacing the existing mapping, and
      //     return true. Otherwise, add a new mapping and return false.
      //     
      // Note: "add" is faster than "store" in the case where no
      //     mapping exists. When "store" is used on a key that has
      //     just been yielded by the "mapping" iterator, it may be
      //     faster than usual.

    bool store1(KEY, VALUE);
      // same as store except allowAutoResize() is called before store

    bool remove(KEY k, VALUE &v);
      // If there is a mapping for "k", return true and put the
      // corresponding value into "v". Otherwise, return false.

    VALUE remove_fast(KEY k);
      //    Checks: there is a mapping for "k"
      //    Effects: Removes the mapping for "k" and returns the
      //             corresponding value.

    bool find_or_add(KEY k, VALUE const &v1, VALUE &v2);
      //    Effects: If there is a mapping for "k", changes its value to
      //             "v1" and returns "true". Otherwise, places the
      //             current value in "v2" and returns "false".

    void clear();
      //   modifies - this
      //    effects - All bindings are removed from this.

    void predict(int size);
      //    Indicate that you would like the hash table to contain "size"
      //    buckets. This is useful for obtaining optimal
      //    performance when many new mappings are being added at once. It
      //    may be an expensive operation, though it has no semantic effect.
      //    For best performance, "size" should be at least as large as the
      //    number of elements. This operation is O(size()).
        
    void allowAutoResize();
      //    Allow the hash table to rehash its mappings so that optimal
      //    performance is achieved. This operation is O(size()).
      //    The hash table will not rehash the mappings if it is
      //    at a reasonable size already.

    float estimateEfficiency() const;
      //    Return the average expected number of buckets in a bucket chain.
      //    If this number is "n", the expected number of buckets that will be
      //    looked at when a "fetch" or "remove" operation is performed is
      //    roughly "1 + n/2". Efficiency should be 1 for maximal speed.

    float estimateClumping() const;
      //    Evaluate how well the current hash function is working. This number
      //    should be close to one if the hash function is working well. If
      //    the clumping is less than 1.0, the hash function is doing better
      //    than a perfect hash function would be expected to, by magically
      //    avoiding collisions. A clumping of K means that the hash function
      //    is doing as badly as a hash function that only generates a
      //    fraction 1/K of the hash indices.

    int memory_usage() const;
      //    Returns number of bytes used by this data structure.

   To iterate over the mappings in a map, use a generator: a data
   structure that successively produces each of the keys of the map.
   A generator of keys is used in the following fashion:

        Map<KEY, VALUE> m;
        ...
        KeyGenerator<KEY,VALUE> g(m);
        KEY k;
        while (g.get(k)) { ... }

   A generator of key and values is used in the following fashion:

        Map<KEY, VALUE> m;
        ...
        MapGenerator<KEY,VALUE> g(m);
        KEY k;
	VALUE v;
        while (g.get(k, v)) { ... }

   Any number of generators may exist for a map. While any generator
   exists, no non-const method of "Map" may be called on the object that
   the generator is attached to. The "remove" method of a "MapGenerator"
   may be used to modify the map, however.
*/

template <class KEY, class VALUE> struct HashPair {
public:
    HashPair() {} // needed for generators
    HashPair(KEY k, VALUE v) : key_(k), value_(v) {}
    ~HashPair() {}
    KEY key() { return key_; }
    VALUE value() { return value_; }
    int hash() const { return key_.hash(); }
    bool similar(HashPair<KEY, VALUE> const &hp) { return key_ == hp.key_; }
    void operator=(HashPair<KEY, VALUE> const &hp)
	{ key_ = hp.key_; value_ = hp.value_; }

private:
    KEY key_;
public:
    VALUE value_; // might as well expose this
};

/*
    The "Map" class is a easy-to-use map from KEY to VALUE, i.e.
    "Map<IntKey, char>" is a map from integers to characters. It supports
    all the operations described above. See "bhash.h" for implementation
    details.
*/
#define HP HashPair<KEY, VALUE>
#define _superclass BHash<HP, Buckets<HP>, BucketsGenerator<HP> >

template<class KEY, class VALUE> class KeyGenerator;
template<class KEY, class VALUE> class MapGenerator;

template <class KEY, class VALUE> class Map : private _superclass {
public:
    friend class KeyGenerator<KEY, VALUE>;
    friend class MapGenerator<KEY, VALUE>;
    Map(int sizehint) : _superclass(sizehint) {}
    Map() : _superclass()        {}
    int size() const             {  return _superclass::size(); }
    void add(KEY k, VALUE v)     { _superclass::add(HP(k, v)); }
    void add1(KEY k, VALUE v)    {  _superclass::add(HP(k, v)); 
                                    allowAutoResize();}
    bool find(KEY k, VALUE &v) const {
    				    HP hp(k, VALUE());
				    HP *r = _superclass::find(hp);
				    if (!r) return false;
				    v = r->value_;
				    return true; }
    bool find_or_add(KEY k, VALUE const &v1, VALUE &v2)
				 {  HP hp(k, v1);
				    HP *r = _superclass::find_or_add(hp);
				    if (!r) {  allowAutoResize();
					       return false;      } else
				            {  v2 = r->value_;
					       return true;       } }
    VALUE &fetch(KEY k) const    {  HP hp(k, VALUE());
				    return _superclass::find_fast(hp)->value_; }
    VALUE &operator[](KEY k) const { return fetch(k); }
    bool contains(KEY k) const  {   
				    return 0 !=_superclass::find(HP(k,VALUE()));
				}
    bool store(KEY k, VALUE v) {    return _superclass::store(HP(k, v)); }
    bool store1(KEY k, VALUE v) {   allowAutoResize();
    				    return _superclass::store(HP(k, v)); }
    bool remove(KEY k, VALUE &v) {  HP hp(k, VALUE());
				    bool ret = _superclass::remove(hp);
				    v = hp.value_;
				    return ret; }
    VALUE remove_fast(KEY k)   {    HP hp(k, VALUE());
				    _superclass::remove_fast(hp);
				    return hp.value_; }
    void clear()		 {  _superclass::clear(); }
    void predict(int size) {  _superclass::predict(size); }
    void allowAutoResize() {  _superclass::allowAutoResize(); }
    float estimateEfficiency() const
				 {  return _superclass::estimateEfficiency(); }
    float estimateClumping() const
    				 {  return _superclass::estimateClumping(); }
    int memory_usage() {return _superclass::memory_usage();}
};

#define BHG BHashGenerator<HP, Buckets<HP>, BucketsGenerator<HP> >

template<class KEY, class VALUE>
class KeyGenerator: public Generator<KEY>, private BHG {
public:
    KeyGenerator(Map<KEY, VALUE> const &m) :
	BHG((BHash<HP, Buckets<HP>, BucketsGenerator<HP> > &)m) {}
    virtual bool get(KEY &k) {
	HP hp;
	bool ret = BHG::get(hp);
	if (ret) k = hp.key();
	return ret;
    }
};

template<class KEY, class VALUE>
class MapGenerator: public Generator<KEY>, private BHG {
public:
    MapGenerator(Map<KEY, VALUE> &m) : BHG(m) {}
    virtual bool get(KEY &k) {
	HP hp;
	bool ret = BHG::get(hp);
	if (ret) k = hp.key();
	return ret;
    }
    virtual bool get(KEY &k, VALUE &v) {
	HP hp;
	bool ret = BHG::get(hp);
	if (ret) { k = hp.key(); v = hp.value(); }
	return ret;
    }
    virtual void remove() {
	BHG::remove();
    }
};

#undef BHG
#undef HP
#undef _superclass
/*
    The classes "KEY" and "VALUE" must conform to the following declarations:

class KEY {
    // A key must be a hashable value. Often, a "KEY" is a wrapper class
    // that provides an inlined "hash" implementation. Ideally, the hash
    // function does not lose information about the value that is being
    // hashed: a good implementation of a hash function for an integer key
    // would be the identity function.

    KEY(KEY const &);                    // Keys can be copied
    ~KEY();                              // Keys can be destroyed. 
    operator=(KEY const &);              // Keys can be overwritten
    operator==(KEY const &key2) const;   // Returns whether they are equal
    int hash() const;                    // Returns a hash key for this value 
      
    See "valuekey.h" for a macro that creates new key classes out of
    primitive types such as int and long.
};

class VALUE {
    VALUE(VALUE const &);         // Values can be copied.
    ~VALUE();                     // Values can be destroyed.
};

    The following classes are useful if you want to use a primitive type
    as a key, since the primitive types do not have a built-in hash()
    function:
*/

/* A "PtrKey" is a useful class that provides a key for any pointer type.
   T can be any type. */
template <class T>
class PtrKey {
public:
    PtrKey() : val(0) {}
    PtrKey(T *a) : val(a) {}
    void operator=(PtrKey<T> const &x) { val = x.val; }
    int hash() const { return int(*((ptrdiff_t *)&val) >> 2); }
    bool operator==(PtrKey<T> const &x)
        { return (x.val == val) ? true : false; } 
    T * val;
};

/*
    The "PtrMap" class is a easy-to-use map from KEY to VALUE, where KEY
    and VALUE are pointer types, e.g.
    "PtrMap<int *, Foo *>" is a map from integer pointers to Foo *.
    "PtrMap" supports all the operations of BHash.
*/
#define HP HashPair<PtrKey<void>, void *>
#define _superclass BHash<HP, Buckets<HP>, BucketsGenerator<HP> >
#define K PtrKey<void>((void *)k)

template <class KEY, class VALUE> class PtrMap : private _superclass {
public:
    PtrMap() : _superclass(0)    {}
    PtrMap(int sizehint)         : _superclass(sizehint) {}
    int size() const             {  return _superclass::size(); }
    void add(KEY k, VALUE v)   { _superclass::add(HP(K, (void *)v)); }
    bool find(KEY k, VALUE &v) {  HP hp(K, 0);
    				    HP *r = _superclass::find(hp);
				    if (!r) return false;
				    v = (VALUE)r->value_;
				    return true; }
    VALUE fetch(KEY k) const   {  HP hp(K, 0);
				    _superclass::find_fast(hp);
				    return (VALUE)hp.value_; }
    VALUE operator[](KEY k) const { return fetch(k); }
    bool contains(KEY k) const  {  HP hp(K,0);
				    return _superclass::find(hp) ? true:false; }
    bool store(KEY k, VALUE v) {  return _superclass::store(HP(k, v)); }
    bool remove(KEY k, VALUE &v) {
				    HP hp(k, 0);
				    bool ret = _superclass::remove(hp);
				    v = (VALUE)hp.value_;
				    return ret; }
    VALUE remove_fast(KEY k)   {  HP hp(k, 0);
				    _superclass::remove_fast(hp);
				    return (VALUE)hp.value_; }
    void clear()		 {  _superclass::clear(); }
    void predict(int size)       {  _superclass::predict(size); }
    void allowAutoResize()       {  _superclass::allowAutoResize(); }
    float estimateEfficiency() const
				 {  return _superclass::estimateEfficiency(); }
    float estimateClumping() const
    				 {  return _superclass::estimateClumping(); }
};
#undef _superclass
#undef K
#undef HP

#endif /* _HASH_MAP_H */
