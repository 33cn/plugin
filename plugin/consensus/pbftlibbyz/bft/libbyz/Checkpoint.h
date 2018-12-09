#ifndef _Checkpoint_h
#define _Checkpoint_h 1

#include "types.h"
#include "Digest.h"
#include "Message.h"
class Principal;

// 
// Checkpoint messages have the following format:
//
struct Checkpoint_rep : public Message_rep {
  Seqno seqno;
  Digest digest;
  int id;         // id of the replica that generated the message.
  int padding;
  // Followed by a variable-sized signature.
};

class Checkpoint : public Message {
  // 
  //  Checkpoint messages
  //
public:
  Checkpoint(Seqno s, Digest &d, bool stable=false);
  // Effects: Creates a new signed Checkpoint message with sequence
  // number "s" and digest "d". "stable" should be true iff the checkpoint
  // is known to be stable.

  void re_authenticate(Principal *p=0, bool stable=false);
  // Effects: Recomputes the authenticator in the message using the
  // most recent keys. "stable" should be true iff the checkpoint is
  // known to be stable.  If "p" is not null, may only update "p"'s
  // entry. XXXX two default args is dangerous try to avoid it

  Seqno seqno() const;
  // Effects: Fetches the sequence number from the message.

  int id() const;
  // Effects: Fetches the identifier of the replica from the message.

  Digest &digest() const;
  // Effects: Fetches the digest from the message.

  bool stable() const;
  // Effects: Returns true iff the sender of the message believes the
  // checkpoint is stable.

  bool match(const Checkpoint *c) const;
  // Effects: Returns true iff "c" and "this" have the same digest

  bool verify();
  // Effects: Verifies if the message is signed by the replica rep().id.

  static bool convert(Message *m1, Checkpoint *&m2);
  // Effects: If "m1" has the right size and tag of a "Checkpoint",
  // casts "m1" to a "Checkpoint" pointer, returns the pointer in
  // "m2" and returns true. Otherwise, it returns false. Convert also
  // trims any surplus storage from "m1" when the conversion is
  // successfull.
 
private:
  Checkpoint_rep& rep() const;
  // Effects: Casts "msg" to a Checkpoint_rep&
};

inline Checkpoint_rep& Checkpoint::rep() const { 
  th_assert(ALIGNED(msg), "Improperly aligned pointer");
  return *((Checkpoint_rep*)msg); 
}

inline Seqno Checkpoint::seqno() const { return rep().seqno; }

inline int Checkpoint::id() const { return rep().id; }

inline Digest& Checkpoint::digest() const { return rep().digest; }

inline bool Checkpoint::stable() const { return rep().extra == 1; }

inline bool Checkpoint::match(const Checkpoint *c) const { 
  th_assert(seqno() == c->seqno(), "Invalid argument");
  return digest() == c->digest(); 
}

#endif // _Checkpoint_h
