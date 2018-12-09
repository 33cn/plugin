/*
\section{Interface of Failure Handlers}

This file contains the interfaces of various routines that should be
called in the case of failure.  The routines perform various tasks
such as reporting errors and killing the process.
*/

#ifndef _FAIL_H
#define _FAIL_H

#ifdef __cplusplus
extern "C" {
#endif

extern void fail(char const* format, ...);
    /*
     * requires - format is a printf style format string that matches
     *	          the remaining arguments.
     * effects  - print error message followed by a newline and kill
     *		  the program.
     */

extern void sysfail(char const* msg);
    /*
     * requires - last system call failed.
     * effects  - print error message followed by a newline and kill
     * 		  the program.
     */

extern void syswarn(char const* msg);
    /*
     * requires - last system call failed.
     * effects  - print error message followed by a newline.
     */

    /*
     * usage    - FAIL_CHECK(result, exception_name);
     * requires - result is a pointer type.
     *		  exception_name is a C++ identifier (not a string).
     * effects  - If result is a null pointer, call failure with
     *		  the appropriate arguments.  Else do nothing.
     */
#define FAIL_CHECK(status,type)						      \
	do {								      \
	    if (status == 0) {						      \
		fail("%s:%d: unhandled exceptions %s\n",		      \
		     __FILE__,						      \
		     __LINE__,						      \
		     #type);						      \
	    }								      \
        } while (0)

#ifdef __cplusplus
}
#endif

#endif /* _FAIL_H */
