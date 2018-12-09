#ifndef _View_change_h
#define _View_change_h 1

#include "parameters.h"
#include "types.h"
#include "bits.h"
#include "Digest.h"
#include "Principal.h"
#include "Message.h"


//
// Req_infos describe requests that (1) prepared or (2) for which a
// pre-prepare/prepare message was sent. 
//
// In case (1):
// - the request with digest "d" prepared in view "v" with
// sequence number "n";
// - no request prepared with the same sequence number in a later 
// view; and 
// - the last pre-prepare/prepare sent by the replica for this request 
// was for view "lv".
//
// In case (2):
// - a pre-prepare/prepare was sent for a request with digest "d" in 
// view "v" with sequence number "n"; and
// - no request prepared globally with sequence number "n" in any view 
// "v' <= lv".
//
struct Req_info {
  View lv;
  View v;
  Digest d;
};


// 
// View_change messages have the following format:
//
struct View_change_rep : public Message_rep {
  View v;   // sending replica's new view
  Seqno ls; // sequence number of last checkpoint known to be stable

  // Digests for checkpoints held by the replica in order of
  // increasing sequence number. A null digest means the replica does
  // not have the corresponding checkpoint state.
  Digest ckpts[max_out/checkpoint_interval+1]; 

  int id; // sending replica's id 

  short n_ckpts;  // number of entries in ckpts
  short n_reqs;   // number of entries in req_info

  // Bitmap with bits set for requests that are prepared in req_info
  static const int prepared_size = (max_out+INT_BITS-1)/INT_BITS;
  unsigned prepared[prepared_size];

  // digest of the entire message (except authenticator) with d zeroed.
  Digest d;
 
  /* 
     Followed by:
     Req_info req_info[n_reqs];

     // This is followed by an authenticator from principal id. 
   */
};


class View_change : public Message {
  // 
  //  View_change messages
  //
public:
  View_change(View v, Seqno ls, int id);
  // Effects: Creates a new (unauthenticated) View_change message for
  // replica "id" in view "v". The message states that "ls" is the
  // sequence number of the last checkpoint known to be stable but the
  // message has an empty set of requests and checkpoints.

  void add_checkpoint(Seqno n, Digest& d);
  // Requires: "n%checkpoint_interval = 0", and "last_stable() <= n <=
  // last_stable()+max_out".
  // Effects: Sets the digest of the checkpoint with sequence number
  // "n" to "d".

  void add_request(Seqno n, View v, View lv, Digest &d, bool prepared);
  // Requires: "last_stable() < n <= last_stable()+max_out".
  // Effects: Sets the Req_info for the request with sequence number
  // "n" to "{lv, v, d}" and records whether the request is prepared.

  int id() const;
  // Effects: Fetches the identifier of the replica from the message.

  View view() const;
  // Effects: Returns the view in the message.

  Digest& digest();
  // Effects: Returns the digest of this message (excluding the
  // authenticator).

  Seqno last_stable() const;
  // Effects: Returns the sequence number of the last stable
  // checkpoint.
  
  Seqno max_seqno() const;
  // Effects: Returns the maximum sequence number refered to in this.

  bool last_ckpt(Digest &d, Seqno &n);
  // Effects: If this contains some checkpoint digest, returns true
  // and sets "d" to the digest of the checkpoint with the highest
  // sequence number "n" int this.

  bool ckpt(Seqno n, Digest &d);
  // Effects: If there is a checkpoint with sequence number "n" in the
  // message, sets "d" to its digest and returns true. Otherwise,
  // returns false without modifying "d".
  
  bool proofs(Seqno n, View &v, View &lv, Digest &d, bool &prepared);
  // Effects: If there is a request with sequence number "n" in the
  // message, sets "v, lv", and "d" to the values in the request's
  // Req_info, sets prepared to true iff the request is prepared and
  // returns true. Otherwise, returns false without other effects. 

  View req(Seqno n, Digest &d);
  // Requires: n > last_stable()
  // Effects: Returns the view and sets "d" to the digest associated with
  // the request with sequence number "n" in the message.

  void re_authenticate(Principal *p=0);
  // Effects: Recomputes the authenticator in the message using the
  // most recent keys. If "p" is not null, may only update "p"'s
  // entry in the authenticator.
  
  bool verify();
  // Effects: Verifies if the message is syntactically correct and
  // authenticated by the principal rep().id.

  bool verify_digest();
  // Effects: Returns true iff digest() is correct.

  static bool convert(Message *m1, View_change *&m2);
  // Effects: If "m1" has the right size and tag of a "View_change",
  // casts "m1" to a "View_change" pointer, returns the pointer in
  // "m2" and returns true. Otherwise, it returns false.
  // If the conversion is successful it trims excess allocation.

private:
  View_change_rep &rep() const;
  // Effects: Casts "msg" to a View_change_rep&
  
  Req_info *req_info();
  // Effects: Returns a pointer to the prep_info array.
    
  void mark(int i);
  // Effects: Marks request with index i (sequence number
  // "i+last_stable+1") prepared.

  bool test(int i);
  // Effects: Returns true iff the request with index i (sequence
  // number "i+last_stable+1") is prepared.
};


inline View_change_rep &View_change::rep() const { 
  th_assert(ALIGNED(msg), "Improperly aligned pointer");
  return *((View_change_rep*)msg); 
}
  
inline Req_info *View_change::req_info() { 
  Req_info *ret = (Req_info *)(contents()+sizeof(View_change_rep));
  return ret; 
}

inline void View_change::mark(int index) {
  th_assert(index >= 0 && index < View_change_rep::prepared_size*INT_BITS, "Out of bounds");
  unsigned *chunk = rep().prepared+index/INT_BITS;
  *chunk |=  (1 << (index%INT_BITS));
}

inline bool View_change::test(int index) {
  th_assert(index >= 0 && index < View_change_rep::prepared_size*INT_BITS, "Out of bounds");
  unsigned chunk = *(rep().prepared+index/INT_BITS);
  return (chunk & (1 << (index%INT_BITS))) ? true : false;
}

inline int View_change::id() const { return rep().id; }

inline  View View_change::view() const { return rep().v; }

inline  Digest& View_change::digest() { 
  return rep().d;
}

inline  Seqno View_change::last_stable() const { return rep().ls; }

inline Seqno View_change::max_seqno() const { return rep().ls + rep().n_reqs; }

inline  bool View_change::last_ckpt(Digest& d, Seqno &n) { 
  if (rep().n_ckpts > 0) {
    d = rep().ckpts[rep().n_ckpts-1];
    n = (rep().n_ckpts-1)*checkpoint_interval+rep().ls;
    th_assert(d != Digest(), "Invalid state");

    return true;
  }

  return false;
}

#endif // _View_change_h
