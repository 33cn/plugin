/* $Id: test_passfd.C 3256 2008-05-14 04:02:23Z max $ */

#include "async.h"

int
main (int argc, char **argv)
{
  int pid;
  int xfd;
  int rfd;
  int wfd;
  int fds[2];
  char msg[] = "Test pattern";
  char buf[sizeof (msg)];

  setprogname (argv[0]);

  if (socketpair (AF_UNIX, SOCK_STREAM, 0, fds) < 0)
    fatal ("socketpair: %m\n");

  pid = fork ();
  if (pid == -1)
    fatal ("fork: %m\n");
  else if (pid == 0) {
    setprogname ( strdup ("child"));
    xfd = fds[0];
    close (fds[1]);
  } else {
    setprogname ( strdup ("parent"));
    xfd = fds[1];
    close (fds[0]);
  }

  if (pipe (fds) < 0)
    fatal ("pipe: %s\n", strerror (errno));
  wfd = fds[1];
  if (writefd (xfd, "", 1, fds[0]) < 0)
    fatal ("writefd: %m\n");
  close (fds[0]);

#if 0
  {
    char c;
    write (xfd, "\173", 1);
    read (xfd, &c, 1);
  }
#endif

  char c;
  if (readfd (xfd, &c, 1, &rfd) < 0)
    fatal ("readfd: %m\n");

  if (pid) {
    if (write (wfd, msg, sizeof (msg)) != sizeof (msg))
      fatal ("write: %s\n", strerror (errno));
    if (read (rfd, buf, sizeof (msg)) != sizeof (msg))
      fatal ("read: %s\n", strerror (errno));
    if (strncmp (msg, buf, sizeof (msg)))
      fatal ("Message corrupt\n");
  }
  else {
    if (read (rfd, buf, sizeof (msg)) != sizeof (msg))
      fatal ("read: %s\n", strerror (errno));
    if (write (wfd, buf, sizeof (msg)) != sizeof (msg))
      fatal ("write: %s\n", strerror (errno));
  }

  exit (0);
}
