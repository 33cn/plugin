#ifndef _Rep_info_h
#define _Rep_info_h 1

#include <sys/time.h>
#include "types.h"
#include "Time.h"
#include "Digest.h"
#include "Array.h"
#include "Reply.h"
#include "State_defs.h"

class Req_queue;

class Rep_info {
  //
  // Holds the last replies sent to each principal.
  //
public:
#ifndef NO_STATE_TRANSLATION
  Rep_info(int nps);
  // Effects: Creates a new object that stores data for "nps"
  // principals.
#else
  Rep_info(char *mem, int sz, int nps);
  // Requires: "mem" points to an array of "size" bytes and is virtual
  // memory page aligned.
  // Effects: Creates a new object that stores data in "mem" for "nps"
  // principals.
#endif

  ~Rep_info();

  int size() const;
  // Effects: Returns the actual number of bytes (a multiple of the
  // virtual memory page size) that was consumed by this from the
  // start of the "mem" argument supplied to the constructor.

  Request_id req_id(int pid);
  // Requires: "pid" is a valid principal identifier.
  // Effects: Returns the timestamp in the last message sent to
  // principal "pid".

  Digest &digest(int pid);
  // Requires: "pid" is a valid principal identifier.
  // Effects: Returns a reference to the digest of the last reply
  // value sent to pid.

  Reply* reply(int pid);
  // Requires: "pid" is a valid principal identifier.
  // Effects: Returns a pointer to the last reply value sent to "pid"
  // or 0 if no such reply was sent.

  bool new_state(Req_queue *rset);
  // Effects: Updates this to reflect the new state and removes stale
  // requests from rset. If it removes the first request in "rset",
  // returns true; otherwise returns false.
  
  char *new_reply(int pid, int &size);
  // Requires: "pid" is a valid principal identifier.  
  // Effects: Returns a pointer to a buffer where the new reply value
  // for principal "pid" can be placed. The length of the buffer in bytes
  // is returned in "size". Sets the reply to tentative.

  void end_reply(int pid, Request_id rid, int size);
  // Requires: "pid" is a valid principal identifier.
  // Effects: Completes the construction of a new reply value: this is
  // informed that the reply value is size bytes long and computes its
  // digest.

  void commit_reply(int pid);
  // Requires: "pid" is a valid principal identifier.
  // Effects: Mark "pid"'s last reply committed.

  bool is_committed(int pid);
  // Requires: "pid" is a valid principal identifier.
  // Effects: Returns true iff the last reply sent to "pid" is
  // committed.

  void send_reply(int pid, View v, int id, bool tentative=true);
  // Requires: "pid" is a valid principal identifier and end_reply was
  // called after the last call to new_reply for "pid"
  // Effects: Sends a reply message to "pid" for view "v" from replica
  // "id" with the latest reply value stored in the buffer returned by
  // new_reply. If tentative is omitted or true, it sends the reply as
  // tentative unless it was previously committed

#ifndef NO_STATE_TRANSLATION
  char *rep_info_mem();
  // Returns: pointer to the beggining of the mem region used to store the 
  // replies
#endif

private:
  int nps;
  char *mem;
  Array<Reply*> reps; // Array of replies indexed by principal id.
  static const int Max_rep_size = 8192;

  struct Rinfo {
    bool tentative;       // True if last reply is tentative and was not committed.
    Time lsent;           // Time at which reply was last sent.
  };
  Array<Rinfo> ireps;
};

inline int Rep_info::size() const {
  return (nps+1)*Max_rep_size;
}

inline void Rep_info::commit_reply(int pid) {
  ireps[pid].tentative = false;
  ireps[pid].lsent = zeroTime();
}

inline bool Rep_info::is_committed(int pid) {
  return !ireps[pid].tentative;
}

inline Request_id Rep_info::req_id(int pid) {
  return reps[pid]->request_id();
}

inline Digest& Rep_info::digest(int pid) {
  return reps[pid]->digest();
}

inline Reply* Rep_info::reply(int pid) {
  return reps[pid];
}

#ifndef NO_STATE_TRANSLATION
inline char* Rep_info::rep_info_mem() {
  return mem;
}
#endif

#endif // _Rep_info_h
