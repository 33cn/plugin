#include <string.h>
#include "th_assert.h"
#include "Message_tags.h"
#include "Data.h"

#ifndef NO_STATE_TRANSLATION
Data::Data(int i, Seqno lm, char *data, int totalsz, int chunkn)
#else
Data::Data(int i, Seqno lm, char *data)
#endif
: Message(Data_tag, sizeof(Data_rep)) {
  rep().index = i;
  rep().padding = 0;
  rep().lm = lm;
#ifndef NO_STATE_TRANSLATION
  rep().total_size = totalsz;
  rep().chunk_no = chunkn;
  int data_size = (totalsz / Fragment_size == chunkn ?
		   totalsz % Fragment_size : Fragment_size);
  memcpy(rep().data, data, data_size);
  set_size(sizeof(Data_rep) - Fragment_size + data_size);
  //  fprintf(stderr, "Size of msg is now %d\n", rep().size);
#else
  // TODO: Avoid this copy using sendmsg with iovecs.
  memcpy(rep().data, data, Block_size);                 
#endif
} 


bool Data::convert(Message *m1, Data  *&m2) {
#ifndef NO_STATE_TRANSLATION
  if (!m1->has_tag(Data_tag, sizeof(Data_rep) - Fragment_size))
#else
  if (!m1->has_tag(Data_tag, sizeof(Data_rep)))
#endif
    return false;

  m2 = (Data*)m1;
  m2->trim();
  return true;
}
