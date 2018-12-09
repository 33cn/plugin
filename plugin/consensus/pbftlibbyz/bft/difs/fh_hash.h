#ifndef _FH_HASH
#define _FH_HASH 1

#ifdef __cplusplus
extern "C" {
#endif

struct fhandle;

void init_map();

void add(fhandle *key, int value);
      // Checks: the table does not already contain a mapping for the key.
      // Effects: Adds a new mapping to the hash table.

int find(fhandle *key);
      // Returns the value corresponding to this key
      //    or -1 if no such mapping exists.

int remove_fh(fhandle *key);
      //    Checks: there is a mapping for "key"
      //    Effects: Removes the mapping for "key" and returns the
      //             corresponding value or -1 if no such mapping exists


void fh_map_clear();
      //    effects - All bindings are removed 

#ifdef __cplusplus
}
#endif

#endif
