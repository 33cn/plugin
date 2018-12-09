/*
\section{Assertion Checking Routines}

These assertion checking routines provide more information than the
system "assert" routine.
*/

#ifndef _TH_ASSERT_H
#define _TH_ASSERT_H

#include <assert.h>

/*

\begin{verbatim}
th_assert(expr, msg)

    If compiled without -NDEBUG, check that "expr" evaluates to true. If not
    true, cause the program to dump core and print a message like:

	    Assertion failed: expr
	    msg, in file xxx.c, at line nnn
    
    If compiled with -DNDEBUG, do nothing.

    "msg" must be a string constant.

th_fail(msg)

    Always causes the program to dump core, and prints a message like

	    Assertion failed: FATAL ERROR
	    msg, in file xxx.c, at line nnn


    "expr" must be an numeric or pointer expression, and "msg" must be
    a string constant.

th_fail_str(buf)

    Always causes the program to dump core, and prints a message like

	    Assertion failed: str, in file xxx.c, at line nnn
   
    "str" may be a string constant or a variable of type "char[]" or "char *".
    This macro should be used only if "th_fail" cannot be.
\end{verbatim}

*/

#include "fail.h"

#ifdef __cplusplus
extern "C" {
#endif

#if defined(__osf__) || defined(ultrix) || defined(sun) || defined(__linux__)

#if defined(__STDC__) || defined(__cplusplus)

/* In some configurations __assert may be a macro so be careful here */

#ifndef __assert
#if defined (__linux__)

#if 1
//#define __assert(str, file, line) fail("%s %s %d", str, file, line)
#else
#define __assert(str, file, line) __assert_fail(str,file, line,__ASSERT_FUNCTION)
#endif

#else
extern void __assert(char *, char *, int);
#endif
#endif

#ifndef NDEBUG
#define th_assert(expr,msg) (void)(((expr) != 0) ? 0 : (__assert(#expr "\n" msg, __FILE__, __LINE__), 1))
#else
#define th_assert(expr,msg) ((void)0)
#endif /* NDEBUG */

#define th_fail(msg) __assert("FATAL ERROR\n" msg, __FILE__, __LINE__)
#define th_fail_str(buf) __assert(buf, __FILE__, __LINE__)

#else /* defined(__STDC__) || defined(__cplusplus) */

#ifndef NDEBUG
#define th_assert(expr,msg) if ((expr) != 0); else__assert(msg, __FILE__, __LINE__)
#else
#define th_assert(expr,msg)
#endif

/* In some configurations __assert may be a macro so be careful here */
#ifndef __assert
extern void __assert(char *, char *, int);
#endif

#define th_fail(msg) __assert(msg, __FILE__, __LINE__)
#define th_fail_str(buf) __assert(buf, __FILE__, __LINE__)

#endif /* defined(__STDC__) || defined(__cplusplus) */

#else

#error Unknown operating system, underlying assert call not known.

#endif /* ! (defined(__osf__) || defined(ultrix) || defined(sun)) */

#ifdef __cplusplus
}
#endif

#endif /* TH_ASSERT_H */
