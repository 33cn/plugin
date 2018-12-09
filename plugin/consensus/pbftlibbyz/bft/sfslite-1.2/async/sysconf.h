/* $Id: sysconf.h 3758 2008-11-13 00:36:00Z max $ */

/*
 *
 * Copyright (C) 1998 David Mazieres (dm@uun.org)
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


#ifndef _ASYNC_SYSCONF_H_
#define _ASYNC_SYSCONF_H_ 1


#ifdef __cplusplus
extern "C" {
#endif /* __cplusplus */

#ifdef HAVE_CONFIG_H
#include <config.h>
#endif /* HAVE_CONFIG_H */

#include "autoconf.h"

#ifdef __cplusplus
#undef inline
#undef const
#endif /* __cplusplus */

#if __GNUC__ < 2
/* Linux has some pretty broken header files */
#undef _EXTERN_INLINE
#define _EXTERN_INLINE static inline
#endif /* !gcc2 */

#define PORTMAP

#ifdef TIME_WITH_SYS_TIME
# include <sys/time.h>
# include <time.h>
#elif defined (HAVE_SYS_TIME_H)
# include <sys/time.h>
#else /* !TIME_WITH_SYS_TIME && !HAVE_SYS_TIME_H */
# include <time.h>
#endif /* !TIME_WITH_SYS_TIME && !HAVE_SYS_TIME_H */

#ifdef HAVE_TIMES
# include <sys/times.h>
#endif /* HAVE_TIMES */

#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include <stddef.h>
#include <ctype.h>

#include <limits.h>
#include <unistd.h>

#include <sys/stat.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <sys/uio.h>
#include <sys/param.h>
#include <sys/mman.h>

#include <sys/ioctl.h>
#ifdef HAVE_SYS_SOCKIO_H
# include <sys/sockio.h>
#endif /* HAVE_SYS_SOCKIO_H */
#ifdef HAVE_SYS_FILIO_H
# include <sys/filio.h>
#endif /* HAVE_SYS_FILIO_H */
#ifdef HAVE_SYS_FILE_H
# include <sys/file.h>
#endif /* HAVE_SYS_FILE_H */

#if HAVE_SYS_WAIT_H
# include <sys/wait.h>
#endif
#ifndef WEXITSTATUS
# define WEXITSTATUS(stat_val) ((unsigned)(stat_val) >> 8)
#endif
#ifndef WIFEXITED
# define WIFEXITED(stat_val) (((stat_val) & 255) == 0)
#endif

#include <sys/resource.h>
#ifdef HAVE_SYS_RUSAGE_H
# include <sys/rusage.h>
#endif /* HAVE_SYS_RUSAGE_H */
#ifdef NEED_GETRUSAGE_DECL
int getrusage (int who, struct rusage *rusage);
#endif /* NEED_GETRUSAGE_DECL */
#ifdef NEED_FCHDIR_DECL
int fchdir (int);
#endif /* NEED_FCHDIR_DECL */

#include <assert.h>
#include <errno.h>
#include <fcntl.h>
#include <netdb.h>
#include <pwd.h>
#ifdef SETGROUPS_NEEDS_GRP_H
#include <grp.h>
#endif /* SETGROUPS_NEEDS_GRP_H */
#include <signal.h>

/* Some systems have an ip_opts field in the ip_opts structure */
#define __concatenate(t1, t2) t1##t2
#define concatenate(t1, t2) __concatenate(t1, t2)
#define ip_opts concatenate(ip_opts_kludge, __LINE__)
#include <netinet/in.h>
#undef __concatenate
#undef concatenate
#undef ip_opts
#include <arpa/inet.h>

#if defined(PYMALLOC) && defined(DMALLOC)
# undef DMALLOC
#endif /* PYMALLOC && DMALLOC */

#if STDC_HEADERS
# include <string.h>
# ifndef bzero
#  define bzero(a,b)   memset((a), 0, (b))
# endif /* !bzero */
#else /* !STDC_HEADERS */
# ifndef DMALLOC
#  ifndef HAVE_STRCHR
#   define strchr index
#   define strrchr rindex
#  endif /* !HAVE_STRCHR */
#  ifdef __cplusplus
char *strchr (const char *, int);
char *strrchr (const char *, int);
#  else /* !__cplusplus */
char *strchr ();
char *strrchr ();
#  endif /* !__cplusplus */
#  ifdef HAVE_MEMCPY
#   define bzero(a,b)   memset((a), 0, (b))
#  else /* !HAVE_MEMCPY */
#   define memcpy(d, s, n) bcopy ((s), (d), (n))
#   define memmove(d, s, n) bcopy ((s), (d), (n))
#  endif /* !HAVE_MEMCPY */
# endif /* !DMALLOC || PYMALLOC */
#endif


#if defined (__linux__) && !defined (__GNUC__)
/* GNU libc has some really really broken header files. */
# ifdef __cplusplus
extern "C++" {
inline bool
operator== (dev_t d1, dev_t d2)
{
  return !memcmp (&d1, &d2, sizeof (d1));
}
inline bool
operator!= (dev_t d1, dev_t d2)
{
  return memcmp (&d1, &d2, sizeof (d1));
}
}
# endif /* __cplusplus */
#endif /* __linux__ && !__GNUC__ */

#ifndef UIO_MAXIOV
#define UIO_MAXIOV 16
#endif /* UIO_MAXIOV */

/*
 * Compiler/arhcitecture attributes
 */

#if __GNUC__ >= 2 
# ifdef __cplusplus
#  if __GNUC__ == 2 && __GNUC_MINOR__ < 91 /* !egcs */
#   define NO_TEMPLATE_FRIENDS 1
#  endif /* !egcs */
#  define PRIVDEST friend class stupid_gcc_disallows_private_destructors;
# endif /* __cplusplus */
#else /* !gcc2 */
# ifndef __attribute__
#  define __attribute__(x)
# endif /* !__attribute__ */
#endif /* !gcc 2 */

#if __GNUC__ >= 3
# define Xtmpl
#else /* gcc < 3 */
# define Xtmpl template
#endif /* gcc < 3 */

#ifndef PRIVDEST
#define PRIVDEST /* empty */
#endif /* !PRIVDEST */

#ifndef HAVE_SSIZE_T
typedef int ssize_t;
#endif /* !HAVE_SSIZE_T */
#ifndef HAVE_INT32_T
typedef int int32_t;
#endif /* !HAVE_INT32_T */
#ifndef HAVE_U_INT32_T
typedef unsigned int u_int32_t;
#endif /* !HAVE_U_INT32_T */
#ifndef HAVE_U_INT16_T
typedef unsigned short u_int16_t;
#endif /* !HAVE_U_INT16_T */
#ifndef HAVE_U_INT8_T
typedef unsigned char u_int8_t;
#endif /* !HAVE_U_INT8_T */
#ifndef HAVE_MODE_T
typedef unsigned short mode_t;
#endif /* !HAVE_MODE_T */
#ifndef HAVE_U_CHAR
typedef unsigned char u_char;
#endif /* !HAVE_U_CHAR */
#ifndef HAVE_U_INT
typedef unsigned int u_int;
#endif /* !HAVE_U_INT */
#ifndef HAVE_U_LONG
typedef unsigned long u_long;
#endif /* !HAVE_U_LONG */
#ifndef HAVE_RLIM_T
typedef int rlim_t;
#endif /* !HAVE_RLIM_T */

#ifndef HAVE_INT64_T
# if SIZEOF_LONG == 8
typedef long int64_t;
# elif SIZEOF_LONG_LONG == 8
typedef long long int64_t;
# else /* Can't find 64-bit type */
#  error "Cannot find any 64-bit data types"
# endif /* !SIZEOF_LONG_LONG */
#endif /* !HAVE_INT64_T */

#ifndef HAVE_U_INT64_T
# if SIZEOF_LONG == 8
typedef unsigned long u_int64_t;
# elif SIZEOF_LONG_LONG == 8
typedef unsigned long long u_int64_t;
# else /* Can't find 64-bit type */
#  error "Cannot find any 64-bit data types"
# endif /* !SIZEOF_LONG_LONG */
# define HAVE_U_INT64_T 1	/* XXX */
#endif /* !HAVE_INT64_T */

#if SIZEOF_LONG == 8
# define INT64(n) n##L
#elif SIZEOF_LONG_LONG == 8
# define INT64(n) n##LL
#else /* Can't find 64-bit type */
# error "Cannot find any 64-bit data types"
#endif /* !SIZEOF_LONG_LONG */

#ifdef U_INT64_T_IS_LONG_LONG
# if defined (__sun__) && defined (__svr4__)
#  define U64F "ll"
# else /* everyone else */
#  define U64F "q"
# endif /* everyone else */
#else /* !U_INT64_T_IS_LONG_LONG */
# define U64F "l"
#endif /* !U_INT64_T_IS_LONG_LONG */

#ifdef WORDS_BIGENDIAN
#define htonq(n) n
#define ntohq(n) n
#else /* little endian */
static inline u_int64_t
htonq (u_int64_t v)
{
  return htonl ((u_int32_t) (v >> 32))
    | (u_int64_t) htonl ((u_int32_t) v) << 32;
}
static inline u_int64_t
ntohq (u_int64_t v)
{
  return ntohl ((u_int32_t) (v >> 32))
    | (u_int64_t) ntohl ((u_int32_t) v) << 32;
}
#endif /* little endian */


/*
 * OS/library features
 */

#ifdef HAVE_PREAD
# ifndef HAVE_PREAD_DECL
ssize_t pread(int, void *, size_t, off_t);
# endif /* !HAVE_PREAD_DECL */
#endif /* !HAVE_PREAD */

#ifdef HAVE_PWRITE
# ifndef HAVE_PWRITE_DECL
ssize_t pwrite(int, const void *, size_t, off_t);
# endif /* !HAVE_PWRITE_DECL */
#endif /* !HAVE_PWRITE */

/* XXX - some OSes don't put the __attribute__ ((noreturn)) on exit
 * and abort.  However, on some linuxes redeclaring these functions
 * causes compilation errors, as the originals are declared throw ()
 * for C++. */
#ifdef __THROW
void abort (void) __THROW __attribute__ ((noreturn));
void exit (int) __THROW __attribute__ ((noreturn));
#else /* !__THROW */
void abort (void) __attribute__ ((noreturn));
void exit (int) __attribute__ ((noreturn));
#endif /* !__THROW */
void _exit (int) __attribute__ ((noreturn));

#if !defined(HAVE_STRERROR) && !defined(strerror)
extern int sys_nerr;
extern char *sys_errlist[];
#define strerror(n) ((unsigned) (n) < sys_nerr \
		     ? sys_errlist[n] : "large error number")
#endif /* !HAVE_STRERROR && !strerror */

#ifndef INADDR_NONE
# define INADDR_NONE 0xffffffffU
#endif /* !INADDR_NONE */

#ifndef HAVE_INET_ATON
# define inet_aton(a,b)  (((b)->s_addr = inet_addr((a))) != INADDR_NONE)
#endif

/* constants for flock */
#ifndef LOCK_SH
# define LOCK_SH 1		/* shared lock */
# define LOCK_EX 2		/* exclusive lock */
# define LOCK_NB 4		/* don't block when locking */
# define LOCK_UN 8		/* unlock */
#endif /* !LOCK_SH */
#ifdef NEED_FLOCK_DECL
int flock (int fd, int operation);
#endif /* NEED_FLOCK_DECL */

#ifndef SHUT_RD
# define SHUT_RD 0
# define SHUT_WR 1
# define SHUT_RDWR 2
#endif /* !SHUT_RD */

#if defined (SIGPOLL) && !defined (SIGIO)
# define SIGIO SIGPOLL
#endif /* SIGPOLL && !SIGIO */

#if defined (FASYNC) && !defined (O_ASYNC)
# define O_ASYNC FASYNC
#endif /* FASYNC && !O_ASYNC */

#ifndef MAP_FILE
# define MAP_FILE 0
#endif /* !MAP_FILE */

#if defined (MS_ASYNC) || defined (MS_SYNC)
# define HAVE_MSYNC_FLAGS 1
# ifndef MS_SYNC
#  define MS_SYNC 0
# endif /* !MS_SYNC */
# ifndef MS_ASYNC
#  define MS_SYNC 0
# endif /* !MS_ASYNC */
#endif /* MS_ASYNC || MS_SYNC */

/* Format specifier for printing off_t variables */
#ifdef HAVE_OFF_T_64
# define OTF "q"
#else /* !HAVE_OFF_T_64 */
# define OTF "l"
#endif /* !HAVE_OFF_T_64 */

/* Type of iov_base */
typedef char *iovbase_t;

#ifndef HAVE_SOCKLEN_T
#define socklen_t int
#endif /* !HAVE_SOCKLEN_T */

#ifndef HAVE_STRUCT_SOCKADDR_STORAGE
#define sockaddr_storage sockaddr
#endif /* !HAVE_STRUCT_SOCKADDR_STORAGE */

#ifndef HAVE_TIMESPEC
struct timespec {
  time_t tv_sec;
  long tv_nsec;
};
#endif /* !HAVE_TIMESPEC */


#ifndef HAVE_ST_MTIMESPEC
# ifdef HAVE_ST_MTIM
#  define st_mtimespec st_mtim
#  define HAVE_ST_MTIMESPEC 1
# endif /* HAVE_ST_MTIM */
#endif /* !HAVE_ST_MTIMESPEC */


#ifdef NEED_MKSTEMP_DECL
int mkstemp (char *);
#endif /* NEED_MKSTEMP_DECL */


#ifndef HAVE_CLOCK_GETTIME_DECL
int clock_gettime (int, struct timespec *);
#endif /* !HAVE_CLOCK_GETTIME_DECL */

#ifndef HAVE_CLOCK_GETTIME
# undef CLOCK_REALTIME
# undef CLOCK_VIRTUAL
# undef CLOCK_PROF
#endif /* !HAVE_CLOCK_GETTIME */

#ifndef CLOCK_REALTIME
# define CLOCK_REALTIME 0
#endif /* !CLOCK_REALTIME */
#ifndef CLOCK_VIRTUAL
# define CLOCK_VIRTUAL 1
#endif /* !CLOCK_VIRTUAL */
#ifndef CLOCK_PROF
# define CLOCK_PROF 2
#endif /* !CLOCK_PROF */

#ifndef PATH_MAX
# if defined (MAXPATH)
#  define PATH_MAX MAXPATH
# elif defined (FILENAME_MAX)
#  define PATH_MAX FILENAME_MAX
# else /* !PATH_MAX && !FILENAME_MAX */
#  define PATH_MAX 1024
# endif /* !PATH_MAX && !FILENAME_MAX */
#endif /* PATH_MAX */


#if 0 /* doesn't work on new linuxes, just forget it */
# ifndef HAVE_GETPEEREID
#  ifdef SO_PEERCRED
static inline int
getpeereid (int fd, uid_t *u, gid_t *g)
{
  struct ucred cred;
  socklen_t credlen = sizeof (cred);
  int r = getsockopt (fd, SOL_SOCKET, SO_PEERCRED, &cred, &credlen);
  if (r >= 0) {
    if (u)
      *u = cred.uid;
    if (g)
      *g = cred.gid;
  }
  return r;
}
#  define HAVE_GETPEEREID 1
#  endif /* SO_PEERCRED */
# endif /* !HAVE_GETPEEREID */
#endif /* 0 */

/*
 * Debug malloc
 */

#define __stringify(s) #s
#define stringify(s) __stringify(s)
#define __FL__ __FILE__ ":" stringify (__LINE__)
const char *__backtrace (const char *file, int lim);
#define __BACKTRACE__ (__backtrace (__FL__, -1))

#ifdef DMALLOC

#define CHECK_BOUNDS 1

#define DMALLOC_FUNC_CHECK
#ifdef HAVE_MEMORY_H
#include <memory.h>
#endif /* HAVE_MEMORY_H */
#include <dmalloc.h>

#undef memcpy
#undef xfree
#undef xstrdup

#if DMALLOC_VERSION_MAJOR < 5  || \
     (DMALLOC_VERSION_MAJOR == 5 && DMALLOC_VERSION_MINOR < 5)

#define memcpy(to, from, len) \
  _dmalloc_memcpy((char *) (to), (const char *) (from), len)
#define memmove(to, from, len) \
  _dmalloc_bcopy((const char *) (from), (char *) (to), len)
#define xstrdup(__s) ((char *) dmalloc_strdup(__FILE__, __LINE__, __s, 1))

#else /* version >= 5.5 */

#define xstrdup(__s) \
((char *) dmalloc_strndup(__FILE__, __LINE__, __s, (-1), 0))

#endif /* version <=> 5.5 */

/* Work around Dmalloc's misunderstanding of free's definition */
#if DMALLOC_VERSION_MAJOR >= 5
#define _xmalloc_leap(f, l, s) \
  dmalloc_malloc (f, l, s, DMALLOC_FUNC_MALLOC, 0, 1)
#define _malloc_leap(f, l, s) \
  dmalloc_malloc (f, l, s, DMALLOC_FUNC_MALLOC, 0, 0)
#define _xfree_leap(f, l, p) dmalloc_free (f, l, p, DMALLOC_FUNC_FREE)


#endif /* DMALLOC_VERSION_MAJOR >= 5 */

static inline void
_xfree_wrap (const char *file, int line, void *ptr)
{
  if (ptr)
    _xfree_leap(file, line, ptr);
}
static inline void
xfree (void *ptr)
{
  if (ptr)
    _xfree_leap("unknown file", 0, ptr);
}
#define xfree(ptr) _xfree_wrap(__FILE__, __LINE__, ptr)

const char *stktrace (const char *file);
extern int stktrace_record;
#define txmalloc(size) _xmalloc_leap (stktrace (__FILE__), __LINE__, size)

#else /* !DMALLOC */

void *xmalloc (size_t);
void *xrealloc (void *, size_t);
#ifdef PYMALLOC
# include <Python.h>
# define xfree PyMem_Free
#else /* !PYMALLOC (i.e., the default condition) */
# define xfree free
#endif /* PYMALLOC */
char *xstrdup (const char *s);
#define txmalloc(size) xmalloc (size)

#endif /* !DMALLOC */

#ifndef HAVE_STRCASECMP
#ifdef DMALLOC
/* These funcitons are only implemented on systems that actually have
 * strcasecmp. */
#undef strcasecmp
#undef strncasecmp
#endif /* DMALLOC */
int strcasecmp (const char *, const char *);
int strncasecmp (const char *, const char *, int);
#endif /* HAVE_STRCASECMP */


/*
 * Other random stuff
 */

/* Some versions of linux have a weird offsetof definition */
#define xoffsetof(type, member)  ((size_t)(&((type *)0)->member))

#ifdef __cplusplus
/* Egcs g++ without '-ansi' has some strange ideas about what NULL
 * should be. */
#undef NULL
#define NULL 0
#undef __null
#define __null 0
#endif /* __cplusplus */

#undef sun

extern void panic (const char *msg, ...)
     __attribute__ ((noreturn, format (printf, 1, 2)));


#define MALLOCRESV 16		/* Best to allocate 2^n - MALLOCRESV bytes */

#ifndef MAINTAINER
# define MAINTAINER 1
#elif !MAINTAINER
# undef MAINTAINER
#endif /* !MAINTAINER */

/*
 *  Setuid or setgid programs must contain
 *
 *     int suidprotect = 1;
 *
 *  in the file that contains main.  Otherwise, bad things may happen
 *  in global constructors, even before the first line of main gets to
 *  execute.  In C++, clearing dangerous environment variables in main
 *  is insecure, as it may already be too late.
 *
 *  The function suidsafe() returns 1 if setuidprotect was not
 *  initialized to 1 or the program was not setuid or the program was
 *  run by root.  Otherwise, it returns 0.
 *
 *  The function safegetenv() returns the output of getenv unless
 *  suidsafe() is 1, in which case it always returns NULL.
 */
#ifdef __cplusplus
extern "C"
#endif /* __cplusplus */
int suidprotect;		/* Set at compile time in suid programs */
#ifdef __cplusplus
extern "C"
#endif /* __cplusplus */
int execprotect;		/* Set if program will drop privileges */
int suidsafe (void);
int execsafe (void);
char *safegetenv (const char *);
#ifdef PUTENV_COPIES_ARGUMENT
#define xputenv putenv
#else /* !PUTENV_COPIES_ARGUMENT */
extern int xputenv (const char *s);
#endif /* !PUTENV_COPIES_ARGUMENT */

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_ASYNC_SYSCONF_H_ */
