#ifndef _Bitmap_h
#define _Bitmap_h 1

#include <stdio.h>
#include "bits.h"
#include "th_assert.h"

typedef unsigned long Chunk;
const unsigned long ChunkBits = sizeof(Chunk)*byte_bits;

class Bitmap {

public:
  Bitmap(Uint size, bool value=false);
  // Creates a bitmap with all booleans to "value" 

  Bitmap(Bitmap const & other);

  ~Bitmap();

  Bitmap& operator=(Bitmap const &other);

  Chunk *bitvec() { return chunks; }

  void clear();
  // Clears the bitmap. 

  void setAll();
  // Sets all the bits to 1.
  
  bool test(unsigned int i) const;
  // requires: "i" is within bounds.      
  // effects:  returns ith boolean.      

  int size() const;
  // effects:  returns the size of the bitmap 

  void assign(unsigned int i, bool value);
  // requires: "i" is within bounds. value is true or false      
  // effects:  Sets ith boolean to val.  

  void set(unsigned int i);
  // requires: "i" is within bounds.      
  // effects:  Sets ith boolean to true.  
  // Note: this is faster than "assign"ing it to true.
  
  void reset(unsigned int i);
  // requires: "i" is within bounds.      
  // effects:  Sets ith boolean to false. 
  
  void setRange(Uint min, Uint max);
  // requires: min is within bounds, max is within bounds+1.
  // effects: sets all bits in [min, max[.  
  
  void resetRange(Uint min, Uint max);
  // requires: min is within bounds, max is within bounds+1.
  // effects: resets all bits in [min, max[.  

  bool all_zero() const;
  // efffects: returns true iff the bitmap contains only zeros.

  bool operator== (Bitmap const other) const;
  // Returns true if this and other are equal, false if not.

  void print ();
  // effects: Prints a human readable of "this" on stdout 


  int total_set();
  // Returns: total number of bits that are set.

  class Iter {
    // An iterator for yielding the booleans in a bool bitmap
    // Once created, an iterator must be used before any
    // changes are made to the iterated object. The effect is undefined
    // if an iterator method is called after such a change.

  public:
    Iter(Bitmap *bitmap, bool val=true);
    // Creates iterator to yield indices and values of the bitmap in order.
    // Only those booleans set to value are yielded.

    bool get(unsigned int& index);
    // modifies: index
    // effects: Sets "index" to the next index that has the value
    // specified in the constructor. Returns false iff there is no
    // such index.
       
  private:
    Bitmap *bitmap;        // The bitmap being yielded. 
    bool  value;           // The value being searched. 
    unsigned int index;    // Index of next boolean to be tested 
    Chunk ignoreValue;    // Chunk with all bits set to the value not
                           // being searched.
  };

  bool encode(FILE* o);
  bool decode(FILE* i);
  // Effects: Encodes and decodes object state from stream. Return
  // true if successful and false otherwise.

private:
  friend class Iter;

  Uint num;      // size of the bitmap
  Chunk *chunks; // array of chunks storing booleans
  Uint nc;       // number of chunks


  Chunk bitSelector(unsigned int i) const;
  // requires: "ChunkBits" is a power of 2.
  // effects: Computes the position p of the ith boolean within its
  // chunk and returns a chunk with pth bit set
  

  Chunk valBitSelector(unsigned int i, bool val) const;
  // requires: "ChunkBits" is a power of 2. val is true or false
  // effects:  Computes the position of the ith boolean within its
  // chunk, say p, and returns a chunk with ith bit set to val
  //   and other bits to 0
};

inline Bitmap::Bitmap(Uint sz, bool value) {
  num = sz;
  nc = (sz + ChunkBits - 1)/ChunkBits;
  chunks = new Chunk[nc];
  for (Uint i = 0; i < nc; i++) {
    chunks[i] = value? (~0UL) : 0UL;
  }
}

inline Bitmap::Bitmap(Bitmap const & other) {
  num = other.num;
  chunks = new Chunk[nc];
  for (Uint i = 0; i < nc; i++) {
    chunks[i] = other.chunks[i];
  }
}

inline Bitmap& Bitmap::operator=(Bitmap const &other) {
  if (this == &other) return *this;
  if (nc != other.nc) {
    delete [] chunks;
    chunks = new Chunk[other.nc];
    nc = other.nc;
  }
  num = other.num;

  for (Uint i = 0; i < nc; i++) {
    chunks[i] = other.chunks[i];
  }
  return *this;
}

inline Bitmap::~Bitmap() { delete [] chunks; }

inline Chunk Bitmap::bitSelector(unsigned int i) const{
  return 1UL << (i%ChunkBits);
}


inline Chunk Bitmap::valBitSelector(unsigned int i, bool val) const{
  return ((unsigned long)val) << (i%ChunkBits);
}

inline void Bitmap::clear() {
  for (Uint i = 0; i < nc; i++) chunks[i] = 0UL;
}
    
inline void Bitmap::setAll() {
  for (Uint i = 0; i < nc; i++) chunks[i] = ~(0UL);
}


inline bool Bitmap::test(unsigned int i) const {
    th_assert(i < num, "Index out of bounds\n");
    return (chunks[i/ChunkBits] & bitSelector(i)) != 0;
}


inline int Bitmap::size() const {
    return num;
}

inline void Bitmap::set(unsigned int i) {
    th_assert(i < num, "Index out of bounds\n");
    chunks[i/ChunkBits] |= bitSelector(i);
}


inline void Bitmap::assign(unsigned int i, bool val) {
    th_assert(i < num, "Index out of bounds\n");
    chunks[i/ChunkBits] &= (~ bitSelector(i));
    chunks[i/ChunkBits] |= valBitSelector(i, val);
}


inline void Bitmap::reset(unsigned int i) {
    th_assert(i < num, "Index out of bounds\n");
    chunks[i/ChunkBits] &= (~ bitSelector(i));
}

inline void Bitmap::setRange(Uint min, Uint max) {
  th_assert(min < num && max < num, "Index out of bounds\n");
  for (Uint i = min; i < max; i++) set(i);  
}

inline void Bitmap::resetRange(Uint min, Uint max) {
  th_assert(min < num && max < num, "Index out of bounds\n");
  for (Uint i = min; i < max; i++) reset(i);  
}

inline void  Bitmap::print() {
  for (Uint i = 0; i < nc; i++) {
    printf("%lx", chunks[i]);
  }
}


inline int Bitmap::total_set() {
  int retval=0;
  for (int i=0;i<size();i++)
    if (test(i))
      retval++;
  return retval;
}

inline bool Bitmap::operator== (Bitmap const other) const {
  if (num != other.num) return (false);
  for (Uint i = 0; i < nc; i++) 
    if (chunks[i] != other.chunks[i]) return (false);
  return (true);
}


inline bool Bitmap::all_zero() const {
  for (Uint i = 0; i < nc; i++) 
    if (chunks[i] != 0) return false;
  return true;
}


inline Bitmap::Iter::Iter(Bitmap *b, bool v):
    bitmap(b), value(v), index(0), ignoreValue((v) ? 0 : ~0){}


inline bool Bitmap::Iter::get(Uint& ind) {
  while (index < bitmap->num) {
    if (bitmap->chunks[index/ChunkBits] == ignoreValue) {
      index += ChunkBits;
      continue;
    } 
    
    bool this_value = bitmap->test(index);
    if (this_value == value) {
      ind = index++;
      return true;
    }
    index++;
  }
  return false;
}


inline bool Bitmap::encode(FILE* o) {
  size_t sz = fwrite(&num, sizeof(Uint), 1, o);
  sz += fwrite(&nc, sizeof(Uint), 1, o);
  sz += fwrite(chunks, sizeof(Uint), nc, o);

  return sz == 2+nc;
}


inline bool Bitmap::decode(FILE* i) {
  size_t sz = fread(&num, sizeof(Uint), 1, i);
  sz += fread(&nc, sizeof(Uint), 1, i);

  delete [] chunks;
  chunks = new Chunk[nc];

  sz += fread(chunks, sizeof(Uint), nc, i);

  return ((sz == 2+nc) && (num <= nc*ChunkBits));
}


//
// Generic char[] bitmap manipulation
//

static inline void Bits_set(char* bmap, int i) {
  char *byte = bmap+(i/byte_bits);
  *byte |=  (1 << (i%byte_bits));
}

static inline void Bits_reset(char* bmap, int i) {
  char *byte = bmap+(i/byte_bits);
  *byte &=  ~(1 << (i%byte_bits));
}

static inline bool Bits_test(char* bmap, int i) {
  char *byte = bmap+(i/byte_bits);
  return (*byte & (1 << (i%byte_bits))) ? true : false;
}

#endif // Bitmap.h
