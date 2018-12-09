#ifndef _Big_req_table_h
#define _Big_req_table_h 1

#include "types.h"
#include "Digest.h"
#include "Array.h"
#include "map.h"


class BR_entry;
class Request;


class Big_req_table {
  //
  // Overview: Table that holds entries for big requests. The entries
  // contain the requests digest and a pointer to the request or if
  // the request is not cached a list of pre-prepare messages the
  // reference the request. These entries are used to match
  // pre-prepares with their big requests. (Big requests are those
  // whose size is greater than Request::big_req_thresh.)
  //
public:
  Big_req_table();
  // Effects: Creates an empty table.

  ~Big_req_table();
  // Effects: Deletes table and any requests it references.

  void add_pre_prepare(Request* r, Seqno n, View v);
  // Effects: Records that request "r" is referenced by the
  // pre-prepare with sequence number "n" and that this information is
  // current in view "v".
  
  bool add_pre_prepare(Digest& rd, int i, Seqno n, View v);
  // Effects: Records that the i-th reference to a big request in the
  // pre-prepare with sequence number "n" is to the request with
  // digest "rd", and that this information is current in view
  // "v". Returns true if the request is in the table; otherwise,
  // returns false.

  bool add_request(Request* r, bool verified=true);
  // Requires: r->size() > Request::big_req_thresh & verified == r->verify() 
  // Effects: If there is an entry for digest "r->digest()", the entry
  // does not already contain a request and the authenticity of the
  // request can be verified, then it adds "r" to the entry, calls
  // "add_request" on each pre-prepare-info whose pre-prepare is
  // waiting on the entry, and returns true. Otherwise, returns false
  // and has no other effects (in particular it does not delete "r").

  Request* lookup(Digest& rd);
  // Effects: Returns the request in this with digest "rd" or 0 if
  // there is no such request.

  void mark_stable(Seqno ls);
  // Effects: Discards entries that were only referred to by
  // pre-prepares that were discarded due to checkpoint "ls" becoming
  // stable.

  void view_change(View v);
  // Effects: Discards entries that were only referred to by
  // pre-prepares that were discarded due to view changing to view
  // "v".

private:
  bool check_pcerts(BR_entry* bre);
  // Requires: replica->has_new_view()
  // Effects: Returns true iff there is some pre-prepare in
  // bre->waiting that has f matching prepares in its prepared
  // certificate.

  void remove_unmatched(BR_entry* bre);
  // Requires: bre->r != 0
  // Effects: Zeros entry in unmatched.

  Map<Digest,BR_entry*> breqs;
  int max_entries;  // Maximum number of entries allowed in the table.
  Array<Request*> unmatched; // Array of requests that have no waiting pre-prepares
                             // indexed by client id.
};

#endif // _Big_req_table_h
