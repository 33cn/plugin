
// -*-c++-*-

#include "async.h"
#include "rtftp_prot.h"

bool check_file (const rtftp_file_t &f);
int write_file (const str &nm, const str &dat);
int open_file (const str &nm, int d);
