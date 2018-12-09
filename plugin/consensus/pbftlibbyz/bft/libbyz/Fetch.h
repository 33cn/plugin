#ifndef _Fetch_h
#define _Fetch_h 1

#include "types.h"
#include "Digest.h"
#include "Message.h"
#include "State_defs.h"

class Principal;

// 
// Fetch messages have the following format:
//
struct Fetch_rep : public Message_rep {
  Request_id rid;     // sequence number to prevent replays
  int level;          // level of partition
  int index;          // index of partition within level
  Seqno lu;           // information for partition is up-to-date till seqno lu
  Seqno rc;           // specific checkpoint requested (-1) if none
  int repid;          // id of designated replier (valid if c >= 0)
  int id;             // id of the replica that generated the message.
#ifndef NO_STATE_TRANSLATION
  int chunk_no;       // number of the fragment we are requesting
  int padding;
#endif

  // Followed by an authenticator.
};

class Fetch : public Message {
  // 
  // Fetch messages
  //
public:
  Fetch(Request_id rid, Seqno lu, int level, int index,
#ifndef NO_STATE_TRANSLATION
	int chunkn = 0,
#endif
	Seqno rc=-1, int repid=-1);
  // Effects: Creates a new authenticated Fetch message.

  void re_authenticate(Principal *p=0);
  // Effects: Recomputes the authenticator in the message using the
  // most recent keys. If "p" is not null, may only update "p"'s
  // entry.

  Request_id request_id() const;
  // Effects: Fetches the request identifier from the message.

  Seqno last_uptodate() const;
  // Effects: Fetches the last up-to-date sequence number from the message.

  int level() const;
  // Effects: Returns the level of the partition  

  int index() const;
  // Effects: Returns the index of the partition within its level

  int id() const;
  // Effects: Fetches the identifier of the replica from the message.

#ifndef NO_STATE_TRANSLATION
  int chunk_number() const;
  // Effects: Returns the number of the fragment that is being requested
#endif


  Seqno checkpoint() const;
  // Effects: Returns the specific checkpoint requested or -1

  int replier() const;
  // Effects: If checkpoint() > 0, returns the designated replier. Otherwise,
  // returns -1;

  bool verify();
  // Effects: Verifies if the message is correctly authenticated by
  // the replica id().

  static bool convert(Message *m1, Fetch *&m2);
  // Effects: If "m1" has the right size and tag, casts "m1" to a
  // "Fetch" pointer, returns the pointer in "m2" and returns
  // true. Otherwise, it returns false. 
 
private:
  Fetch_rep &rep() const;
  // Effects: Casts contents to a Fetch_rep&

};


inline Fetch_rep &Fetch::rep() const { 
  th_assert(ALIGNED(msg), "Improperly aligned pointer");
  return *((Fetch_rep*)msg); 
}

inline Request_id Fetch::request_id() const { return rep().rid; }

inline  Seqno  Fetch::last_uptodate() const { return rep().lu; }

inline int Fetch::level() const { return rep().level; }

inline int Fetch::index() const { return rep().index; }

inline int Fetch::id() const { return rep().id; }

#ifndef NO_STATE_TRANSLATION
inline int Fetch::chunk_number() const { return rep().chunk_no; }
#endif

inline Seqno Fetch::checkpoint() const { return rep().rc; }

inline int Fetch::replier() const { return rep().repid; }



#endif // _Fetch_h
