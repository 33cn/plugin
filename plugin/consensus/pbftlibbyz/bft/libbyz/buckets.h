#ifndef _BUCKETS_H
#define _BUCKETS_H

#define BK_TEMPLATE template <class ELEM>

BK_TEMPLATE struct BucketsImpl;
BK_TEMPLATE class BucketsGenerator;

BK_TEMPLATE class Buckets {
public:
/*
    A "Buckets" is a standard implementation of a SET (see "map.h")
    using a linked list of buckets. It is suitable for use with a "BHash".
*/
    Buckets();
    ~Buckets();
    Buckets(Buckets<ELEM> const &);
    void operator=(Buckets<ELEM> const &);
    ELEM *find(ELEM const &) const;
    ELEM *find_fast(ELEM const &) const;
    void add(ELEM const &);
    bool store(ELEM const &);
    bool remove(ELEM &);
    void remove_fast(ELEM &);

    void clear(void);
    static int sizeof_BucketsImpl();
    // returns the sizeof BucketsImpl being used.
protected:
    friend class BucketsGenerator<ELEM>;

    ELEM *find_loop(ELEM const &, BucketsImpl<ELEM> *) const;

    BucketsImpl<ELEM> *pairs;
#ifdef __GNUC__
    static BucketsGenerator<ELEM> *dummy; // force g++ expansion
    // this previous line actually makes cxx screw up for some reason
#endif
};

BK_TEMPLATE class BucketsImpl;

BK_TEMPLATE class BucketsGenerator {
public:
    BucketsGenerator(Buckets<ELEM> &buckets) {
	pairs = 0;
	last = &buckets.pairs;
    }
	virtual ~BucketsGenerator() {}
    virtual bool get(ELEM &e);
    void remove();
protected:
    BucketsImpl<ELEM> *pairs;
      // points to the current mapping, or 0 if the current mapping is
      // invalid (i.e. we are at the beginning or the current elt has been
      // removed.
    BucketsImpl<ELEM> **last;
      // points to the "next" field of the previous mapping
};

#endif /* _BUCKETS_H */
