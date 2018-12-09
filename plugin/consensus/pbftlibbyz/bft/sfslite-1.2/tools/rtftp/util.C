
#include "rtftp.h"
#include "rtftp_prot.h"
#include "sha1.h"
#include "serial.h"

//-----------------------------------------------------------------------

bool
check_file (const rtftp_file_t &f)
{
  char buf[RTFTP_HASHSZ];
  sha1_hash (buf, f.data.base (), f.data.size ());
  return (memcmp (buf, f.hash.base (), RTFTP_HASHSZ) == 0);
}

//-----------------------------------------------------------------------

static int
dir_ok (const char *d)
{
  int rc;
  struct stat sb;
  if (stat (d, &sb) == 0) {
    if (S_ISDIR (sb.st_mode)) rc = 1;
    else rc = -1;
  } else {
    rc = 0;
  }
  return rc;
}

//-----------------------------------------------------------------------

static const char *
mywd ()
{
#define SZ 1024
  static char buf[SZ];
  return getcwd (buf, SZ);
#undef SZ
}

//-----------------------------------------------------------------------

static int 
mymkdir (char *dir, char *end)
{
  assert (*end == '/');
  assert (dir < end);
  *end = 0;
  int rc = dir_ok (dir);
  if (rc < 0) {
    warn ("cannot access dir %s from dir %s\n", dir, mywd ());
  } else if (rc == 1) {
    rc = 0;
  } else if ((rc = mkdir (dir, 0777)) < 0) {
    warn ("mkdir failed in %s on dir %s: %m\n", mywd (), dir);
  }
  *end = '/';
  return rc;
}

//-----------------------------------------------------------------------

static int
mkdir_p (char *s)
{
  int rc = 0;
  for (char *p = s; rc == 0 && p && *p; ) {
    p = strchr (p, '/');
    if (p && *p) {
      rc = mymkdir (s, p);
      p++;
    }
  }
  if (rc != 0) {
    warn ("cannot create file %s due to mkdir failures\n", s);
  } else if (!s || !*s) {
    warn ("no file to write!\n");
    rc = -1;
  }
  return rc;
}

//-----------------------------------------------------------------------

//
// write out the data 's' to the file given by _file.  Might need to make
// some parent directories if the filename contains slashes.
//
// return -1 if failure, and 0 if wrote the file out OK.
//
int write_file (const str &nm, const str &dat)
{   
  char *s = strdup (nm.cstr ());
  int rc = 0;

  rc = mkdir_p (s);

  if (rc == 0 && !str2file (nm, dat, 0666, false, NULL, true)) {
    warn ("cannot create file %s: %m\n", nm.cstr ());
    rc = -1;
  } else {
    rc = 0;
  }

  if (s) free (s);
  return rc;
}

//-----------------------------------------------------------------------

int
open_file (const str &nm, int flags)
{
  char *s = strdup (nm.cstr ());
  int rc = 0;

  rc = mkdir_p (s);
  if (rc == 0)
    rc = ::open (nm.cstr (), flags, 0666);

  if (s) free (s);
  return rc;
}

//-----------------------------------------------------------------------
