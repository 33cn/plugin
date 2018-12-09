#include <stdlib.h>
#include <strings.h>
#include "th_assert.h"
#include "Message_tags.h"
#include "Request.h"
#include "Node.h"
#include "Principal.h"
#include "MD5.h"

#include "Statistics.h"

// extra & 1 = read only
// extra & 2 = signed

Request::Request(Request_id r, short rr) : 
  Message(Request_tag, Max_message_size) {
  rep().cid = node->id();
  rep().rid = r;
  rep().replier = rr;
  rep().command_size = 0;   
  set_size(sizeof(Request_rep));
}


Request* Request::clone() const {
   Request* ret = (Request*)new Message(max_size);
   memcpy(ret->msg, msg, msg->size);
   return ret;
}  

char *Request::store_command(int &max_len) {
  int max_auth_size = MAX(node->principal()->sig_size(), node->auth_size());
  max_len = msize()-sizeof(Request_rep)-max_auth_size;
  return contents()+sizeof(Request_rep);
}


inline void Request::comp_digest(Digest& d) {
  INCR_OP(num_digests);
  START_CC(digest_cycles);
  
  MD5_CTX context;
  MD5Init(&context);
  MD5Update(&context, (char*)&(rep().cid), sizeof(int)+sizeof(Request_id)+rep().command_size);
  MD5Final(d.udigest(), &context);
  
  STOP_CC(digest_cycles);
}


void Request::authenticate(int act_len, bool read_only) {
  th_assert((unsigned)act_len <= 
	    msize()-sizeof(Request_rep)-node->auth_size(), 
	    "Invalid request size");

  rep().extra = ((read_only) ? 1 : 0);
  rep().command_size = act_len;
  if (rep().replier == -1)
    rep().replier = lrand48()%node->n();
  comp_digest(rep().od);
 

  int old_size = sizeof(Request_rep)+act_len;
  set_size(old_size+node->auth_size());
  node->gen_auth_in(contents(), sizeof(Request_rep), contents()+old_size);
}


void Request::re_authenticate(bool change, Principal *p) {
  if (change) {
    rep().extra &= ~1;
  } 
  int new_rep = lrand48()%node->n();
  rep().replier = (new_rep != rep().replier) ? new_rep : (new_rep+1)%node->n();

  int old_size = sizeof(Request_rep)+rep().command_size;
  if ((rep().extra & 2) == 0) {
    node->gen_auth_in(contents(), sizeof(Request_rep), contents()+old_size);
  } else {
    node->gen_signature(contents(), sizeof(Request_rep), contents()+old_size);
  }
}


void Request::sign(int act_len) {
  th_assert((unsigned)act_len <= 
	    msize()-sizeof(Request_rep)-node->principal()->sig_size(), 
	    "Invalid request size");

  rep().extra |= 2;
  rep().command_size = act_len;
  comp_digest(rep().od);

  int old_size = sizeof(Request_rep)+act_len;
  set_size(old_size+node->principal()->sig_size());
  node->gen_signature(contents(), sizeof(Request_rep), contents()+old_size);
}


Request::Request(Request_rep *contents) : Message(contents) {}


bool Request::verify() {
  const int nid = node->id();
  const int cid = client_id();
  const int old_size = sizeof(Request_rep)+rep().command_size;
  Principal* p = node->i_to_p(cid);
  Digest d;

  comp_digest(d);
  if (p != 0 && d == rep().od) {
    if ((rep().extra & 2) == 0) {
      // Message has an authenticator.
      if (cid != nid && cid >= node->n() && size()-old_size >= node->auth_size(cid)) 
	return node->verify_auth_out(cid, contents(), sizeof(Request_rep), 
				       contents()+old_size);
    } else {
      // Message is signed.
      if (size() - old_size >= p->sig_size()) 
	return p->verify_signature(contents(), sizeof(Request_rep), 
				   contents()+old_size, true);
    }
  }
  return false;
}


bool Request::convert(Message *m1, Request  *&m2) {
  if (!m1->has_tag(Request_tag, sizeof(Request_rep)))
    return false;

  m2 = (Request*)m1;
  m2->trim();
  return true;
}


bool Request::convert(char *m1, unsigned max_len, Request &m2) {
  if (!Message::convert(m1, max_len, Request_tag, sizeof(Request_rep),m2)) 
    return false;
  return true;
}
 
