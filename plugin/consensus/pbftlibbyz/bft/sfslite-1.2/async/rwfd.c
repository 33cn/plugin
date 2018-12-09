/* $Id: rwfd.c 2325 2006-11-14 21:05:30Z max $ */

/*
 *
 * Copyright (C) 2000 David Mazieres (dm@uun.org)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 *
 */

#include "sysconf.h"

#if !defined (HAVE_ACCRIGHTS) && !defined (HAVE_CMSGHDR)
#error Do not know how to pass file descriptors
#endif /* !HAVE_ACCRIGHTS && !HAVE_CMSGHDR */

ssize_t
writevfd (int fd, const struct iovec *iov, int iovcnt, int wfd)
{
  struct msghdr mh;
#ifdef HAVE_CMSGHDR
  struct cmsghdr *cmh;
  char cmhbuf[CMSG_SPACE(sizeof(int))];
#else /* !HAVE_CMSGHDR */
  int fdp[1];
  *fdp = wfd;
#endif /* !HAVE_CMSGHDR */

  bzero (&mh, sizeof mh);
  mh.msg_iov = (struct iovec *) iov;
  mh.msg_iovlen = iovcnt;

#ifdef HAVE_CMSGHDR
  mh.msg_control = (caddr_t)cmhbuf;
  bzero(cmhbuf, sizeof(cmhbuf));
  mh.msg_controllen = CMSG_LEN(sizeof(int));
  cmh = CMSG_FIRSTHDR(&mh);
  cmh->cmsg_level = SOL_SOCKET;
  cmh->cmsg_type = SCM_RIGHTS;
  cmh->cmsg_len = CMSG_LEN(sizeof(int));
  *(int *)CMSG_DATA(cmh) = wfd;
#else /* !HAVE_CMSGHDR */
  mh.msg_accrights = (char *) fdp;
  mh.msg_accrightslen = sizeof (fdp);
#endif /* !HAVE_CMSGHDR */

  return sendmsg (fd, &mh, 0);
}

ssize_t
writefd (int fd, const void *buf, size_t len, int wfd)
{
  struct iovec iov[1];
  iov->iov_base = (void *) buf;
  iov->iov_len = len;
  return writevfd (fd, iov, 1, wfd);
}


ssize_t
readvfd (int fd, const struct iovec *iov, int iovcnt, int *rfdp)
{
  struct msghdr mh;
#ifdef HAVE_CMSGHDR
  char cmhbuf[CMSG_SPACE(sizeof(int))];
  struct cmsghdr *cmh;
#else /* !HAVE_CMSGHDR */
  int fdp[1];
#endif /* !HAVE_CMSGHDR */
  int n;

  bzero (&mh, sizeof mh);
  mh.msg_iov = (struct iovec *) iov;
  mh.msg_iovlen = iovcnt;

#ifdef HAVE_CMSGHDR
  mh.msg_control = (caddr_t) cmhbuf;
  mh.msg_controllen = sizeof(cmhbuf);
#else /* !HAVE_CMSGHDR */
  *fdp = -1;
  mh.msg_accrights = (char *) fdp;
  mh.msg_accrightslen = sizeof (fdp);
#endif /* !HAVE_CMSGHDR */

  n = recvmsg (fd, &mh, 0);

  if (n == -1)
      return n;

  if (n >= 0) {
#ifdef HAVE_CMSGHDR
      *rfdp = -1;
      cmh = CMSG_FIRSTHDR(&mh);
      if (cmh) {
	  if (n == 0) {
	      n = -1;
	      errno = EAGAIN;
	  }
	  if (cmh->cmsg_type == SCM_RIGHTS) {
	      *rfdp = (*(int *)CMSG_DATA(cmh));
	  }
      }
#else /* !HAVE_CMSGHDR */
      *rfdp = *fdp;
      if (n == 0 && *rfdp >= 0) {
	  n = -1;
	  errno = EAGAIN;
      }
#endif /* !HAVE_CMSGHDR */
  }
  return n;
}

ssize_t
readfd (int fd, void *buf, size_t len, int *rfdp)
{
  struct iovec iov[1];
  iov->iov_base = buf;
  iov->iov_len = len;
  return readvfd (fd, iov, 1, rfdp);
}
