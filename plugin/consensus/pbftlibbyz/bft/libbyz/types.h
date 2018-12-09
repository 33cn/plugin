#ifndef _types_h
#define _types_h 1

/*
 * Definitions of various types.
 */
#include <limits.h>
#include <sys/socket.h>
#include <sys/param.h>
#include <netinet/in.h> 
#include <sys/time.h>
#include <unistd.h>

#ifndef MAX
#define MAX(a,b) ((a) > (b) ? (a) : (b))
#endif


#ifdef __alpha
typedef long Long;
typedef unsigned long ULong;
#else
typedef long long Long;
typedef unsigned long long ULong;
#endif

typedef Long Seqno;
typedef Seqno View;
typedef ULong Request_id; 

typedef struct sockaddr_in Addr;

#ifdef __alpha
static const Long Long_max = 9223372036854775807L;
static const View View_max = 9223372036854775807L;
static const Seqno Seqno_max = 9223372036854775807L;
#else
static const Long Long_max = 9223372036854775807LL;
static const View View_max = 9223372036854775807LL;
static const Seqno Seqno_max = 9223372036854775807LL;
#endif

typedef Long BR_map;

#endif // _types_h
