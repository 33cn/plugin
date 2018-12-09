#ifndef _FILEID_HASH
#define _FILEID_HASH 1

#include <stdio.h>

#ifdef __cplusplus
extern "C" {
#endif

void init_map_fileid();

void add_fileid(unsigned int key, int value);
      // Checks: the table does not already contain a mapping for the key.
      // Effects: Adds a new mapping to the hash table.

int find_fileid(unsigned int key);
      // Returns the value corresponding to this key
      //    or -1 if no such mapping exists.

int remove_fileid(unsigned int key);
      //    Checks: there is a mapping for "key"
      //    Effects: Removes the mapping for "key" and returns the
      //             corresponding value or -1 if no such mapping exists

void fileid_map_clear();
      //    effects - All bindings are removed 

void save_fileid_map(FILE *o);

void read_fileid_map(FILE *i);

#ifdef __cplusplus
}
#endif

#endif
