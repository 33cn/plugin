#ifndef _FPRINT_H_
#define _FPRINT_H_

#include "async.h"

class fprint {
public:
  virtual ~fprint() {};

  virtual void stop() = 0;

  // These functions return a vector of ints (possibly NULL) that
  // specify where the next block boundaries are as offset from the
  // previous boundary
  virtual ptr<vec<unsigned int> > chunk_data (const unsigned char *data,
					      size_t size) = 0;
  virtual ptr<vec<unsigned int> > chunk_data (suio *in_data) = 0;
};

#endif // _FPRINT_H_
