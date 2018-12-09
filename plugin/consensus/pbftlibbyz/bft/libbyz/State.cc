#include <stdio.h>
#include <stdlib.h>
#include <strings.h>
#include <limits.h>
#include <unistd.h>
#include <sys/mman.h>

//! JC: got rid of extern C to help compile
//extern "C" {
#include "gmp.h"
//}

#include "th_assert.h"
#include "State.h"
#include "Replica.h"
#include "Fetch.h"
#include "Data.h"
#include "Meta_data.h"
#include "Meta_data_d.h"
#include "Meta_data_cert.h"
#include "MD5.h"
#include "map.h"
#include "Array.h"
#include "valuekey.h"

#include "Statistics.h"
#include "State_defs.h"


// Force template instantiation
#include "Array.t"
#include "Log.t"
#include "bhash.t"
#include "buckets.t"
template class Log<Checkpoint_rec>;

#ifdef NO_STATE_TRANSLATION
// External pointers to memory and bitmap for macros that check cow
// bits.
unsigned long *_Byz_cow_bits = 0;
char *_Byz_mem = 0;
#endif

//
// The memory managed by the state abstraction is partitioned into
// blocks.
//
struct Block {
#ifdef NO_STATE_TRANSLATION
  char data[Block_size];
#else
  char *data;
  int size;
#endif  
  
  inline Block() {
#ifndef NO_STATE_TRANSLATION
    data = NULL;
    size = 0;
#endif
  }

  inline Block(Block const & other) {
#ifndef NO_STATE_TRANSLATION
    size = other.size;
    data = new char[size];
    memcpy(data, other.data, size);	  
#else
    memcpy(data, other.data, Block_size);
#endif
  }
  
#ifndef NO_STATE_TRANSLATION
  inline ~Block() { if (data) delete [] data; }

  inline void Block::init_from_ptr(char *ptr, int psize) {
    if (data) delete [] data;
    data = ptr;
    size = psize;  
  }
  
#endif
  
  inline Block& operator=(Block const &other) {
    if (this == &other) return *this;
#ifndef NO_STATE_TRANSLATION
    if (size != other.size) {
      if (data)
	delete [] data;
      data = new char[other.size];
    }
    size = other.size; 
    memcpy(data, other.data, other.size);
#else
    memcpy(data, other.data, Block_size);
#endif
    return *this;
  }

  inline Block& operator=(char const *other) {
    if (this->data == other) return *this;
#ifndef NO_STATE_TRANSLATION
    if (size != Block_size) {
      if (data)
    	  delete [] data;
      data = new char[Block_size];
    }
    size = Block_size; 
#endif
    memcpy(data, other, Block_size);
    return *this;
  }  
};


// Blocks are grouped into partitions that form a hierarchy.
// Part contains information about one such partition.
struct Part {
  Seqno lm;  // Sequence number of last checkpoint that modified partition
  Digest d;  // Digest of partition

#ifndef NO_STATE_TRANSLATION
  int size; // Size of object for level 'PLevels-1'
#endif
  
  Part() { lm = 0; }
};

// Information about stale partitions being fetched.
struct FPart {
  int index;
  Seqno lu; // Latest checkpoint seqno for which partition is up-to-date
  Seqno lm; // Sequence number of last checkpoint that modified partition
  Seqno c;  // Sequence number of checkpoint being fetched 
  Digest d; // Digest of checkpoint being fetched
#ifndef NO_STATE_TRANSLATION
  int size; // Size of leaf object
#endif
};

class FPartQueue : public Array<FPart> {};

// Information about partitions whose digest is being checked.
struct CPart {
  int index;
  int level;
};
class CPartQueue : public Array<CPart> {};


// Copy of leaf partition (used in checkpoint records)
struct BlockCopy : public Part {
  Block data; // Copy of data at the time the checkpoint was taken

  BlockCopy() : Part() {}

#ifndef NO_STATE_TRANSLATION
  BlockCopy(char *d, int sz) : Part() { data.init_from_ptr(d, sz); }
#endif
};


//
// Checkpoint records.
//

// Key for partition map in checkpoint records
class  PartKey {  
public:                                        
  inline PartKey() {}
  inline PartKey(int l, int i) : level(l), index(i) {}

  inline void operator=(PartKey const &x) { 
    level = x.level;
    index = x.index;  
  }

  inline int hash() const { 
    return index << (PLevelSize[PLevels-1]+level); 
  }

  inline bool operator==(PartKey const &x) { 
    return (level == x.level) && (index == x.index);
  }
  
  int level;
  int index;
};


// Checkpoint record  
class Checkpoint_rec {
public:
  Checkpoint_rec();
  // Effects: Creates an empty checkpoint record.

  ~Checkpoint_rec();
  // Effects: Deletes record an all parts it contains
  
  void clear();
  // Effects: Deletes all parts in record and removes them.

  bool is_cleared();
  // Effects: Returns true iff Checkpoint record is not in use.

  void append(int l, int i, Part *p);
  // Requires: fetch(l, i) == 0
  // Effects: Appends partition index "i" at level "l" with value "p"
  // to the record.

  void appendr(int l, int i, Part *p);
  // Effects: Like append but without the requires clause. If fetch(l,
  // i) != 0 it retains the old mapping.
  
  Part* fetch(int l, int i);
  // Effects: If there is a partition with index "i" from level "l" in
  // this, returns a pointer to its information. Otherwise, returns 0.
 
  int num_entries() const;
  // Effects: Returns the number of entries in the record.

  class Iter {
  public:
    inline Iter(Checkpoint_rec* r) : g(r->parts) {}
    // Effects: Return an iterator for the partitions in r.
    
    inline bool get(int& level, int& index, Part*& p) {
      // Effects: Modifies "level", "index" and "p" to contain
      // information for the next partition in "r" and returns
      // true. Unless there are no more partitions in "r" in which case
      // it returns false.
      PartKey k;
      if (g.get(k, p)) {
	level = k.level;
	index = k.index;
	return true;
      }
      return false;
    }
    
  private:
    MapGenerator<PartKey,Part*> g;
  };
  friend class Iter;

  void print();
  // Effects: Prints description of this to stdout

  Digest sd; // state digest at the time the checkpoint is taken 

private:
  // Map for partitions that were modified since this checkpoint was
  // taken and before the next checkpoint.
  Map<PartKey,Part*> parts;
};


inline Checkpoint_rec::Checkpoint_rec() : parts(256) {}


inline Checkpoint_rec::~Checkpoint_rec() {
  clear();
}


inline void Checkpoint_rec::append(int l, int i, Part *p) {
  th_assert(!parts.contains(PartKey(l,i)), "Invalid state");
  parts.add(PartKey(l,i), p);
}
  
  
inline void Checkpoint_rec::appendr(int l, int i, Part *p) {
  if (parts.contains(PartKey(l,i)))
    return;

  append(l, i, p);
}


inline Part* Checkpoint_rec::fetch(int l, int i) {
  Part* p;
  if (parts.find(PartKey(l,i), p)) {
    return p ;
  }
  return 0;
}


inline bool Checkpoint_rec::is_cleared() { 
  return sd.is_zero(); 
}


inline int Checkpoint_rec::num_entries() const {
  return parts.size();
}


void Checkpoint_rec::print() {
  printf("Checkpoint record: %d blocks \n", parts.size()); 
  MapGenerator<PartKey,Part*> g(parts);
  PartKey k;
  Part* p;
  while (g.get(k, p)) {
    printf("Block: level= %d index=  %d  ", k.level, k.index);
    printf("last mod=%qd ", p->lm);
    p->d.print();
    printf("\n");
  }
}

void Checkpoint_rec::clear() {
  if (!is_cleared()) {
    MapGenerator<PartKey,Part*> g(parts);
    PartKey k;
    Part* p;
    while (g.get(k, p)) {
      if (k.level == PLevels-1) {
	//	/* debug */ fprintf(stderr, "Clearing leaf %d\t", k.index);
	delete ((BlockCopy*)p);
      }
      else
	delete p;
      g.remove();
    }
    sd.zero();
  }
}

#ifndef NO_STATE_TRANSLATION

// Page Mapping: holds a set of pages that were fetched
class Page_mapping {
public:
  Page_mapping() {}
  // Effects: Creates an empty Page Mapping record.

  ~Page_mapping();
  // Effects: Deletes record an all parts it contains
  
  void clear();
  // Effects: Deletes all parts in record and removes them.

  void append(int l, int i, BlockCopy *b);
  // Requires: fetch(l, i) == 0
  // Effects: Appends partition index "i" at level "l" with value "p"
  // to the record.

  void append_or_replace(int l, int i, BlockCopy *b);
  // Effects: Like append but without the requires clause. If fetch(l,
  // i) != 0 it replaces the old mapping, deleting it.
  
  void append_or_replace_without_delete(int l, int i, BlockCopy *b);
  // Effects: Like append but without the requires clause. If fetch(l,
  // i) != 0 it replaces the old mapping, WITHOUT deleting it.
  
  BlockCopy* fetch(int l, int i);
  // Effects: If there is a partition with index "i" from level "l" in
  // this, returns a pointer to its information. Otherwise, returns 0.
 
  int num_entries() const;
  // Effects: Returns the number of entries in the record.

  class Iter {
  public:
    inline Iter(Page_mapping* r) : g(r->parts) {}
    // Effects: Return an iterator for the partitions in r.
    
    inline bool get(int& level, int& index, BlockCopy*& p) {
      // Effects: Modifies "level", "index" and "p" to contain
      // information for the next partition in "r" and returns
      // true. Unless there are no more partitions in "r" in which case
      // it returns false.
      PartKey k;
      if (g.get(k, p)) {
	level = k.level;
	index = k.index;
	return true;
      }
      return false;
    }
    
  private:
    MapGenerator<PartKey,BlockCopy*> g;
  };
  friend class Iter;

  void print();
  // Effects: Prints description of this to stdout

private:
  // Map for objects that have been fetched
  Map<PartKey,BlockCopy*> parts;
};


Page_mapping::~Page_mapping() {
  clear();
}


inline void Page_mapping::append(int l, int i, BlockCopy *b) {
  th_assert(!parts.contains(PartKey(l,i)), "Invalid state");
  parts.add(PartKey(l,i), b);
}
    
inline void Page_mapping::append_or_replace(int l, int i, BlockCopy *b) {
  BlockCopy *tmp;
  if (parts.remove(PartKey(l,i), tmp))
    delete tmp;
  append(l, i, b);
}

inline void Page_mapping::append_or_replace_without_delete(int l, int i, BlockCopy *b) {
  BlockCopy *tmp;
  parts.remove(PartKey(l,i), tmp);
  append(l, i, b);
}


inline BlockCopy* Page_mapping::fetch(int l, int i) {
  BlockCopy* b;
  if (parts.find(PartKey(l,i), b)) {
    return b ;
  }
  return 0;
}


void Page_mapping::clear() {
  MapGenerator<PartKey,BlockCopy*> g(parts);
  PartKey k;
  BlockCopy* p;
  while (g.get(k, p)) {
    delete p;
  }
  parts.clear();
}


inline int Page_mapping::num_entries() const {
  return parts.size();
}

void Page_mapping::print() {
  printf("Page Map: %d blocks \n", parts.size()); 
  MapGenerator<PartKey,BlockCopy*> g(parts);
  PartKey k;
  BlockCopy* p;
  while (g.get(k, p)) {
    printf("Block: level= %d index=  %d  ", k.level, k.index);
    printf("last mod=%qd ", p->lm);
    p->d.print();
    printf("\n");
  }
}

#endif


//
// Sums of digests modulo a large integer
//
struct DSum {
  static const int nbits = 256;
  static const int mp_limb_bits = sizeof(mp_limb_t)*8;
  static const int nlimbs = (nbits+mp_limb_bits-1)/mp_limb_bits;
  static const int nbytes = nlimbs*sizeof(mp_limb_t);
  static DSum* M; // Modulus for sums must be at most nbits-1 long.

  mp_limb_t sum[nlimbs];
  //  char dummy[56];

  inline DSum() { bzero(sum, nbytes); }
  // Effects: Creates a new sum object with value 0

  inline DSum(DSum const & other) {
    memcpy(sum, other.sum, nbytes);
  }
  
  inline DSum& operator=(DSum const &other) {
    if (this == &other) return *this;
    memcpy(sum, other.sum, nbytes);
    return *this;
  }

  void add(Digest &d);
  // Effects: adds "d" to this

  void sub(Digest &d);
  // Effects: subtracts "d" from this.
};

//
// PageCache: LRU cache of copies of objects at the last checkpoint taken
//            after the object was modified.
//
class PageCache {

public:

  PageCache() { lru = mru = NULL; parts.predict(MaxElems); }
  // Effects: Creates an empty PageCache record.

  ~PageCache();
  // Effects: Deletes record an all parts it contains
  
  void clear();
  // Effects: Deletes all parts in record and removes them.

  void append(int i, BlockCopy *b);
  // Effects: Appends object index "i" with value "b" to the record.
  //          Updates LRU info and evicts LRU entry if the size of
  //          the cache has been exceeded

  BlockCopy* fetch_and_remove(int i);
  // Effects: If there is an obj with index "i" in
  // this, returns a pointer to its information. Otherwise, returns 0.
  // removes the object from cache and its index from the LRU list.
 
  int num_entries() const;
  // Effects: Returns the number of entries in the record.

  char* get_obj(int i);
  // Effects: if we hold a copy of object with index "i" return it,
  //          otherwise returns NULL

  class Iter {
  public:
    inline Iter(PageCache* r) : g(r->parts) {}
    // Effects: Return an iterator for the partitions in r.
    
    inline bool get(int& index, BlockCopy*& p) {
      // Effects: Modifies "level", "index" and "p" to contain
      // information for the next partition in "r" and returns
      // true. Unless there are no more partitions in "r" in which case
      // it returns false.
      IntKey k;
      if (g.get(k, p)) {
	index = (int)k.val;
	return true;
      }
      return false;
    }
    
  private:
    MapGenerator<IntKey,BlockCopy*> g;
  };
  friend class Iter;

  void print();
  // Effects: Prints description of this to stdout

  struct DoublyLL {
    int elem;
    DoublyLL *prev, *next;
  };

private:
  // Cache size:
  static const int MaxElems = 100;
  // Map for copies of objects:
  Map<IntKey,BlockCopy*> parts;
  DoublyLL *lru, *mru;
};


PageCache::~PageCache() {
  clear();
}


inline void PageCache::append(int i, BlockCopy *b) {
  BlockCopy *tmp;
  if (parts.remove(i, tmp)) {
    // The object was already present in the cache. Delete previous copy,
    // and update LRU list so that this obj is now the MRU entry
    delete tmp;
    if (mru && mru->elem != i) {
      DoublyLL *ptr = mru->next;
      while (ptr) {
	if (ptr == lru)
	  lru = ptr->prev;
	if (ptr->elem == i) {
	  if (ptr->prev)
	    ptr->prev->next = ptr->next;
	  if (ptr->next)
	    ptr->next->prev = ptr->prev;
	  ptr->next = mru;
	  ptr->prev = NULL;
	  mru->prev = ptr;
	  mru = ptr;
	  ptr = NULL;
	}
	else
	  ptr = ptr->next;
      }
    }
  }
  else { /* The obj was not in the cache. If full evict someone */
      
    if (parts.size() == MaxElems) {
      if (parts.remove(lru->elem, tmp))
	delete tmp;
      else
	fprintf(stderr, "Wrong LRU?");
      DoublyLL *ptr = lru->prev;
      ptr->next = NULL;
      delete lru;
      lru = ptr;
    }
    if (!mru) {
      th_assert(!lru, "Wrong LRU?");
      mru = lru = new DoublyLL;
      mru->elem = i;
      mru->next = NULL;
      mru->prev = NULL;
    }
    else {
      DoublyLL *ptr = new DoublyLL;
      ptr->elem = i;
      ptr->next = mru;
      ptr->prev = NULL;
      mru->prev = ptr;
      mru = ptr;
    }
  }
  parts.add(i, b);
}
    
inline char* PageCache::get_obj(int i) {
  BlockCopy *b;
  if (parts.find(i, b))
    return b->data.data;
  else
    return NULL;
}

inline BlockCopy* PageCache::fetch_and_remove(int i) {
  BlockCopy* b;
  if (parts.remove(i, b)) {
     DoublyLL *ptr = mru;
    while (ptr) {
      if (ptr->elem == i) {
	if (ptr==mru)
	  mru = ptr->next;
	else
	  ptr->prev->next = ptr->next;
	if (ptr==lru)
	  lru = ptr->prev;
	else
	  ptr->next->prev = ptr->prev;
	delete ptr;
	ptr = NULL;
      }
      else {
	ptr = ptr->next;
	th_assert(ptr, "not found in LRU list");
      }
    }
   return b ;    
  }
  return 0;
}


void PageCache::clear() {
  MapGenerator<IntKey,BlockCopy*> g(parts);
  IntKey k;
  BlockCopy* p;
  while (g.get(k, p)) {
    delete p;
  }
  parts.clear();
  lru = mru = NULL;
}


inline int PageCache::num_entries() const {
  return parts.size();
}

void PageCache::print() {
  printf("Page cache: %d blocks \n", parts.size()); 
  MapGenerator<IntKey,BlockCopy*> g(parts);
  IntKey k;
  BlockCopy* p;
  while (g.get(k, p)) {
    printf("Block: index=  %d  ", k.val);
    printf("last mod=%qd ", p->lm);
    p->d.print();
    printf("\n");
  }
}


DSum* DSum::M = 0;

inline void DSum::add(Digest &d) { 
  mp_limb_t ret = mpn_add((mp_ptr) sum, (mp_srcptr) sum, nlimbs, 
			  (mp_srcptr)d.udigest(), sizeof(d)/sizeof(mp_limb_t));
  th_assert(ret == 0, "sum and d should be such that there is no carry");

  if (mpn_cmp(sum, M->sum, nlimbs) >= 0) {
    mpn_sub((mp_ptr) sum, (mp_srcptr) sum, nlimbs, (mp_srcptr)M->sum, nlimbs);
  }
}


void  DSum::sub(Digest &d) {
  int dlimbs = sizeof(d)/sizeof(mp_limb_t);
  bool gt = false;
  for (int i = nlimbs-1; i >= dlimbs; i--) {
    if (sum[i] != 0) {
      gt = true;
      break;
    }
  }

  mp_limb_t ret;
  if (!gt && mpn_cmp(&sum[dlimbs-1], (mp_limb_t*)d.udigest(), dlimbs) < 0) {
    ret = mpn_add((mp_ptr) sum, (mp_srcptr) sum, nlimbs, (mp_srcptr)M->sum, nlimbs);
    th_assert(ret == 0, "There should be no carry");
  }

  ret = mpn_sub((mp_ptr) sum, (mp_srcptr) sum, nlimbs, 
		(mp_srcptr)d.udigest(), sizeof(d)/sizeof(mp_limb_t));
  th_assert(ret == 0, "sum and d should be such that there is no borrow");

}


//
// State methods:
//


#ifndef NO_STATE_TRANSLATION
State::State(Replica *rep, int num_objs,
	     int (*gets)(int, char **),
	     void (*puts)(int, int *, int *, char **),
	     void (*shutdown_p)(FILE *o),
	     void (*restart_p)(FILE *i)) : 
  replica(rep), nb(num_objs+replica->used_state_pages()), 
  cowb(nb), clog(max_out*2, 0), lc(0), get_segment(gets), put_segments(puts),
  shutdown_proc(shutdown_p), restart_proc(restart_p)
{

#else

State::State(Replica *rep, char *memory, int num_bytes) : 
  replica(rep), mem((Block*)memory), nb(num_bytes/Block_size), 
  cowb(nb), clog(max_out*2, 0), lc(0), last_fetch_t(0){
#endif

#ifndef NO_STATE_TRANSLATION
  next_chunk = 0;
  reassemb = NULL;
  total_size = 0;

  rep_mem = new Block[replica->used_state_pages()];
  for (int i=0; i<replica->used_state_pages(); i++)
    rep_mem[i].init_from_ptr(rep->rep_info_mem() + i * Block_size, Block_size);
#endif

  for (int i=0; i < PLevels; i++) {
    ptree[i] = new Part[(i != PLevels-1) ? PLevelSize[i] : nb];
    stalep[i] = new FPartQueue;
  }

  // The random modulus for computing sums in AdHASH.
  DSum::M = new DSum;
  mpn_set_str(DSum::M->sum,
    (unsigned char*)"d2a10a09a80bc599b4d60bbec06c05d5e9f9c369954940145b63a1e2",
	      DSum::nbytes, 16);

  if (sizeof(Digest)%sizeof(mp_limb_t) != 0)
    th_fail("Invalid assumption: sizeof(Digest)%sizeof(mp_limb_t)");

  for (int i=0; i < PLevels-1; i++) {
    stree[i] = new DSum[PLevelSize[i]];
  }

  fetching = false;
#ifndef NO_STATE_TRANSLATION
  fetched_pages = new Page_mapping();
  pages_lc = new PageCache;
#endif
  cert = new Meta_data_cert;
  lreplier = 0;

  to_check = new CPartQueue;
  checking = false;
  refetch_level = 0;

#ifdef NO_STATE_TRANSLATION
  // Initialize external pointers to memory and bitmap for macros that 
  // check cow bits.
  _Byz_cow_bits = cowb.bitvec();
  _Byz_mem = (char*)mem;
#endif

}


State::~State() {
  for (int i=0; i < PLevels; i++) {
    delete [] ptree[i];
    delete stalep[i];
  }
  delete cert;
  delete to_check;
#ifndef NO_STATE_TRANSLATION
  delete fetched_pages;
#endif
}


void State::cow_single(int i) {

  BlockCopy* bcp;
  //  fprintf(stderr,"modifying %d\n",i);
  th_assert(i >= 0 && i < nb, "Invalid argument");
  //  th_assert(!cowb.test(i), "Invalid argument");

#ifndef NO_STATE_TRANSLATION
  if (cowb.test(i)) return;
#endif
  INCR_OP(num_cows);
  START_CC(cow_cycles);
  // Append a copy of the block to the last checkpoint
  Part& p = ptree[PLevels-1][i];

#ifndef NO_STATE_TRANSLATION
  if (i>=nb-replica->used_state_pages())
    bcp = new BlockCopy;
#else
  bcp = new BlockCopy;
#endif


#ifndef NO_STATE_TRANSLATION

  if (i<nb-replica->used_state_pages())  // application state
  {
    bcp = pages_lc->fetch_and_remove(i);
    if (!bcp || bcp->lm != ptree[PLevels-1][i].lm) {
      INCR_OP(cache_misses);
      if (bcp) {
	//	printf("wrong %d.", bcp->lm);
	delete bcp;
      }
      char *data;
      int size = get_segment(i, &data);
      bcp = new BlockCopy(data, size);
    }
    else
      INCR_OP(cache_hits);
  }
  else                               // replication library state
    bcp->data = rep_mem[i-nb+replica->used_state_pages()];

#else   // ifndef NO_STATE_TRANSLATION
  bcp->data = mem[i];
#endif
  bcp->lm = p.lm;
  bcp->d = p.d;

  //  fprintf(stderr, "Estou a apendar o i=%d ao lc=%d\n",i,lc);
  clog.fetch(lc).append(PLevels-1, i, bcp);
  cowb.set(i);

  STOP_CC(cow_cycles);
}


void State::cow(char *m, int size) {

#ifdef NO_STATE_TRANSLATION
  th_assert(m > (char*)mem && m+size <= (char*)mem+nb*Block_size, 
  	    "Invalid argument");

  if (size <= 0) return;
  
  // Find index of low and high block 
  int low = (m-(char*)mem)/Block_size;
  int high = (m+size-1-(char*)mem)/Block_size;

  for (int bindex = low; bindex <= high; bindex++) {
    // If cow bit is set do not perform copy.
    if (cowb.test(bindex)) continue;
    cow_single(bindex);
    }
#endif
}


void State::digest(Digest& d, int i, Seqno lm, char *data, int size) {
  // Compute digest for partition p:
  // MD5(i, last modification seqno, (data,size)
  MD5_CTX ctx;
  MD5Init(&ctx);
  MD5Update(&ctx, (char*)&i, sizeof(i));
  MD5Update(&ctx, (char*)&lm, sizeof(Seqno));
  MD5Update(&ctx, data, size);
  MD5Final(d.udigest(), &ctx);
}


inline int State::digest(Digest& d, int l, int i) {
  char *data;
  int size;

#ifndef NO_STATE_TRANSLATION
  BlockCopy* bcp = NULL;
#endif

  if (l == PLevels-1) {
    th_assert(i >= 0 && i < nb, "Invalid argument");

#ifdef NO_STATE_TRANSLATION
    data = mem[i].data;
    size = Block_size;
#else
    
    if (i<nb-replica->used_state_pages()) {  // application state
      size = get_segment(i, &data);
      bcp = new BlockCopy(data, size);
      bcp->lm = ptree[PLevels-1][i].lm;

      pages_lc->append(i, bcp);
      /*
	if (pages_lc[i])
	  delete pages_lc[i];
	pages_lc[i] = bcp;  */

    }
    else {                              // replication library state
      data = rep_mem[i-nb+replica->used_state_pages()].data;
      size = Block_size;
    }
#endif
  } else {
    data = (char *)(stree[l][i].sum);
    size = DSum::nbytes;
  }

  digest(d, i, ptree[l][i].lm, data, size);

#ifndef NO_STATE_TRANSLATION
  if (bcp != NULL)
    bcp->d = d;
#endif

  return size;

}


void State::compute_full_digest() {
  Cycle_counter cc;
  cc.start();
  int np = nb;
  for (int l = PLevels-1; l > 0; l--) {
    for (int i = 0; i < np; i++) {
      Digest &d = ptree[l][i].d;
#ifndef NO_STATE_TRANSLATION
      ptree[l][i].size = digest(d, l, i);
#else
      digest(d, l, i);
#endif
      stree[l-1][i/PSize[l]].add(d);
    }
    np = (np+PSize[l]-1)/PSize[l];
  }

  Digest &d = ptree[0][0].d;
  digest(d, 0, 0);
  
  cowb.clear();
  clog.fetch(0).clear();
  checkpoint(0);
  cc.stop();

  printf("Compute full digest elapsed %qd\n", cc.elapsed());
}


void State::update_ptree(Seqno n) {
  Bitmap* mods[PLevels];
  for (int l=0; l < PLevels-1; l++) {
     mods[l] = new Bitmap(PLevelSize[l]);
  }
  mods[PLevels-1] = &cowb;

  Checkpoint_rec& cr = clog.fetch(lc);

  for (int l = PLevels-1; l > 0; l--) {
    Bitmap::Iter iter(mods[l]);
    unsigned int i;
    while (iter.get(i)) {
      Part& p = ptree[l][i];
      DSum& psum = stree[l-1][i/PSize[l]];
      if (l < PLevels-1) {
	// Append a copy of the partition to the last checkpoint
	Part* np = new Part;
	np->lm = p.lm;
	np->d = p.d;
	cr.append(l, i, np);
      }
      
      // Subtract old digest from parent sum
      psum.sub(p.d);

      // Update partition information
      p.lm = n;
#ifndef NO_STATE_TRANSLATION
      p.size = digest(p.d, l, i);
#else
      digest(p.d, l, i);
#endif
      // Add new digest to parent sum
      psum.add(p.d);

      // Mark parent modified
      mods[l-1]->set(i/PSize[l]);
    }
  }

  if (mods[0]->test(0)) {
    Part& p = ptree[0][0];

    // Append a copy of the root partition to the last checkpoint
    Part* np = new Part;
    np->lm = p.lm;
    np->d = p.d;
    cr.append(0, 0, np);

    // Update root partition.
    p.lm = n;
    digest(p.d, 0, 0);
  }

  for (int l=0; l < PLevels-1; l++) {
     delete mods[l];
  }
}


void State::checkpoint(Seqno seqno) {
  //  printf("Checkp  ");
  INCR_OP(num_ckpts);
  START_CC(ckpt_cycles);

  update_ptree(seqno);

  lc = seqno;
  Checkpoint_rec &nr = clog.fetch(seqno);
  nr.sd = ptree[0][0].d;
 
  //  printf("\n");
  cowb.clear();

  STOP_CC(ckpt_cycles);
}


Seqno State::rollback() {
  //  fprintf(stderr,"ROLLBACK!       >>>>>>>>>> lc %ld <<<<<<<\n", lc);
  th_assert(lc >= 0 && !fetching, "Invalid state");

  INCR_OP(num_rollbacks);
  START_CC(rollback_cycles);

  // Roll back to last checkpoint.
  Checkpoint_rec& cr = clog.fetch(lc);
 
  Bitmap::Iter iter(&cowb);
  unsigned int i;

#ifdef NO_STATE_TRANSLATION
  while (iter.get(i)) {
    BlockCopy* b = (BlockCopy*)cr.fetch(PLevels-1, i);
    mem[i] = b->data;
  }
#else

  int *indices = new int[cowb.total_set()];
  char **pages = new char*[cowb.total_set()];
  int *sizes = new int[cowb.total_set()];
  int index=0;
  while (iter.get(i)) {
    BlockCopy* b = (BlockCopy*)cr.fetch(PLevels-1, i);

    if ((int)i<nb-replica->used_state_pages()) {  // application state
      indices[index] = (int)i;
      sizes[index] = b->data.size;
      pages[index++] = b->data.data;
    }
    else                               // replication library state
      rep_mem[i-nb+replica->used_state_pages()] = b->data;
  }
  put_segments(index, sizes, indices, pages);
  /*  printf("\nSending: ");
  for(int k=0;k<index;k++)
    printf("%d ",indices[k]);
    printf("\n"); */

#endif

  //  fprintf(stderr, "RB is clearing %qd\n", lc);
  cr.clear();
  cowb.clear();
  cr.sd = ptree[0][0].d;

#ifndef NO_STATE_TRANSLATION
  delete [] indices;
  delete [] sizes;
  delete [] pages;
#endif

  STOP_CC(rollback_cycles);
  
  return lc;
}


bool State::digest(Seqno n, Digest &d) {
  if (!clog.within_range(n)) 
    return false;

  Checkpoint_rec &rec = clog.fetch(n);
  if (rec.sd.is_zero())
    return false;

  d = rec.sd;
  return true;
}   

 
void State::discard_checkpoint(Seqno seqno, Seqno le) {
  if (seqno > lc && le > lc && lc >= 0) {
    checkpoint(le);
    lc = -1;
  }

  //  fprintf(stderr, "discard_chkpt: Truncating clog to %qd\n", seqno);
  clog.truncate(seqno);
}

#ifndef NO_STATE_TRANSLATION
char* State::get_cached_obj(int i) {
   return pages_lc->get_obj(i);
}
#endif

//
// Fetching missing state:
//  
#ifndef NO_STATE_TRANSLATION
char* State::get_data(Seqno c, int i, int &objsz) {
#else
char* State::get_data(Seqno c, int i) {
#endif

#ifndef NO_STATE_TRANSLATION
  char *data;


  // The following variables hold all the information about the last object
  // for which there was a get_data. Due to fragmentation of large objects,
  // this is likely to be the same as the one that is being requested now,
  // and this way we a void requesting the same object to the application
  // several times.

  static char *last_obj = NULL; // copy of the last object that was requested
  static int last_sz, last_i = -1, last_c;    // values of the object size, i 
                              // and c for the last object that was requested
#endif

  th_assert(clog.within_range(c) && i >= 0 && i < nb, "Invalid argument");

  if (ptree[PLevels-1][i].lm <= c && !cowb.test(i)) {

#ifdef NO_STATE_TRANSLATION
    return mem[i].data;
#else
    if (i<nb-replica->used_state_pages()) {  // application state
      // check if we already have a copy of the object from last invocation
      if (i == last_i && c == last_c) {  
	// We do, so just return the last object that we fetched
	objsz = last_sz;
	//	fprintf(stderr, "Obj %d, c=%d is CACHED \t",i, c);
	return last_obj;
      }

      // we don't have a copy of the object from last invocation. Delete the
      // last object that was fetched and use "get" to obtain the requested obj
      if (last_obj)
	delete [] last_obj;
      //      fprintf(stderr, "Replacing Obj %d, c=%d for (i,c) = (%d,%d) \t",last_i, last_c, i, c);
      objsz = get_segment(i, &data);
      last_i = i;
      last_c = c;
      last_sz = objsz;
      last_obj = data;
    }
    else {                              // replication library state
      objsz = Block_size;
      return rep_mem[i-nb+replica->used_state_pages()].data;
    }
    return data;
#endif
  }

  for (; c <= lc; c += checkpoint_interval) {
    Checkpoint_rec& r =  clog.fetch(c);
    
    // Skip checkpoint seqno if record has no state.
    if (r.sd.is_zero()) continue;
    
    Part *p = r.fetch(PLevels-1, i);
    if (p) {
#ifndef NO_STATE_TRANSLATION
      if (i<nb-replica->used_state_pages())
    	  objsz = ((BlockCopy*)p)->data.size;
      else
	objsz = Block_size;
#endif
     return ((BlockCopy*)p)->data.data;
    }
  }
  th_assert(0, "Invalid state");
  return NULL;
}


Part& State::get_meta_data(Seqno c, int l, int i) {
  th_assert(clog.within_range(c), "Invalid argument");

  //  fprintf(stderr, "\nIN ptree.lm = %qd, c = %qd. ", ptree[l][i].lm, c);

  Part& p = ptree[l][i];
  if (p.lm <= c)
    return p;

  for (; c <= lc; c += checkpoint_interval) {
    //    fprintf(stderr, "I1 ptree.lm = %qd, c = %qd. ", ptree[l][i].lm, c);
    Checkpoint_rec& r =  clog.fetch(c);
    //    fprintf(stderr, "I2 ptree.lm = %qd, c = %qd. ", ptree[l][i].lm, c);

    // Skip checkpoint seqno if record has no state.
    if (r.sd.is_zero()) continue;

    Part *p = r.fetch(l, i);
    //    fprintf(stderr, "I3 ptree.lm = %qd, c = %qd. ", ptree[l][i].lm, c);
    if (p) {
      return *p;
    }
  }
  th_assert(0, "Invalid state");
  return p; // never reached
}


void State::start_fetch(Seqno le, Seqno c, Digest *cd, bool stable) {

  START_CC(fetch_cycles);

  if (!fetching) {

#ifndef NO_STATE_TRANSLATION
    th_assert(fetched_pages->num_entries()==0,"Fetched Pages map not cleared");
#endif
    INCR_OP(num_fetches);

    fetching = true;
    //XXXXXXit should be like this  keep_ckpts = lc >= 0; see if I can avoid keeping
    // checkpoints if I am fetching a lot of stuff I can allocate more memory than the
    // system can handle.
    keep_ckpts = false;
    lreplier = lrand48()%replica->n();

    //    printf("Starting fetch ...");

    // Update partition information to reflect last modification
    // rather than last checkpointed modification.
    if (lc >= 0 && lc < le)
      checkpoint(le);

    // Initialize data structures.
    cert->clear();
    for (int i=0; i < PLevels; i++) {
      stalep[i]->clear();
    }

    // Start by fetching root information.
    flevel = 0;
    stalep[0]->_enlarge_by(1);
    FPart& p = stalep[0]->high();
    p.index = 0;
    p.lu = ((refetch_level == PLevels) ? -1 : lc);
    p.lm = ptree[0][0].lm;
    p.c = c;
    if (cd) 
      p.d = *cd;

    STOP_CC(fetch_cycles);
    send_fetch(true);
  }
  STOP_CC(fetch_cycles);
}


void State::send_fetch(bool change_replier) {
  START_CC(fetch_cycles);

  last_fetch_t = currentTime();
  Request_id rid = replica->new_rid();
  //  printf("rid = %qd\n", rid);
  replica->principal()->set_last_fetch_rid(rid);

  th_assert(stalep[flevel]->size() > 0, "Invalid state");
  FPart& p = stalep[flevel]->high();

  int replier = -1;   
  if (p.c >= 0) {
    // Select a replier.
    if (change_replier) {
      do {
	lreplier = (lreplier+1)%replica->n();
      } while (lreplier == replica->id());
    }
    replier = lreplier;
  }

#ifdef PRINT_STATS
  if (checking && ptree[flevel][p.index].lm > check_start) {
    if (flevel == PLevels-1) {
      INCR_OP(refetched);
    } else {
      INCR_OP(meta_data_refetched);
    }
  }
#endif // PRINT_STATS

  // Send fetch to all. 
#ifndef NO_STATE_TRANSLATION
  th_assert(flevel==PLevels-1 || next_chunk==0, "weird next_chunk value");
  Fetch f(rid, p.lu, flevel, p.index, next_chunk, p.c, replier);
#else
  Fetch f(rid, p.lu, flevel, p.index, p.c, replier);
#endif
  replica->send(&f, Node::All_replicas);
  //  fprintf(stderr, "Sending fetch message: rid=%qd lu=%qd (%d,%d) c=%qd rep=%d\n",rid, p.lu, flevel, p.index, p.c, replier);

  if (!cert->has_mine()) {
    Seqno ls = clog.head_seqno();
    if (!clog.fetch(ls).is_cleared() && p.c <= lc) {
      // Add my Meta_data_d message to the certificate
      Meta_data_d* mdd = new Meta_data_d(rid, flevel, p.index, ls);

      for (Seqno n=ls; n <= lc; n += checkpoint_interval) {
	if (clog.fetch(n).sd.is_zero()) continue; //XXXXfind a better way to do this
	Part& q = get_meta_data(n, flevel, p.index);
	mdd->add_digest(n, q.d);
      }

      cert->add(mdd, true);
    }
  }

  STOP_CC(fetch_cycles);
}


bool State::handle(Fetch *m, Seqno ls) {
  bool verified = true;

  if (m->verify()) {
    Principal* pi = replica->i_to_p(m->id());
    int l = m->level();
    int i = m->index();

    if (pi->last_fetch_rid() < m->request_id() && (l < PLevels-1 || i < nb)) {
      Seqno rc = m->checkpoint();

      //      fprintf(stderr, "Rx FTCH. ls=%qd rc=%qd lu=%qd lm=%qd. ",ls,rc,m->last_uptodate(),ptree[l][i].lm);

      if (rc >= 0 && m->replier() == replica->id()) {
	Seqno chosen = -1;
	if (clog.within_range(rc) && !clog.fetch(rc).is_cleared()) {
	  // Replica has the requested checkpoint
	  chosen = rc;
	} else if (lc >= rc && ptree[l][i].lm <= rc) {
	  // Replica's last checkpoint has same value as requested
	  // checkpoint for this partition
	  chosen = lc;
	}

	if (chosen >= 0) {
	  if (l == PLevels-1) {
	    // Send data
	    Part& p = get_meta_data(chosen, l, i);
#ifndef NO_STATE_TRANSLATION
	    int sz;
	    char *obj = get_data(chosen, i, sz);
	    if (sz <= m->chunk_number() * Fragment_size) {
	      // object may have shrunk. Resend from the beginning
	      //    fprintf(stderr, "Object shrunk. sending from beginning\n");
	      Data d(i, p.lm, obj, sz, 0);
	      replica->send(&d, m->id());
	    }
	    else {
	      Data d(i, p.lm, obj + m->chunk_number() * Fragment_size, sz,
		     m->chunk_number());
	      replica->send(&d, m->id());
	      //    fprintf(stderr, "Sending Data i=%d lm=%qd\n", i, p.lm);
	    }
#else
	    Data d(i, p.lm, get_data(chosen, i));
	    replica->send(&d, m->id());
#endif
	    //	    fprintf(stderr, "Sending data i=%d lm=%qd sz=%d chunk %d\n", i, p.lm, sz, m->chunk_number());
	  } else {
	    // Send meta-data
	    Part& p = get_meta_data(chosen, l, i);
	    Meta_data md(m->request_id(), l, i, chosen, p.lm, p.d);
	    Seqno thr = m->last_uptodate();

	    l++;
	    int j = i*PSize[l];
	    int max = j+PSize[l];
	    if (l == PLevels-1 && max > nb) max = nb;
	    for (; j < max; j++) {
	      Part& q = get_meta_data(chosen, l, j);
	      if (q.lm > thr)
		md.add_sub_part(j, q.d);
	    }
	    replica->send(&md, m->id());
	    //	    fprintf(stderr, "Sending meta-data l=%d i=%d lm=%qd\n", l-1, i, p.lm);
	  }
	  delete m;
	  return verified;
	}
      }

      if (ls > rc && ls >= m->last_uptodate() && ptree[l][i].lm > rc) {
	// Send meta-data-d
	Meta_data_d mdd(m->request_id(), l, i, ls);
	
	Seqno n = (clog.fetch(ls).is_cleared()) ? ls+checkpoint_interval : ls;
	for (; n <= lc; n += checkpoint_interval) {
	  Part& p = get_meta_data(n, l, i);
	  mdd.add_digest(n, p.d);
	}

	if (mdd.num_digests() > 0) {
	  mdd.authenticate(pi);
	  replica->send(&mdd, m->id());
	  //	  fprintf(stderr, "Sending meta-data-d l=%d i=%d \n", l, i);
	}
      }
    }
  } else {
    verified = false;
  }
  delete m;
  return verified;
}

void State::handle(Data *m) {
  INCR_OP(num_fetched);
  START_CC(fetch_cycles);

  int l = PLevels-1;
  if (fetching && flevel == l) {
    FPart& wp = stalep[l]->high();
    int i = wp.index;
    //    fprintf(stderr, "RxDat i=%d(sthi%d) chnkn %d,tot-sz %d,nchnks %d#total_sz %d,next_chnk %d\n", m->index(), i,m->chunk_number(), m->total_size(),m->num_chunks(),total_size, next_chunk);
    if (m->index() == i) {

#ifndef NO_STATE_TRANSLATION
      if (m->chunk_number() < m->num_chunks() &&
	  ((m->chunk_number() == next_chunk &&
	    (total_size == 0 || m->total_size() == total_size)) ||
	   (m->chunk_number() == 0 && total_size != 0))) {

	if (m->chunk_number() == 0 && total_size != 0) {
	  // Object shrunk. Restart fetch.
	  th_assert(reassemb != NULL, "Unexpected reassembly buffer");
	  delete [] reassemb;
	  reassemb = NULL;
	  next_chunk = 0;
	}

	if (m->chunk_number() < m->num_chunks()-1) { // Not final chunk - perform reassembly and send another fetch for the next fragment
	  if (next_chunk == 0) {  // First chunk - allocate buffer
	    th_assert(reassemb == NULL && m->total_size() > Fragment_size, "Unexpected reassembly buffer");
	    total_size = m->total_size();
	    reassemb = new char[total_size];
	    memcpy(reassemb, m->data(), Fragment_size);
	  }
	  else {  // Not first chunk - copy fragment to reassembly buffer
	    th_assert(reassemb != NULL && m->total_size() > (next_chunk+1)*Fragment_size, "Unexpected reassembly buffer");
	    memcpy(reassemb + next_chunk * Fragment_size, m->data(),
		   Fragment_size);
	  }
	  next_chunk++;
	  STOP_CC(fetch_cycles);
	  send_fetch();
	  delete m;
	  return;
	}
	else { // Final chunk
	  th_assert(m->chunk_number() == m->num_chunks()-1, "num chnks?");
	  if (m->chunk_number() == 0) { // only one message
	    th_assert(reassemb == NULL && m->total_size() <= Fragment_size, "Unexpected reassembly buffer");
	    reassemb = new char[m->total_size()];
	    memcpy(reassemb, m->data(), m->total_size());
	  }
	  else {  // > 1 msg - finish reassembly before computing digest
	    th_assert(reassemb != NULL, "Unexpected reassembly buffer");
	    memcpy(reassemb + next_chunk * Fragment_size, m->data(),
		   m->total_size() - next_chunk * Fragment_size);
	  }
	  Digest d;
	  digest(d, i, m->last_mod(), reassemb, m->total_size());
#else
      Digest d;
      digest(d, i, m->last_mod(), m->data(), Block_size);
#endif
      if (wp.c >= 0 && wp.d == d) {
	INCR_OP(num_fetched_a);
	//	fprintf(stderr, "DDDDDData i=%d, last chunk=%d, sz=%d\n", i,next_chunk,m->total_size());

	Part& p = ptree[l][i];
	DSum& psum = stree[l-1][i/PSize[l]];

	cowb.set(i);
	if (keep_ckpts && !cowb.test(i)) {
	  // Append a copy of p to the last checkpoint

	  BlockCopy* bcp;

#ifdef NO_STATE_TRANSLATION
	  bcp = new BlockCopy;
	  bcp->data = mem[i];
#else
	  if (i<nb-replica->used_state_pages()) {  // application state
	    char *data;
	    int size = get_segment(i, &data);
	    bcp = new BlockCopy(data, size);
	  }
	  else {                               // replication library state
	    bcp = new BlockCopy;
	    bcp->data = rep_mem[i-nb+replica->used_state_pages()];
	  }
#endif
	  bcp->lm = p.lm;
	  bcp->d = p.d;

	  clog.fetch(lc).append(l, i, bcp);
	}

	// Subtract old digest from parent sum
	psum.sub(p.d);

	p.d = wp.d;
	p.lm = m->last_mod();
#ifndef NO_STATE_TRANSLATION
	p.size = m->total_size();
#endif
	
	// Set data to the right value. Note that we set the
	// most current value of the data.

#ifdef NO_STATE_TRANSLATION
	mem[i] = m->data();
#else
	if (i<nb-replica->used_state_pages()) {  // application state
	  char *data_to_put = new char[m->total_size()];
	  memcpy(data_to_put, reassemb, m->total_size());
	  BlockCopy *bc = new BlockCopy(data_to_put, m->total_size());
	  delete [] reassemb;
	  reassemb = NULL;
	  fetched_pages->append_or_replace(PLevels-1,i,bc);
	  //	  fprintf(stderr,"Added %d\t",i);
	  //	  fetched_pages->print();
	  //	  put_pages(1, &i, &tmp);
	    }
	else                               // replication library state
	  rep_mem[i-nb+replica->used_state_pages()] = m->data();
#endif

	// Add new digest to parent sum
	psum.add(p.d);

	FPart& pwp = stalep[l-1]->high();
	th_assert(pwp.index == i/PSize[l], "Parent is not first at level l-1 queue");
	if (p.lm > pwp.lm) {
	  pwp.lm = p.lm;
	}
	
	cert->clear();
	stalep[l]->remove();

	if (stalep[l]->size() == 0) {
	  STOP_CC(fetch_cycles);
#ifndef NO_STATE_TRANSLATION
          next_chunk = 0;
	  total_size = 0;
	  reassemb = NULL;
#endif
	  done_with_level();
	  delete m;
	  return;
	}
      }
      //      else
      //	fprintf(stderr, "Wlong digest?\n"); d.print(); wp.d.print();
#ifndef NO_STATE_TRANSLATION
          next_chunk = 0;
	  total_size = 0;
	  reassemb = NULL;
	}
      }
      else {
	if (total_size && m->total_size() != total_size) {
	  total_size = 0;
	  next_chunk = 0;
	  if (reassemb)
	    delete [] reassemb;
	  reassemb = NULL;
	  send_fetch();
	  delete m;
	  return;
	}
      }
#endif
      STOP_CC(fetch_cycles);
      send_fetch();
    } 
  }
  delete m;
  STOP_CC(fetch_cycles);
}


bool State::check_digest(Digest& d, Meta_data* m) {
  th_assert(m->level() < PLevels-1, "Invalid argument");

  int l = m->level();
  int i = m->index();
  DSum sum = stree[l][i];
  Meta_data::Sub_parts_iter miter(m);
  Digest dp;
  int ip;
  while (miter.get(ip, dp)) {
    if (ip >= nb && l+1 == PLevels-1) 
      break;

    if (!dp.is_zero()) {
      sum.sub(ptree[l+1][ip].d);
      sum.add(dp);
    }
  }
  digest(dp, i, m->last_mod(), (char*)(sum.sum), DSum::nbytes);
  return d == dp;
}


void State::handle(Meta_data* m) {
  INCR_OP(meta_data_fetched);
  INCR_CNT(meta_data_bytes, m->size()); 
  START_CC(fetch_cycles);

  Request_id crid = replica->principal()->last_fetch_rid();
  //  fprintf(stderr, "Got meta_data index %d from %d rid=%qd crid=%qd\n", m->index(), m->id(), m->request_id(), crid);
  if (fetching && flevel < PLevels-1 && m->request_id() == crid && flevel == m->level()) {
    FPart& wp = stalep[flevel]->high();
 
    if (wp.index == m->index() && wp.c >= 0 && m->digest() == wp.d) {
      // Requested a specific digest that matches the one in m
      if (m->verify() && check_digest(wp.d, m)) {
	INCR_OP(meta_data_fetched_a);

	// Meta-data was fetched successfully.
	//printf("Accepted meta_data from %d (%d,%d) \n", m->id(), flevel, wp.index);

	wp.lm = m->last_mod();

	// Queue out-of-date subpartitions for fetching, and if
	// checking, queue up-to-date partitions for checking.
	flevel++;
	th_assert(stalep[flevel]->size() == 0, "Invalid state");
     
	Meta_data::Sub_parts_iter iter(m);
	Digest d;
	int index;
	
	while (iter.get(index, d)) {
	  if (flevel == PLevels-1 && index >= nb)
	    break;

	  Part& p = ptree[flevel][index];

	  if (d.is_zero() || p.d == d) {
	    // Sub-partition is up-to-date
	    if (refetch_level == PLevels && p.lm <= check_start) {
	      to_check->_enlarge_by(1);
	      CPart &cp = to_check->high();
	      cp.level = flevel;
	      cp.index = index;
	    }
	  } else {
	    // Sub-partition is out-of-date
	    stalep[flevel]->_enlarge_by(1);
	    FPart& fp = stalep[flevel]->high();
	    fp.index = index;
	    fp.lu = wp.lu;
	    fp.lm = p.lm;
	    fp.c = wp.c;
	    fp.d = d;
	  }
	}

	cert->clear();
	STOP_CC(fetch_cycles);      

	if (stalep[flevel]->size() == 0) {
	  done_with_level();
	} else {
	  send_fetch();
	}
      }
    }
  }
  
  delete m;
  STOP_CC(fetch_cycles);
}


void State::handle(Meta_data_d* m) {
  INCR_OP(meta_datad_fetched);
  INCR_CNT(meta_datad_bytes, m->size());
  START_CC(fetch_cycles);

  //  printf("Got meta_data_d from %d index %d\n", m->id(),m->index());
  Request_id crid = replica->principal()->last_fetch_rid();
  if (fetching && m->request_id() == crid && flevel == m->level()) {
    FPart& wp = stalep[flevel]->high();
    
    if (wp.index == m->index() && m->last_stable() >= lc && m->last_stable() >= wp.lu) {
      INCR_OP(meta_datad_fetched_a);

      // Insert message in certificate for working partition
      Digest cd;
      Seqno cc;
      if (cert->add(m)) {
	if (cert->last_stable() > lc)
	  keep_ckpts = false;

	if (cert->cvalue(cc, cd)) {
	  // Certificate is now complete. 
	  wp.c = cc;
	  wp.d = cd;
	  //printf("Complete meta_data_d cert (%d,%d) \n", flevel, wp.index);

	  cert->clear(); 

	  th_assert(flevel != PLevels-1 || wp.index < nb, "Invalid state");
	  if (cd == ptree[flevel][wp.index].d) {
	    // State is up-to-date
	    if (refetch_level == PLevels && ptree[flevel][wp.index].lm <= check_start) {
	      to_check->_enlarge_by(1);
	      CPart &cp = to_check->high();
	      cp.level = flevel;
	      cp.index = wp.index;
	    }

	    STOP_CC(fetch_cycles);
	    
	    if (flevel > 0) {
	      stalep[flevel]->remove();
	      if (stalep[flevel]->size() == 0) {
		done_with_level();
	      }
	    } else {
	      flevel++;
              done_with_level();
	    }
	    return;
	  }
	  STOP_CC(fetch_cycles);
	  send_fetch(true);
	}
      }

      STOP_CC(fetch_cycles);
      return;
    }
  }

  STOP_CC(fetch_cycles);      
  delete m;
}


void State::done_with_level() {
  START_CC(fetch_cycles);

  th_assert(stalep[flevel]->size() == 0, "Invalid state");
  th_assert(flevel > 0, "Invalid state");
  
  flevel--;
  FPart& wp = stalep[flevel]->high();
  int i = wp.index;
  int l = flevel;

  wp.lu = wp.c;
  th_assert(wp.c != -1, "Invalid state");

  if (wp.lu >= wp.lm) {
    // partition is consistent: update ptree and stree, and remove it
    // from stalep
    Part& p = ptree[l][i];

    if (keep_ckpts) {
      // Append a copy of p to the last checkpoint
      Part* np = new Part;
      np->lm = p.lm;
      np->d = p.d;
      clog.fetch(lc).appendr(l, i, np);
    }

    if (l > 0) {
      // Subtract old digest from parent sum
      stree[l-1][i/PSize[l]].sub(p.d);
    }

    p.lm = wp.lm;
    p.d = wp.d;

    if (l > 0) {
      // add new digest to parent sum
      stree[l-1][i/PSize[l]].add(p.d);

      FPart& pwp = stalep[l-1]->high();
      th_assert(pwp.index == i/PSize[l], "Parent is not first at level l-1 queue");
      if (p.lm > pwp.lm) {
	pwp.lm = p.lm;
      }
    }

    if (l == 0) {
      // Completed fetch
      fetching = false;
      if (checking && to_check->size() == 0)
	checking = false;

#ifndef NO_STATE_TRANSLATION

      // update app's state:

      Page_mapping::Iter iter(fetched_pages);
      int lev,ind;
      BlockCopy *b;
      int total_objs=fetched_pages->num_entries();
      int *indices=new int[total_objs];
      char **pages=new char*[total_objs];
      int *sizes = new int[total_objs];
      int index=0;
      while (iter.get(lev,ind,b)) {
	indices[index] = ind;
	sizes[index] = b->data.size;
	pages[index++] = b->data.data;
	//	fprintf(stderr, "Put for index %d sz %d\n", ind, b->data.size);
      }
      th_assert(index==total_objs,"index =/= total_pages");
      put_segments(index, sizes, indices, pages);
      // fetched_pages->print();

      fetched_pages->clear();
      delete [] indices;
      delete [] sizes;
      delete [] pages;

#endif

      if (keep_ckpts && lc%checkpoint_interval != 0) {
	// Move parts from this checkpoint to previous one
	Seqno prev = lc/checkpoint_interval * checkpoint_interval;
	if (clog.within_range(prev) && !clog.fetch(prev).is_cleared()) {
	  Checkpoint_rec& pr = clog.fetch(prev);
	  Checkpoint_rec& cr = clog.fetch(lc);
	  Checkpoint_rec::Iter g(&cr);
	  int pl, pi;
	  Part* p;
	  while (g.get(pl, pi, p)) {
	    pr.appendr(pl, pi, p);
	  }
	}
      }

      // Create checkpoint record for current state
      th_assert(lc <= wp.lu, "Invalid state");
      lc = wp.lu;

      if (!clog.within_range(lc)) {
	//	fprintf(stderr, "done w/level: Truncating clog to %qd\n", lc-max_out);
	clog.truncate(lc-max_out);
      }

      Checkpoint_rec &nr = clog.fetch(lc);
      nr.sd = ptree[0][0].d;
      cowb.clear();
      stalep[l]->remove();   
      cert->clear();

      replica->new_state(lc);
      
      // If checking state, stop adding partitions to to_check because
      // all partitions that had to be checked have already been
      // queued.
      refetch_level = 0;
      poll_cnt = 16;
      
      //      printf("Ending fetch state %qd \n", lc);

      STOP_CC(fetch_cycles);
      return;
    } else {
      stalep[l]->remove();
      if (stalep[l]->size() == 0) {
	STOP_CC(fetch_cycles);
	done_with_level();
	return;
      }

      // Moving to a sibling.
      if (flevel <= refetch_level)
	refetch_level = PLevels;
    } 
  } else {
    // Partition is inconsistent
    if (flevel < refetch_level)
      refetch_level = flevel;
  }

  STOP_CC(fetch_cycles);

  send_fetch();
}


//
// State checking:
//
void State::start_check(Seqno le) {
  checking = true;
  refetch_level = PLevels;
  lchecked = -1;
  check_start = lc;
  corrupt = false;
  poll_cnt = 16;

  start_fetch(le);
}


inline bool State::check_data(int i) {
  th_assert(i < nb, "Invalid state");

  //  fprintf(stderr, "Checking data i=%d \t", i);

  Part& p = ptree[PLevels-1][i];
  Digest d;
  digest(d, PLevels-1, i);

  return d == p.d;
}


void State::check_state() {
  START_CC(check_time);

//  if (refetch_level != 0 && replica->has_messages(0))
//    return;

  int count = 1;
  while (to_check->size() > 0) {    
    CPart& cp = to_check->slot(0);
    int min = cp.index*PBlocks[cp.level];
    int max = min+PBlocks[cp.level];
    if (max > nb) max = nb;
      
    if (lchecked < min || lchecked >= max)
      lchecked = min;

    while (lchecked < max) {
      if (count%poll_cnt == 0 && replica->has_messages(0)) {
	STOP_CC(check_time);
	return;
      }

      Part& p = ptree[PLevels-1][lchecked];

      if (p.lm > check_start || check_data(lchecked)) {
	// Block was fetched after check started or has correct digest.
	INCR_OP(num_checked);
	count++;
	lchecked++;
      } else {
	corrupt = true;
	//XXXXXX put these blocks in a queue and fetch them at the end.
	th_fail("Replica's state is corrupt. Should not happen yet");
      }
    }

    to_check->slot(0) = to_check->high();
    to_check->remove();
  }
  STOP_CC(check_time);

  if (!fetching) {
    //    printf("Finished checking and fetching\n");
    checking = false;
    refetch_level = 0;
    replica->try_end_recovery();
  }
}


void State::mark_stale() {
  cert->clear();
}


bool State::shutdown(FILE* o, Seqno ls) {
  bool ret = cowb.encode(o);

  size_t wb = 0;
  size_t ab = 0;
  for (int i=0; i < PLevels; i++) {
    int psize = (i != PLevels-1) ? PLevelSize[i] : nb;
    wb += fwrite(ptree[i], sizeof(Part), psize, o);
    ab += psize;
  }
  
  wb += fwrite(&lc, sizeof(Seqno), 1, o);
  ab++;

  //XXXXXX what if I shutdown while I am fetching.
  //recovery should always start s fetch for digests.
  if (!fetching || keep_ckpts) {
    for (Seqno i=ls; i <= ls+max_out; i++) {
      Checkpoint_rec& rec = clog.fetch(i);
      
      if (!rec.is_cleared()) {
	wb += fwrite(&i, sizeof(Seqno), 1, o);
	wb += fwrite(&rec.sd, sizeof(Digest), 1, o);

	int size = rec.num_entries();
	wb += fwrite(&size, sizeof(int), 1, o);
	ab += 3;
	
	Checkpoint_rec::Iter g(&rec);
	int l,i;
	Part *p;
	
	while (g.get(l, i, p)) {
	  wb += fwrite(&l, sizeof(int), 1, o);
	  wb += fwrite(&i, sizeof(int), 1, o);

	  //	  fprintf(stderr, "Entry l=%d - ", l);
	  
	  int psize = (l == PLevels-1) ? sizeof(BlockCopy) : sizeof(Part);
	  wb += fwrite(p, psize, 1, o);
	  ab += 3;
#ifndef NO_STATE_TRANSLATION
	  if (l == PLevels-1) {
	    //	    fprintf(stderr, "Wrt i=%d, sz=%d \t",i, ((BlockCopy *)p)->data.size);
	    wb += fwrite(((BlockCopy *)p)->data.data, ((BlockCopy *)p)->data.size, 1, o);
	    ab++;
	  }
#endif
	}
      }
    }
  }
  Seqno end=-1;

#ifndef NO_STATE_TRANSLATION
  wb += fwrite(&end, sizeof(Seqno), 1, o);
  ab++;

  wb += fwrite(rep_mem->data, Block_size, replica->used_state_pages(), o);
  ab += replica->used_state_pages();

  shutdown_proc(o);
#else
  wb += fwrite(&end, sizeof(Seqno), 1, o);
  ab++;

  msync(mem, nb*Block_size, MS_SYNC);
#endif

  return ret & (ab == wb);
}


bool State::restart(FILE* in, Replica *rep, Seqno ls, Seqno le, bool corrupt) {
  replica = rep;

#ifndef NO_STATE_TRANSLATION
  pages_lc->clear();
#endif

  if (corrupt) {
    clog.clear(ls);
    lc = -1;
    return false;
  }

  bool ret = cowb.decode(in);
#ifdef NO_STATE_TRANSLATION
  _Byz_cow_bits = cowb.bitvec();
#endif

  size_t rb = 0;
  size_t ab = 0;
  for (int i=0; i < PLevels; i++) {
    int psize = (i != PLevels-1) ? PLevelSize[i] : nb;
    rb += fread(ptree[i], sizeof(Part), psize, in);
    ab += psize;
  }

  // Compute digests of non-leaf partitions assuming leaf digests are
  // correct. This should have negligible cost when compared to the
  // cost of rebooting. This is required for the checking code to ensure
  // correctness of the state.
  int np = nb;
  for (int l=PLevels-1; l > 0; l--) {
    for (int i = 0; i < np; i++) {
      if (i%PSize[l] == 0) 
	stree[l-1][i/PSize[l]] = DSum(); 

      if (l < PLevels-1) {
	Digest d;
	digest(d, l, i);
	if (d != ptree[l][i].d) {
	  ret = false;
	  ptree[l][i].d = d;
	}
      }
	
      stree[l-1][i/PSize[l]].add(ptree[l][i].d);
    }
    np = (np+PSize[l]-1)/PSize[l];    
  }

  Digest d; 
  digest(d, 0, 0);
  if (d != ptree[0][0].d)
    ret = false;

  clog.clear(ls);
  rb += fread(&lc, sizeof(Seqno), 1, in);
  if (lc < ls || lc > le)
    return false;
 
  Seqno n;
  rb += fread(&n, sizeof(Seqno), 1, in);
  ab += 2;

  while (n >= 0) {
    if (n < ls || n > lc)
      return false;

    Checkpoint_rec& rec = clog.fetch(n);

    rb += fread(&rec.sd, sizeof(Digest), 1, in);
    ab++;


    int size;
    rb += fread(&size, sizeof(int), 1, in);
    ab++;
    if (size > 2*nb)
      return false;
    
    for(int k=0; k < size; k++) {
      int l,i;

      rb += fread(&l, sizeof(int), 1, in);
      rb += fread(&i, sizeof(int), 1, in);

      Part* p;
      int psize;
      if (l == PLevels-1) {
	p = new BlockCopy;
	psize = sizeof(BlockCopy);
      } else {
	p = new Part;
	psize = sizeof(Part);
      }

      rb += fread(p, psize, 1, in);
      ab += 3;

      //      fprintf(stderr, "Entry l=%d - ", l);
#ifndef NO_STATE_TRANSLATION
      if (l == PLevels-1) {
	//	fprintf(stderr, "Rd i=%d, sz=%d \t",i, ((BlockCopy *)p)->data.size);
	((BlockCopy *)p)->data.data = NULL;
	char *data = new char[((BlockCopy *)p)->data.size];
	rb += fread(data, ((BlockCopy *)p)->data.size, 1, in);
	ab++;
	((BlockCopy *)p)->data.init_from_ptr(data, ((BlockCopy *)p)->data.size);
      }
#endif

      rec.appendr(l, i, p);
    }

    rb += fread(&n, sizeof(Seqno), 1, in);
    ab++;
  }

#ifndef NO_STATE_TRANSLATION
  rb += fread(rep_mem->data, Block_size, replica->used_state_pages(), in);
  ab += replica->used_state_pages();
  restart_proc(in);
#endif
  return ret & (ab == rb);
}


bool State::enforce_bound(Seqno b, Seqno ks, bool corrupt) {
  bool ret = true;
  for (int i=0; i < PLevels; i++) {
    int psize = (i != PLevels-1) ? PLevelSize[i] : nb;
    for (int j=0; j < psize; j++) {
      if (ptree[i][j].lm >= b) {
	ret = false;
	ptree[i][j].lm = -1;
      }
    }
  }

  if (!ret || corrupt || clog.head_seqno() >= b) {
    lc = -1;
    clog.clear(ks);
    return false;
  }

  return true;
}


void State::simulate_reboot() {
  START_CC(reboot_time);

  static const unsigned long reboot_usec = 30000000;
  //  static const unsigned long reboot_usec = 3000000;
  Cycle_counter inv_time;
  
  // Invalidate state pages to force reading from disk after reboot
  inv_time.start();
#ifdef NO_STATE_TRANSLATION
  msync(mem, nb*Block_size, MS_INVALIDATE);
#endif
  inv_time.stop();

  usleep(reboot_usec - inv_time.elapsed()/clock_mhz);
  STOP_CC(reboot_time);
}

