#ifndef _RABIN_FPRINT_H_
#define _RABIN_FPRINT_H_

// we keep P(t), x, and K the same for whole file system, so two equivalent
// files would have the same breakmarks. for string A, fingerprint of A is
//
//   f(A) = A(t) mod P(t)
//
// we create breakmarks when
//
//   f(A) mod K = x
//
// if we use K = 8192, the average chunk size is 8k. we allow multiple K
// values so we can do multi-level chunking.

#include "async.h"
#include "fprint.h"
#include "rabinpoly.h"

#define FINGERPRINT_PT  0xbfe6b8a5bf378d83LL
#define BREAKMARK_VALUE 0x78
#define MIN_CHUNK_SIZE  2048
#define MAX_CHUNK_SIZE  65535

u_int64_t fingerprint(const unsigned char *data, size_t count);

class rabin_fprint : public fprint {
private:
  window _w;
  size_t _last_pos;
  size_t _cur_pos;
  
  unsigned int _num_chunks;

public:
  rabin_fprint();
  ~rabin_fprint();

  void stop();
  ptr<vec<unsigned int> > chunk_data (const unsigned char *data, size_t size);
  ptr<vec<unsigned int> > chunk_data (suio *in_data);

  static const unsigned chunk_size = 32768;
  static unsigned min_size_suppress;
  static unsigned max_size_suppress;
};

#endif // _RABIN_FPRINT_H_
