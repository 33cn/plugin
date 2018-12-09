#ifndef _BUCKETS_T
#define _BUCKETS_T

#include "bhash.h"
#include "buckets.h"
#include "th_assert.h"

#pragma interface

#ifdef __GNUC__
#define FAST_ALLOC_BUCKETS FALSE
#else
#define FAST_ALLOC_BUCKETS TRUE
#endif
#define CHUNKSIZE 500

#define BucketsT Buckets<ELEM>
#define BucketsImplT BucketsImpl<ELEM>


BK_TEMPLATE struct BucketsImpl {
    BucketsImpl(ELEM const &e, BucketsImpl<ELEM> *n) : elem(e), next(n) {}
	virtual ~BucketsImpl() {}
    ELEM elem;
    BucketsImplT *next;

#if FAST_ALLOC_BUCKETS
    void *operator new(size_t);
    void operator delete(void *);
    static BucketsImpl<ELEM> *next_empty, *bufend, *last_freed;
#endif
};

#if FAST_ALLOC_BUCKETS
BK_TEMPLATE BucketsImplT *BucketsImplT::next_empty = 0;
BK_TEMPLATE BucketsImplT *BucketsImplT::bufend = 0;
BK_TEMPLATE BucketsImplT *BucketsImplT::last_freed = 0;
#endif

BK_TEMPLATE inline BucketsT::Buckets() { pairs = 0; }
BK_TEMPLATE inline BucketsT::~Buckets() { clear(); }

BK_TEMPLATE void BucketsT::clear(void) { 
  BucketsImplT *b = pairs;
  while (b) {
    BucketsImplT *n = b->next;
    delete b;
    b = n;
  }
  pairs = 0;
}

BK_TEMPLATE ELEM *BucketsT::find(ELEM const &e) const {
    /* making this inline seems to slow things by 10%! huh? */
    BucketsImplT *b = pairs;
    if (b) {
	if (b->elem.similar(e)) {
	    return &b->elem;
	} else {
	    return find_loop(e, b->next);
	}
    } else {
	return 0;
    }
}

BK_TEMPLATE inline
ELEM *BucketsT::find_loop(ELEM const &e, BucketsImplT *b) const {
    while (b) {
	if (b->elem.similar(e)) return &b->elem;
	b = b->next;
    }
    return 0;
}

BK_TEMPLATE ELEM *BucketsT::find_fast(ELEM const &e) const {
    BucketsImplT *b = pairs;
    while (!(b->elem.similar(e))) b = b->next;
    return &b->elem;
}

BK_TEMPLATE inline void BucketsT::add(ELEM const &e) {
#ifndef NDEBUG
    assert(!find(e));
#endif
    pairs = new BucketsImplT(e, pairs);
}

BK_TEMPLATE bool BucketsT::store(ELEM const &e) {
    BucketsImplT *b = pairs;
    while (b) {
	if (b->elem.similar(e)) {
	    b->elem = e;
	    return TRUE;
	}
	b = b->next;
    }
    add(e);
    return FALSE;
}

BK_TEMPLATE bool BucketsT::remove(ELEM &e) {
    BucketsImplT **last = &pairs;
    BucketsImplT *b = pairs;
    while (b) {
	if (b->elem.similar(e)) {
	    *last = b->next;
	    e = b->elem;
	    delete b;
	    return TRUE;
	}
	last = &b->next;
	b = *last;
    }
    return FALSE;
}

BK_TEMPLATE inline void BucketsT::remove_fast(ELEM &e) {
    BucketsImplT **last = &pairs;
    BucketsImplT *b = *last;
    while (!(b->elem.similar(e))) { last = &b->next; b = *last; }
    *last = b->next;
    e = b->elem;
    delete b;
}

BK_TEMPLATE void BucketsT::operator=(BucketsT const &buckets) {
    BucketsImplT **last = &pairs;
    BucketsImplT *b = buckets.pairs;
    pairs = 0;
    while (b) {
	BucketsImplT *copy = new BucketsImplT(b->elem, 0);
	*last = copy;
	last = &copy->next;
	b = b->next;
    }
}

BK_TEMPLATE int BucketsT::sizeof_BucketsImpl() { 
    return sizeof(BucketsImplT);
}

BK_TEMPLATE bool BucketsGenerator<ELEM>::get(ELEM &e) {
    BucketsImplT *b = pairs;
    if (b) last = &b->next;
    b = *last;
    pairs = b;
    if (b) {
	e = b->elem;
	return TRUE;
    } else {
	return FALSE;
    }
}

BK_TEMPLATE void BucketsGenerator<ELEM>::remove() {
    BucketsImplT *b = pairs->next;
    *last = b;
    delete pairs;
    pairs = 0;
}

#if FAST_ALLOC_BUCKETS
BK_TEMPLATE inline void *BucketsImplT::operator new(size_t s) {
    th_assert(s == sizeof(BucketsImplT), "Bad size passed to new");
    BucketsImplT *ret = next_empty;
    if (last_freed) {
	ret = last_freed;
	last_freed = ret->next;
    } else {
	if (ret != bufend) {
	    next_empty = ret + 1;
	} else {
	    ret = (BucketsImplT *)new char[CHUNKSIZE *
					   sizeof(BucketsImplT)];
	    th_assert(ret, "New failed");
	    bufend = ret + CHUNKSIZE;
	    next_empty = ret + 1;
	}
    }
    return ret;
}

BK_TEMPLATE inline void BucketsImplT::operator delete(void *x) {
    BucketsImplT *mx = (BucketsImplT *)x;
    mx->next = last_freed;
    last_freed = mx;
}
#endif

#undef CHUNKSIZE
#undef FAST_ALLOC_BUCKETS

#endif /* _BUCKETS_T */
