#ifndef _New_key_h
#define _New_key_h 1

#include "types.h"
#include "Digest.h"
#include "Message.h"
#include "Principal.h"

// 
// New_key messages have the following format:
//
struct New_key_rep : public Message_rep {
  Request_id rid;
  int id; // id of the replica that generated the message.
  int padding;

  // Followed by keys for all replicas except "id" in order of
  // increasing identifiers.  Each key has size Nonce_size bytes and
  // is encrypted with the public-key of the corresponding
  // replica. This is all followed by a signature from principal id
};

class New_key : public Message {
  // 
  //  New_key messages
  //
public:
  New_key();
  // Effects: Creates a new signed New_key message and updates "node"
  // accordingly (i.e., updates the in-keys for all principals.) 

  int id() const;
  // Effects: Fetches the identifier of the replica from the message.

  bool verify();
  // Effects: Verifies if the message is signed by the principal
  // rep().id. If the message is correct updates the entry for
  // rep().id accordingly (i.e., out-key, tstamp.)

  static bool convert(Message *m1, New_key *&m2);
  // Effects: If "m1" has the right size and tag of a "New_key",
  // casts "m1" to a "New_key" pointer, returns the pointer in
  // "m2" and returns true. Otherwise, it returns false.
  // If the conversion is successful it trims excess allocation.

private:
  New_key_rep &rep() const;
  // Effects: Casts "msg" to a New_key_rep&
};


inline New_key_rep& New_key::rep() const { 
  th_assert(ALIGNED(msg), "Improperly aligned pointer");
  return *((New_key_rep*)msg); 
}

inline int New_key::id() const { return rep().id; }

#endif // _New_key_h
