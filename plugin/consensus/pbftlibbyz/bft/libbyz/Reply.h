#ifndef _Reply_h
#define _Reply_h 1

#include "types.h"
#include "Message.h"
#include "Digest.h"

class Principal;
class Rep_info;

// 
// Reply messages have the following format.
//
struct Reply_rep : public Message_rep {
  View v;                // current view
  Request_id rid;        // unique request identifier
  Digest digest;         // digest of reply.
  int replica;           // id of replica sending the reply
  int reply_size;        // if negative, reply is not full.
  // Followed by a reply that is "reply_size" bytes long and
  // a MAC authenticating the reply to the client. The MAC is computed
  // only over the Reply_rep. Replies can be empty or full. Full replies
  // contain the actual reply and have reply_size >= 0. Empty replies
  // do not contain the actual reply and have reply_size < 0.
};

class Reply : public Message {
  // 
  // Reply messages
  //
public:
  Reply() : Message() {}

  Reply(Reply_rep *r);

  Reply(View view, Request_id req, int replica);
  // Effects: Creates a new (full) Reply message with an empty reply and no
  // authentication. The method store_reply and authenticate should
  // be used to finish message construction.

  Reply* copy(int id) const;
  // Effects: Creates a new object with the same state as this but
  // with replica identifier "id"

  char *store_reply(int &max_len);
  // Effects: Returns a pointer to the location within the message
  // where the reply should be stored and sets "max_len" to the number of
  // bytes available to store the reply. The caller can copy any reply
  // with length less than "max_len" into the returned buffer. 

  void authenticate(Principal *p, int act_len, bool tentative);
  // Effects: Terminates the construction of a reply message by
  // setting the length of the reply to "act_len", appending a MAC,
  // and trimming any surplus storage.

  void re_authenticate(Principal *p);
  // Effects: Recomputes the authenticator in the reply using the most
  // recent key.

  Reply(View view, Request_id req, int replica, Digest &d, 
	Principal *p, bool tentative);
  // Effects: Creates a new empty Reply message and appends a MAC for principal
  // "p".

  void commit(Principal *p);
  // Effects: If this is tentative converts this into an identical
  // committed message authenticated for principal "p".  Otherwise, it
  // does nothing.

  View view() const;
  // Effects: Fetches the view from the message

  Request_id request_id() const;
  // Effects: Fetches the request identifier from the message.

  int id() const;
  // Effects: Fetches the reply identifier from the message.

  Digest &digest() const;
  // Effects: Fetches the digest from the message.

  char *reply(int &len);
  // Effects: Returns a pointer to the reply and sets len to the
  // reply size.
  
  bool full() const;
  // Effects: Returns true iff "this" is a full reply.

  bool verify();
  // Effects: Verifies if the message is authenticated by rep().replica.

  bool match(Reply *r);
  // Effects: Returns true if the replies match.
  
  bool is_tentative() const;
  // Effects: Returns true iff the reply is tentative.

  static bool convert(Message *m1, Reply *&m2);
  // Effects: If "m1" has the right size and tag of a "Reply", casts
  // "m1" to a "Reply" pointer, returns the pointer in "m2" and
  // returns true. Otherwise, it returns false. Convert also trims any
  // surplus storage from "m1" when the conversion is successfull.

private:
  Reply_rep &rep() const;
  // Effects: Casts "msg" to a Reply_rep&
  
  friend class Rep_info;
};

inline Reply_rep& Reply::rep() const { 
  th_assert(ALIGNED(msg), "Improperly aligned pointer");
  return *((Reply_rep*)msg); 
}
   
inline View Reply::view() const { return rep().v; }

inline Request_id Reply::request_id() const { return rep().rid; }

inline int Reply::id() const { return rep().replica; }

inline Digest& Reply::digest() const { return rep().digest; }

inline char* Reply::reply(int &len) { 
  len = rep().reply_size;
  return contents()+sizeof(Reply_rep);
}
  
inline bool Reply::full() const { return rep().reply_size >= 0; }

inline bool Reply::is_tentative() const { return rep().extra; }

inline bool Reply::match(Reply *r) {
  return (rep().digest == r->rep().digest) & 
    (!is_tentative() | r->is_tentative() | (view() == r->view()));
}
  


#endif // _Reply_h
