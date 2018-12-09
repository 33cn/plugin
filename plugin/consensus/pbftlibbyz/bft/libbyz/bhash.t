/* Copyright (c) 1995 Andrew C. Myers */

#ifndef _BHASH_T
#define _BHASH_T

#include "bhash.h"

#include "th_assert.h"
#include <stdlib.h>
#include <unistd.h>

#define BH_PRECBITS (sizeof(int)<<3)

#define BH_PHI 1.618033989
/* The golden ratio */

#define BH_HASHMULT 1737350767
/* This is int(PHI * (1<<PRECBITS)), suggested by Knuth as a good hash
   multiplier.
*/

#define BH_MIN_SIZE 4
/* Size below which predict does not shrink the hash table */

// moved to header file
#if 0
#define BH_MAX_DESIRED_ALPHA 3
/* "alpha" is the ratio of items to the hash table to slots, as described
   in CLR, Chapter 12.
*/
#endif

#define BH_EXCESSIVE_RECIPROCAL_ALPHA 20
/* The point at which the hash table tries to shrink because it has
   become too sparse.
*/

BH_TEMPLATE bool BHashGenerator<ELEM, SET, GEN>::get(ELEM &e) {
    if (gen.get(e)) return TRUE;
    do {
	slot++;
	if (slot == hash_table->numSlots) {
#ifndef NDEBUG
	    // benevolent side-effects for debugging
	    slot = -1;
#endif
	    return FALSE;
	}
	gen = GEN(hash_table->buckets[slot]);
    } while (!gen.get(e));
    return TRUE;
}

BH_TEMPLATE void BHashGenerator<ELEM, SET, GEN>::remove() {
    gen.remove(); 
    hash_table->numItems--;
}

BH_TEMPLATE BHashGenerator<ELEM, SET, GEN>::~BHashGenerator() {}

BH_TEMPLATE inline int BHashT::do_hash(ELEM const &e) const {
    unsigned tmp = e.hash() * BH_HASHMULT;
    return tmp >> right_shift;
//  return  ((NUM)e.hash() * (NUM)BH_HASHMULT) >> right_shift
//  gcc generates a sra for this shorter code... oops!
}

#undef NUM

BH_TEMPLATE BHashT::BHash(int size_hint) {
    numItems = 0;
    sizeup(size_hint);
    buckets = new SET[numSlots];
}

BH_TEMPLATE BHashT::BHash() {
    numItems = 0;
    sizeup(1);
    buckets = new SET[numSlots];
}

BH_TEMPLATE BHashT::~BHash() {
    //int index = 0;
    //int ns = numSlots;
    delete [] buckets;
}

BH_TEMPLATE void BHashT::copy_items(BHashT const &bh) {
    numItems = 0;
    sizeup(bh.numItems);
    buckets = new SET[numSlots];
    BHashGenerator<ELEM, Buckets<ELEM>, BucketsGenerator<ELEM> >
	g((BHashT &)bh);
    ELEM e;
    while (g.get(e)) add(e);
}

BH_TEMPLATE BHashT::BHash(BHashT const &bh) {
     copy_items(bh);
}
 
BH_TEMPLATE BHashT const &BHashT::operator=(BHashT const &bh) {
    if (this == &bh) return *this;
    delete [] buckets;
    copy_items(bh);
    return *this;
}

BH_TEMPLATE int BHashT::size() const {
    return numItems;
}

BH_TEMPLATE void BHashT::add(ELEM const &e) {
    assert(!find(e));
    buckets[do_hash(e)].add(e);
    numItems++;
}

BH_TEMPLATE ELEM *BHashT::find_or_add(ELEM const &e) {
    SET &bucket = buckets[do_hash(e)];
    ELEM *e2 = bucket.find(e);
    if (e2) {
	return e2;
    } else {
	bucket.add(e);
	numItems++;
	return 0;
    }
}

BH_TEMPLATE ELEM *BHashT::find(ELEM const &e) const {
    return buckets[do_hash(e)].find(e);
}

BH_TEMPLATE ELEM *BHashT::find_fast(ELEM const &e) const {
    return buckets[do_hash(e)].find_fast(e);
}

BH_TEMPLATE bool BHashT::store(ELEM const &e) {
    bool result = buckets[do_hash(e)].store(e);
    if (!result) numItems++;
    return result;
}

BH_TEMPLATE bool BHashT::remove(ELEM &e) {
    bool result = buckets[do_hash(e)].remove(e);
    if (result) numItems--;
    return result;
}

BH_TEMPLATE void BHashT::remove_fast(ELEM &e) {
    numItems--;
    buckets[do_hash(e)].remove_fast(e);
}

BH_TEMPLATE void BHashT::clear() {
    int n = numSlots;
    for (int i = 0; i < n; i++) buckets[i].clear();
    numItems = 0;
}

// inline function moved to header file
#if 0
BH_TEMPLATE inline void BHashT::allowAutoResize() {
    predict(numItems / BH_MAX_DESIRED_ALPHA);
}
#endif

BH_TEMPLATE void BHashT::sizeup(int desired_size) {
    numSlots = 2;
    right_shift = BH_PRECBITS - 1;
// We start right_shift at BH_PRECBITS because a right_shift of zero leads
// to a shift right in do_hash that may go outside the allowed range,
// leading to implementation-defined results.
    while (numSlots < desired_size && right_shift > 0)
	{ numSlots <<= 1; right_shift--; }
	// invariant: numSlots = 2^(BH_PRECBITS - right_shift)
}

BH_TEMPLATE void BHashT::predict(int desired_size)
{
    int ns = numSlots;
    if (ns >= desired_size &&
	ns < desired_size * BH_EXCESSIVE_RECIPROCAL_ALPHA + BH_MIN_SIZE) return;
    if (right_shift == 0 && ns < desired_size) return;
	    // can't make it any bigger!

    resize(desired_size);
}


BH_TEMPLATE void BHashT::resize(int desired_size) {
    BHashT &self = (BHashT &)(*this);
    int old_slots = numSlots;
    SET *old_buckets = buckets;

    self.sizeup(desired_size);
    self.buckets = new SET[numSlots];
    self.numItems = 0;
    
    int i;
    for (i=0; i<old_slots; i++) {
	GEN g(old_buckets[i]);
	ELEM e;
	while (g.get(e)) self.add(e);
    }
    delete [] old_buckets;
}

extern "C" { float sqrtf(float); }

BH_TEMPLATE float BHashT::estimateEfficiency() const {
    int sx = 0;
    int n = 0;
    int ns = numSlots;
    for (int i=0; i<ns; i++) {
	bool nonempty = FALSE;
	GEN g(buckets[i]);
	ELEM e;
	while (g.get(e)) {
	    sx++;
	    nonempty = TRUE;
	}
	n += (nonempty?1:0);
    }
    return float(sx)/float(n);
}

BH_TEMPLATE float BHashT::estimateClumping() const {
    int i;
    double sx2=0.;
    int n=numSlots;
    float m = numItems;
    for (i=0; i<n; i++) {
	int x=0;
	GEN g(buckets[i]);
	ELEM e;
	while (g.get(e)) x++;
        sx2 += x * x;
    }
    return sx2/m - m/n;
}

BH_TEMPLATE int BHashT::memory_usage() const {
    return sizeof(*this) // the top level struct
	+ numSlots*sizeof(Buckets<ELEM>) // the array of pointers to chains
	+ numItems*Buckets<ELEM>::sizeof_BucketsImpl(); // actual buckets
}

#endif /* _BHASH_T */
