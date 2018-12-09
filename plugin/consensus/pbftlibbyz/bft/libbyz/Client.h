#ifndef _Client_h
#define _Client_h 1

#include <stdio.h>
#include "types.h"
#include "Node.h"
#include "Certificate.h"

class Reply;
class Request;
class ITimer;
extern void rtimer_handler();

class Client : public Node {
public:
  Client(FILE *config_file, FILE *config_priv, short port=0);
  // Effects: Creates a new Client object using the information in
  // "config_file" and "config_priv". The line of config assigned to
  // this client is the first one with the right host address (if
  // port==0) or the first with the right host address and port equal
  // to "port".

  virtual ~Client();
  // Effects: Deallocates all storage associated with this.

  bool send_request(Request *req);
  // Effects: Sends request m to the service. Returns FALSE iff two
  // consecutive request were made without waiting for a reply between
  // them.

  Reply *recv_reply();
  // Effects: Blocks until it receives enough reply messages for
  // the previous request. returns a pointer to the reply. The caller is
  // responsible for deallocating the request and reply messages.

  Request_id get_rid() const;
  // Effects: Returns the current outstanding request identifier. The request
  // identifier is updated to a new value when the previous message is
  // delivered to the user.

  void reset();
  // Effects: Resets client state to ensure independence of experimental
  // points.

private:
  Request *out_req;     // Outstanding request
  bool need_auth;       // Whether to compute new authenticator for out_req
  Request_id out_rid;   // Identifier of the outstanding request
  int n_retrans;        // Number of retransmissions of out_req
  int rtimeout;         // Timeout period in msecs

  // Maximum retransmission timeout in msecs
  static const int Max_rtimeout = 1000;

  // Minimum retransmission timeout after retransmission 
  // in msecs
  static const int Min_rtimeout = 10;

  Cycle_counter latency; // Used to measure latency.

  // Multiplier used to obtain retransmission timeout from avg_latency
  static const int Rtimeout_mult = 4; 

  Certificate<Reply> t_reps; // Certificate with tentative replies (size 2f+1)
  Certificate<Reply> c_reps; // Certificate with committed replies (size f+1)

  friend void rtimer_handler();
  ITimer *rtimer;       // Retransmission timer

  void retransmit();
  // Effects: Retransmits any outstanding request and last new-key message.

  void send_new_key();
  // Effects: Calls Node's send_new_key, and cleans up stale replies in
  // certificates.
};

inline Request_id Client::get_rid() const { return out_rid; } 

#endif // _Client_h

