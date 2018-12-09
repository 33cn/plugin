/* $Id: nfsconf.h,v 1.1 2002/05/01 18:27:29 rfink Exp $ */

#ifndef _NFSCONF_H_
#define _NFSCONF_H_ 1

/*
 * system-specific NFS mount stuff
 */

#ifdef __sun__
# define NFSCLIENT 1
#endif /* __SunOS__ */

#define export _export		/* C++ keyword gets used in C */

#ifdef __linux__
/* Linux has some pretty broken header files. */
# ifndef _BSD_SOURCE
#  define _BSD_SOURCE 1
# endif /* !_BSD_SOURCE */
#endif /* __linux__ */

#ifdef __FreeBSD__
# define NFS 1
#include <sys/time.h>
#include <nfs/rpcv2.h>
#endif /* __FreeBSD__ */

#include <sys/types.h>
#include <sys/param.h>
#include <sys/mount.h>
#if NEED_NFS_NFS_H
# include <nfs/nfs.h>
#endif /* NEED_NFS_NFS_H */
#if NEED_NFS_MOUNT_H
# include <nfs/mount.h>
#endif /* NEED_NFS_MOUNT_H */
#ifdef HAVE_SYS_MNTENT_H
# include <sys/mntent.h>
#endif /* HAVE_SYS_MNTENT_H */

#if __linux__
# ifndef __GLIBC__
#  include <linux/fs.h> /* patch up broken pieces */
#  undef LIST_HEAD      /* yes, it's *really* broken. */
# endif
# include <sys/socket.h>
# include <netinet/in.h>
# include <arpa/inet.h>
# include <rpc/rpc.h> /* linux libc5 is *broken*! */

# include <rpcsvc/nfs_prot.h> /*linux libc5 does not #ifdef bracket this code*/
# ifndef _NFS_PROT_H_RPCGEN
#  define _NFS_PROT_H_RPCGEN
# endif /* !_NFS_PROT_H_RPCGEN */
#define _LINUX_IN_H
# include <linux/nfs_mount.h> /* can't help but include this one. */
# define nfs_args nfs_mount_data
# define NFS_ARGSVERSION NFS_MOUNT_VERSION
#endif /* __linux__ */

#if defined(__ultrix)
/* This doesn't work yet! */
# include <sys/fs_types.h>
# include <nfs/nfs_gfs.h>
# define nfs_args nfs_gfs_mount
# define MOUNT_NFS GT_NFS
#endif /* __ultrix */

#ifdef HAVE_TIUSER_H
# include <tiuser.h>
#endif /* HAVE_TIUSER_H */

#undef export

#if defined (__OpenBSD__) || defined (__FreeBSD__) || defined (__NetBSD__)
/* Define this if you can mount a file system on the current working
 * directory. */
# define MOUNT_DOT 1
#endif /* systems where you can mount on "." */

#ifndef MOUNT_NFS
# ifdef MNTTYPE_NFS
#  define MOUNT_NFS MNTTYPE_NFS
# else /* !MNTTYPE_NFS */
#  define MOUNT_NFS "nfs"
# endif /* !MNTTYPE_NFS */
#endif /* !MOUNT_NFS */

#ifndef MOUNT_NFS3
# ifdef MNTTYPE_NFS3
#  define MOUNT_NFS3 MNTTYPE_NFS3
# else /* !MNTTYPE_NFS3 */
#  define MOUNT_NFS3 MOUNT_NFS
# endif /* !MNTTYPE_NFS3 */
#endif /* !MOUNT_NFS3 */

#ifndef MNT_NOSUID
# if defined (M_NOSUID)
#  define MNT_NOSUID M_NOSUID
# elif defined (MS_NOSUID)
#  define MNT_NOSUID MS_NOSUID
# else /* no MNT_NOSUID substitute found */
#  define NO_NOSUID 1
#  define MNT_NOSUID 0
# endif /* no MNT_NOSUID substitute found */
#endif /* !MNT_NOSUID */

#ifndef MNT_NODEV
# if defined (M_NODEV)
#  define MNT_NODEV M_NODEV
# elif defined (MS_NODEV)
#  define MNT_NODEV MS_NODEV
# else /* no nodev mount flag */
#  define MNT_NODEV 0
#  ifndef NFSMNT_NODEV
#   define NO_NODEVS 1
#  endif /* !NFSMNT_NODEV */
# endif /* no nodev mount flag */
#endif /* !MNT_NODEV */

#ifndef MNT_RDONLY
# if defined (M_RDONLY)
#  define MNT_RDONLY M_RDONLY
# elif defined (MS_RDONLY)
#  define MNT_RDONLY MS_RDONLY
# else /* no MNT_RDONLY substitute found */
#  define MNT_RDONLY 0
# endif /* no MNT_RDONLY substitute found */
#endif /* !MNT_RDONLY */

#ifndef MNT_UPDATE
# if defined (MS_REMOUNT)
#  define MNT_UPDATE MS_REMOUNT
# elif defined (M_UPDATE)
#  define MNT_UPDATE M_UPDATE
# endif /* M_UPDATE */
#endif /* !MNT_UPDATE */

#ifndef MNT_FORCE
# define MNT_FORCE 0
#endif /* !MNT_FORCE */

#ifdef M_RDONLY
/* Reaaly a "normal" mound syscall, don't let extraneous MS_DATA throw us. */
#undef MS_DATA
#endif /* M_RDONLY */

/* Some more strange Linuxisms */
#if defined(NFS_MOUNT_SOFT) && !defined(NFSMNT_SOFT) 
# define NFSMNT_SOFT NFS_MOUNT_SOFT
#endif /* NFS_MOUNT_SOFT && !NFSMNT_SOFT */
#if defined(NFS_MOUNT_INTR) && !defined(NFSMNT_INT) 
# define NFSMNT_INT NFS_MOUNT_INTR
#endif /* NFS_MOUNT_INTR && !NFSMNT_INT */
#if defined(NFS_MOUNT_NOAC) && !defined(NFSMNT_NOAC) 
# define NFSMNT_NOAC NFS_MOUNT_NOAC
#endif /* NFS_MOUNT_NOAC && !NFSMNT_NOAC */
#if defined(NFS_MOUNT_NOAC) && !defined(NFSMNT_NOAC) 
# define NFSMNT_NOAC NFS_MOUNT_NOAC
#endif /* NFS_MOUNT_NOAC && !NFSMNT_NOAC */
#if defined(NFS_MOUNT_TCP) && !defined(NFSMNT_TCP) 
# define NFSMNT_TCP NFS_MOUNT_TCP
#endif /* NFS_MOUNT_TCP && !NFSMNT_TCP */
#if defined(NFS_MOUNT_VER3) && !defined(NFSMNT_NFSV3) 
# define NFSMNT_NFSV3 NFS_MOUNT_VER3
#endif /* NFS_MOUNT_VER3 && !NFSMNT_NFSV3 */

#if defined(__linux__)
#define SYS_MOUNT(hostname, type, dir, mntflags, args) \
    mount (hostname, dir, type, MS_MGC_VAL | mntflags, args)

#elif defined(__ultrix)
/* This doesn't work yet! */
#define SYS_MOUNT(hostname, type, dir, mntflags, args)	\
    mount (hostname, dir, mntflags, type, args)

#elif defined(MS_DATA) /* SVR4 6 argument mount */
#define SYS_MOUNT(hostname, type, dir, mntflags, args)	\
     mount (hostname, dir, MS_DATA|mntflags,		\
	    type, args, sizeof (*(args)))

#elif M_NEWTYPE /* SunOS 4 */
#define SYS_MOUNT(hostname, type, dir, mntflags, args)	\
    mount (type, dir, M_NEWTYPE|mntflags, args)

#elif HAVE_VFSMOUNT /* HPUX9 */
#define SYS_MOUNT(hostname, type, dir, mntflags, args)	\
    vfsmount (type, dir, mntflags, args)

#else /* normal mount syscall */
#define SYS_MOUNT(hostname, type, dir, mntflags, args)	\
    mount (type, dir, mntflags, (char *) (args))

#endif /* normal mount syscall */

#define SYS_NFS_MOUNT(type, dir, mntflags, args)		\
    SYS_MOUNT ((args)->hostname, type, dir, mntflags, args)

#ifdef HAVE_UNMOUNT
# define __unmount unmount
#else /* !HAVE_UNMOUNT */
# define __unmount umount
#endif /* !HAVE_UNMOUNT */

#ifdef UNMOUNT_FLAGS
# define SYS_UNMOUNT(path, flags) __unmount (path, flags)
#else /* !UNMOUNT_FLAGS */
# define SYS_UNMOUNT(path, flags) __unmount (path)
#endif /* !UNMOUNT_FLAGS */

#endif /* _NFSCONF_H_ */
