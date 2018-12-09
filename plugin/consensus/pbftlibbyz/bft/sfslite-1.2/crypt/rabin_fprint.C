/* XXX - Copyright on this code is GPL encumbered. */

#include <sys/types.h>
#include <stdio.h>
#include <stdlib.h>

#include "rabin_fprint.h"
#include "rabinpoly.h"
  
unsigned rabin_fprint::min_size_suppress = 0;
unsigned rabin_fprint::max_size_suppress = 0;

u_int64_t 
fingerprint(const unsigned char *data, size_t count)
{
  u_int64_t poly = FINGERPRINT_PT;
  window w (poly);
  w.reset();
  u_int64_t fp = 0;
  for (size_t i = 0; i < count; i++)
    fp = w.append8 (fp, data[i]);
  return fp;
}

rabin_fprint::rabin_fprint()
  : _w(FINGERPRINT_PT)
{
  _last_pos = 0;
  _cur_pos = 0;
  _w.reset();
  _num_chunks = 0;
}

rabin_fprint::~rabin_fprint()
{
}

void
rabin_fprint::stop()
{
}

ptr<vec<unsigned int> >
rabin_fprint::chunk_data (suio *in_data)
{
  unsigned char *buf = New unsigned char[in_data->resid()];
  in_data->copyout(buf, in_data->resid());
  return chunk_data(buf, in_data->resid());
  delete[] buf;
}

ptr<vec<unsigned int> >
rabin_fprint::chunk_data(const unsigned char *data, size_t size)
{
  ptr<vec<unsigned int> > iv = NULL;
  u_int64_t f_break = 0;
  size_t start_i = 0;
  for (size_t i=0; i<size; i++, _cur_pos++) {
    f_break = _w.slide8 (data[i]);
    size_t cs = _cur_pos - _last_pos;
    if ((f_break % chunk_size) == BREAKMARK_VALUE && cs < MIN_CHUNK_SIZE) 
      min_size_suppress++;
    else if (cs == MAX_CHUNK_SIZE)
      max_size_suppress++;
    if (((f_break % chunk_size) == BREAKMARK_VALUE && cs >= MIN_CHUNK_SIZE) 
        || cs >= MAX_CHUNK_SIZE) {
      if (!iv) {
        iv = new refcounted<vec<unsigned int> >;
      }
      _w.reset();
      iv->push_back(cs);
      _last_pos = _cur_pos;
      start_i = i;
    }
  }
  return iv;
}
