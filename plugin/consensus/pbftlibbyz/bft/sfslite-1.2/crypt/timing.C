#include "arc4.h"
#include "sha1.h"

const char key[] = "secret key";
const int trials = 1024 * 8;
const int bufsize = 1024;
typedef unsigned char hash_t[160];

int main ()
{
  arc4 a;
  hash_t h;
  char buffer[bufsize];
  struct timeval start, end;

  sha1_hash (h, key, sizeof (key));
  a.setkey (h, sizeof (h));

  gettimeofday (&start, NULL);
  char *e = buffer + sizeof (buffer);
  for (int i = 0; i < trials; i++)
    for (char *p = &buffer[0]; p < e; p++)
      *p ^= a.getbyte ();
  gettimeofday (&end, NULL);

  long long diff = (end.tv_sec * 1000000 + end.tv_usec) - 
    (start.tv_sec * 1000000 + start.tv_usec);

  printf ("Encryption Rate %g KB/s\n",
	  (trials * bufsize / 1000.0) / (diff / 1000000.0));

  gettimeofday (&start, NULL);
  for (int i = 0; i < trials; i++)
    sha1_hash (h, buffer, sizeof (buffer));
  gettimeofday (&end, NULL);

  diff = (end.tv_sec * 1000000 + end.tv_usec) - 
    (start.tv_sec * 1000000 + start.tv_usec);

  printf ("SHA-1 Rate %g KB/s\n",
	  (trials * bufsize / 1000.0) / (diff / 1000000.0));
}
