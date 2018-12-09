#ifndef _STATS_H
#define _STATS_H 1

#ifdef __cplusplus
extern "C" {
#endif

#define STATS 1

#define RECOVER_WRAPPER_STATE 0
#define MAX_NUM_RECOV 5
#define GET_FILE_GET_ATTR 5
#define GET_FILE_RESULT_OBJ 6
#define GET_FILE_READ_CONTENTS 7

#define GET_DIR_GET_ATTR 8
#define GET_DIR_READ_DIR 9
#define GET_DIR_RESULT_OBJ 10

#define GET_FREE 11

#define INIT_REQ 12

#define READ_CACHE_TIME 13
#define LOOKUP_CACHE_TIME 14

#define PUT_SATTR_WRITE_FILE 15
#define PUT_SCAN_UPTODATE 16
#define PUT_READDIR 17
#define PUT_CREATE_MISSING_ENTRIES 18
#define PUT_CHECK_EXISTING_ENTRIES 19

#define MIN_CALL_STATS (PUT_CHECK_EXISTING_ENTRIES + 1)
#define NUM_STATS_PER_CALL 7
#define NUM_CALLS 18
#define MAX_CALLS (MIN_CALL_STATS + NUM_STATS_PER_CALL * NUM_CALLS - 1)

#define NUM_COUNTERS (MAX_CALLS + 1)

void init_stats();

void start_counter(int num);

void stop_counter(int num);

void show_stats();

#ifdef __cplusplus
}
#endif

#endif
