/* Copyright (c) 1995 Andrew C. Myers */

#ifndef _BHASH_H
#define _BHASH_H

#pragma interface

//  Definitions for classes that are used to construct the Map classes.
//  A key piece is the "BHash" class, a hash table implementation that
//  is independent of the implementation of the individual buckets.
//
//  The "BHash" class is not intended for end use, since it exposes
//  internal pointers. In a multi-threaded application, data structures
//  using it will need to use locks to prevent problems.

#define BH_MAX_DESIRED_ALPHA 3
// Placed here so that allowAutoResize() can be inlined
// "alpha" is the ratio of items to the hash table to slots, as described
// in CLR, Chapter 12.

#define BH_TEMPLATE template <class ELEM, class SET, class GEN>

BH_TEMPLATE class BHashGenerator;

BH_TEMPLATE class BHash {
/*
    A "BHash" is a hash table that implements a set of ELEM type.
    The hash table is implemented as a parameterized class.
    
    The specifications of the methods are as described above.
    
    Each element in the bucket array is of another type SET, which is
    assumed also to implement a set of ELEM with at least O(n) performance.
    The constraints on the type SET are that it have mostly the methods
    of BHash, though it is not expected that BHash itself be used as the
    SET argument. The "Buckets" type is a standard implementation of SET.

    The set will never contain two similar elements, as reported by the
    method "similar" (which ELEM must support). However, elements may be
    similar without being exactly equal, which is useful when using "BHash"
    to implement maps.
*/
    friend class BHashGenerator<ELEM, SET, GEN>;
public:
    BHash(int size_hint);
      // Create a new hash table whose number of elements is expected to
      // be around "size_hint".
    BHash();
      // Create a new, empty hash table
    int size() const;
      // The number of elements in the set
    void add(ELEM const &);
      // Add a new element to the set. Checks that that element does
      // not already exist.
    ELEM *find(ELEM const &) const;
      // Return a pointer to the similar element contained in the set, 0
      // if none.
    ELEM *find_fast(ELEM const &) const;
      // Return a pointer to the similar element contained in the set.
      // Checks that there is such an element.
    ELEM *find_or_add(ELEM const &);
      // If the element is not in the set, add it and return 0.
      // Otherwise, return a pointer to the existing similar element
      // without modifying it.
    bool store(ELEM const &);
      // Put a new element in the set, or replace the equivalent element.
    bool remove(ELEM &e);
      // If the element is in the set, remove it and return true, placing
      // a copy of it in "e".
    void remove_fast(ELEM &e);
      // Remove the similar element from the set, replacing "e" with its
      // contents. Checks that such an element exists.
    void clear();
      // Remove all elements from the set.
    void predict(int size);
    void allowAutoResize() { predict(numItems / BH_MAX_DESIRED_ALPHA); }
    float estimateEfficiency() const;
    float estimateClumping() const;
    int memory_usage() const;
      // See "map.h"
//
// operations for stack-allocated and inlined hash tables
//
    ~BHash();
    BHash(BHash<ELEM, SET, GEN> const &);
    BHash<ELEM, SET, GEN> const &operator=(BHash<ELEM, SET, GEN> const &);
protected:
    int numItems;         // Number of elements in the hash table
    int numSlots;         // Size of the bucket array.
    unsigned right_shift; // How much to shift a hash value right by to make
                          // it fit within "0..numSlots-1"
    SET *buckets;         // The array of top-level buckets

    int do_hash(ELEM const &e) const; // compute hash value modulo "numSlots"
    void sizeup(int size);      // set slotBits, numSlots appropriately
    void copy_items(BHash<ELEM, SET, GEN> const &);
    void resize(int size);


#ifdef __GNUC__
    static BHashGenerator<ELEM, SET, GEN> *dummy; // force g++ expansion
    // this previous line actually makes cxx screw up for some reason
#endif

    /* Abstraction function: the hash table contains all elements that are
           found in all its non-empty buckets.
       Rep Invariant: The total number of non-empty buckets is equal to
           "numItems".  No two buckets contain the same key, and every
           bucket is found in a bucket chain whose index in "buckets"
           is what the bucket's key hashes to, modulo the number of
           slots in the bucket array. The number of slots must be a
           power of two, and "numSlots == 1<<slotBits". "slotBits" must
	   be at least 1 (this has to do with keeping some non-ANSI C
	   compilers happy).
    */
};

// The class SET must conform to the following declaration:

#if 0 // specification
class SET {
    // A "SET" represents a set of elements.
    // Typically, these sets will be very small, so insertion and
    // deletion should be optimized on this basis.
    // A "SET" is used as a inlined abstraction.
    SET();
        // Results: an empty SET.
    ~SET();
        // Results: destroys this SET.
    SET(SET const &);
	// create a copy
    void operator=(SET const &);
        // overwrite with a copy
    ELEM *find(ELEM const &e) const;
        // Place the element similar to "e" into "e"
        // and return "true". Return "false" if no such element exists,
	// and do not change "e".
    ELEM *find_fast(ELEM const &e) const;
        // Checks: there is an element similar to "e".
        // "e" is replaced with that element.
    void add(ELEM const &e);
        // Checks: No element in the set is similar to "e".
        // Results: Adds "e" to the set.
    bool store(ELEM const &e);
        // Results: Replaces the element similar to "e" with "e", or
	// adds "e" to the set if no such element exists.
	// Returns "true" if it replaces with "e".
    bool remove(ELEM &e);
        // Results: The element similar to "e" is removed and true is
	// returned. If no such element exists, false is returned and there
	// is no effect on "e".
    void remove_fast(ELEM &e);
        // Checks: there is an equivalent element
        // Results: the equivalent element is removed and copied into "e"
};
#endif // specification

#define BHashT BHash<ELEM,SET,GEN> 

BH_TEMPLATE class BHashGenerator {
/*
    See "generator.h".
    
    GEN must be a generator for SET. It must support operator= and
    the constructor GEN(SET &).

*/
public:
    BHashGenerator(BHashT &h) : gen(h.buckets[0]) {
	slot = 0;
	hash_table = &h;
    }
    virtual ~BHashGenerator();
    virtual bool get(ELEM &e);
    virtual void remove();
protected:
    BHashT *hash_table;
    int slot;
    GEN gen;
};

#endif /* _BHASH_H */
