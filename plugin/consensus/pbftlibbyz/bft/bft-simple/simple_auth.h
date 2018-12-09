#ifndef _simple_auth_h
#define _simple_auth_h 1

#include "MD5.h"
#define SIMPLE_MAC_SIZE 10

inline void gen_mac(const char *s, unsigned l, char *d) {
  MD5_CTX context;
  unsigned int digest[4];

  MD5Init(&context);
  MD5Update(&context, s, l);
  MD5Update(&context, "ah37dkdnkjsdhjda", 16);
  MD5Final(digest, &context);
  bcopy((char*)digest, d, SIMPLE_MAC_SIZE);
}

inline bool verify_mac(const char *s, unsigned l, const char *mac) {
  char computed[SIMPLE_MAC_SIZE];
  gen_mac(s, l, computed);
  return !bcmp(computed, mac, SIMPLE_MAC_SIZE);
}

#endif // _simple_auth_h
