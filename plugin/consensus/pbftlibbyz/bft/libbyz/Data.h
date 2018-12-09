#ifndef _Data_h
#define _Data_h 1

#include "types.h"
#include "Message.h"
#include "Partition.h"
#include "State_defs.h"

#ifndef NO_STATE_TRANSLATION
static const int Fragment_size = 4096;
#endif


// 
// Data messages have the following format:
//
struct Data_rep : public Message_rep {
  int index;        // index of this page within level
  int padding;
#ifndef NO_STATE_TRANSLATION
  int total_size;   // size of the object
  int chunk_no;     // current fragment number
#endif
  Seqno lm;         // Seqno of last checkpoint in which data was modified
#ifndef NO_STATE_TRANSLATION
  char data[Fragment_size];
#else
  char data[Block_size];
#endif
};

class Data : public Message {
  // 
  // Data messages
  //
public:
#ifndef NO_STATE_TRANSLATION
  Data(int i, Seqno lm, char *data, int totalsz, int chunkn);
#else
  Data(int i, Seqno lm, char *data);
#endif
  // Effects: Creates a new Data message. 
  //          (if we are using BASE) "totalsz" is the size of the object,
  //          chunkn is the number of the fragment that we are sending
  //          and "data" points to the beginning of the fragment (not to
  //          the beginning of the object - XXX change this?)

  int index() const;
  // Effects: Returns index of data page

  Seqno last_mod() const;
  // Effects: Returns the seqno of last checkpoint in which data was
  // modified

#ifndef NO_STATE_TRANSLATION
  int total_size() const;
  // Effects: Returns total size of data (this msg may contain only a
  //          fragment of smaller size)

  int chunk_number() const;
  // Effects: Returns the number of the fragment that these data correspond
  //          to, in case it is a fragmented message.

  int num_chunks() const;
  // Effects: Returns the total number of fragments needed to send the Data
  //          block

#endif

  char *data() const;
  // Effects: Returns a pointer to the data page.

  static bool convert(Message *m1, Data *&m2);
  // Effects: If "m1" has the right size and tag, casts "m1" to a
  // "Data" pointer, returns the pointer in "m2" and returns
  // true. Otherwise, it returns false. 

private:
  Data_rep &rep() const;
  // Effects: Casts contents to a Data_rep&
};

inline Data_rep &Data::rep() const { 
  th_assert(ALIGNED(msg), "Improperly aligned pointer");
  return *((Data_rep*)msg); 
}

inline int Data::index() const { return rep().index; }

inline Seqno Data::last_mod() const { return rep().lm; }

#ifndef NO_STATE_TRANSLATION
inline int Data::total_size() const { return rep().total_size; }

inline int Data::chunk_number() const { return rep().chunk_no; }

inline int Data::num_chunks() const { return rep().total_size / Fragment_size +
				 (rep().total_size % Fragment_size ? 1 : 0); }
#endif

inline char *Data::data() const { return rep().data; }


#endif // _Data_h
