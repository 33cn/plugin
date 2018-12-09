#ifndef _Commit_h
#define _Commit_h 1

#include "types.h"
#include "Message.h"
class Principal;

// 
// Commit messages have the following format.
//
struct Commit_rep : public Message_rep {
  View view;       
  Seqno seqno;
  int id;         // id of the replica that generated the message.
  int padding;
  // Followed by a variable-sized signature.
};

class Commit : public Message {
  // 
  // Commit messages
  //
public:
  Commit(View v, Seqno s);
  // Effects: Creates a new Commit message with view number "v"
  // and sequence number "s".

  Commit(Commit_rep *contents);
  // Requires: "contents" contains a valid Commit_rep. If
  // contents may not be a valid Commit_rep use the static
  // method convert.
  // Effects: Creates a Commit message from "contents". No copy
  // is made of "contents" and the storage associated with "contents"
  // is not deallocated if the message is later deleted.

  void re_authenticate(Principal *p=0);
  // Effects: Recomputes the authenticator in the message using the
  // most recent keys. If "p" is not null, may only update "p"'s
  // entry.

  View view() const;
  // Effects: Fetches the view number from the message.

  Seqno seqno() const;
  // Effects: Fetches the sequence number from the message.

  int id() const;
  // Effects: Fetches the identifier of the replica from the message.

  bool match(const Commit *c) const;
  // Effects: Returns true iff this and c match.

  bool verify();
  // Effects: Verifies if the message is signed by the replica rep().id.

  static bool convert(Message *m1, Commit *&m2);
  // Effects: If "m1" has the right size and tag of a "Commit",
  // casts "m1" to a "Commit" pointer, returns the pointer in
  // "m2" and returns true. Otherwise, it returns false. 

  static bool convert(char *m1, unsigned max_len, Commit &m2);
  // Requires: convert can safely read up to "max_len" bytes starting
  // at "m1".
  // Effects: If "m1" has the right size and tag of a
  // "Commit_rep" assigns the corresponding Commit to m2 and
  // returns true.  Otherwise, it returns false.  No copy is made of
  // m1 and the storage associated with "contents" is not deallocated
  // if "m2" is later deleted.

 
private:
  Commit_rep &rep() const;
  // Effects: Casts "msg" to a Commit_rep&
};


inline Commit_rep& Commit::rep() const { 
  th_assert(ALIGNED(msg), "Improperly aligned pointer");
  return *((Commit_rep*)msg); 
}

inline View Commit::view() const { return rep().view; }

inline Seqno Commit::seqno() const { return rep().seqno; }

inline int Commit::id() const { return rep().id; }

inline bool Commit::match(const Commit *c) const {
  th_assert(view() == c->view() && seqno() == c->seqno(), "Invalid argument");
  return true; 
}

#endif // _Commit_h
