#include <stdlib.h>
#include "th_assert.h"
#include "Message.h"
#include "Node.h"

Log_allocator *Message::a = 0;

Message::Message(unsigned sz) : msg(0), max_size(ALIGNED_SIZE(sz)) {
  if (sz != 0) {
    msg = (Message_rep*) a->malloc(max_size);
    th_assert(ALIGNED(msg), "Improperly aligned pointer");
    msg->tag = -1;
    msg->size = 0;
    msg->extra = 0;
  }
}
 
Message::Message(int t, unsigned sz) {
  max_size = ALIGNED_SIZE(sz);
  msg = (Message_rep*) a->malloc(max_size);
  th_assert(ALIGNED(msg), "Improperly aligned pointer");
  msg->tag = t;
  msg->size = max_size;
  msg->extra = 0;
}
 
Message::Message(Message_rep *cont) {
  th_assert(ALIGNED(cont), "Improperly aligned pointer"); 
  msg = cont;
  max_size = -1; // To prevent contents from being deallocated or trimmed
}
 
Message::~Message() { 
  if (max_size > 0) a->free((char*)msg, max_size); 
}

void Message::trim() {
  if (max_size > 0 && a->realloc((char*)msg, max_size, msg->size)) {
    max_size = msg->size;
  }
}


void Message::set_size(int size) {
  th_assert(msg && ALIGNED(msg), "Invalid state");
  th_assert(max_size < 0 || ALIGNED_SIZE(size) <= max_size, "Invalid state");
  int aligned = ALIGNED_SIZE(size);
  for (int i=size; i < aligned; i++) ((char*)msg)[i] = 0;
  msg->size = aligned;
} 


bool Message::convert(char *src, unsigned len, int t, 
			     int sz, Message &m) {
  // First check if src is large enough to hold a Message_rep
  if (len < sizeof(Message_rep)) return false;

  // Check alignment.
  if (!ALIGNED(src)) return false;

  // Next check tag and message size
  Message ret((Message_rep*)src);
  if (!ret.has_tag(t, sz)) return false;

  m = ret;
  return true;
}


bool Message::encode(FILE* o) {
  int csize = size();

  size_t sz = fwrite(&max_size, sizeof(int), 1, o);
  sz += fwrite(&csize, sizeof(int), 1, o);
  sz += fwrite(msg, 1, csize, o);
  
  return sz == 2U+csize;
}


bool Message::decode(FILE* i) {
  delete msg;
  
  size_t sz = fread(&max_size, sizeof(int), 1, i);
  msg = (Message_rep*) a->malloc(max_size);

  int csize;
  sz += fread(&csize, sizeof(int), 1, i);
  
  if (msg == 0 || csize < 0 || csize > max_size)
    return false;

  sz += fread(msg, 1, csize, i);
  return sz == 2U+csize;
}


void Message::init() { 
  a = new Log_allocator(); 
}


const char *Message::stag() {
  static const char *string_tags[] = {"Free_message", 
				  "Request", 
				  "Reply",
				  "Pre_prepare", 
				  "Prepare",
				  "Commit", 
				  "Checkpoint", 
				  "Status", 
				  "View_change", 
				  "New_view",
				  "View_change_ack", 
				  "New_key", 
				  "Meta_data", 
                                  "Meta_data_d",
				  "Data_tag", 
				  "Fetch",
                                  "Query_stable",
                                  "Reply_stable"};
  return string_tags[tag()];
}
