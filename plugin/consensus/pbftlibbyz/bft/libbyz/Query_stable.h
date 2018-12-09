#ifndef _Query_stable_h
#define _Query_stable_h 1

#include "types.h"
#include "Message.h"
#include "Principal.h"

// 
// Query_stable messages have the following format:
//
struct Query_stable_rep : public Message_rep {
  int id;         // id of the replica that generated the message.
  int nonce;
  // Followed by a variable-sized signature.
};

class Query_stable : public Message {
  // 
  //  Query_stable messages
  //
public:
  Query_stable();
  // Effects: Creates a new authenticated Query_stable message.

  void re_authenticate(Principal *p=0);
  // Effects: Recomputes the authenticator in the message using the
  // most recent keys. 

  int id() const;
  // Effects: Fetches the identifier of the replica from the message.

  int nonce() const;
  // Effects: Fetches the nonce in the message.

  bool verify();
  // Effects: Verifies if the message is signed by the replica rep().id.

  static bool convert(Message *m1, Query_stable *&m2);
  // Effects: If "m1" has the right size and tag of a "Query_stable",
  // casts "m1" to a "Query_stable" pointer, returns the pointer in
  // "m2" and returns true. Otherwise, it returns false. Convert also
  // trims any surplus storage from "m1" when the conversion is
  // successfull.
 
private:
  Query_stable_rep& rep() const;
  // Effects: Casts "msg" to a Query_stable_rep&
};

inline Query_stable_rep& Query_stable::rep() const { 
  th_assert(ALIGNED(msg), "Improperly aligned pointer");
  return *((Query_stable_rep*)msg); 
}

inline int Query_stable::id() const { return rep().id; }

inline int Query_stable::nonce() const { return rep().nonce; }

#endif // _Query_stable_h
