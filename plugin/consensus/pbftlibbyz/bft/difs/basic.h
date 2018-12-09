/* \section{Common Types and Macros} */
#ifndef _BASIC_H
#define _BASIC_H

#include <assert.h>
/* The assert() macro */

#ifndef TRUE
#define TRUE 1
#endif

#ifndef FALSE
#define FALSE 0
#endif


#if 0
#ifndef __cplusplus
typedef unsigned long bool; /* I really hate this -- ACM */
#endif

#ifndef __GNUC__
typedef unsigned char bool;
#define false FALSE
#define true TRUE
#endif
#endif

#define Loop for (;;)
/* An infinite loop construct */

#define _BEGIN_ do {
#define _END_   } while (0)
/*
    The macros _BEGIN_ and _END_ are used when defining macros whose
    result is a statement. The definition should start with _BEGIN_ and
    end with _END_. In between these tokens may appear any sequence of
    valid C or C++ statements, as though _BEGIN_ and _END_ were braces.
*/

#endif /* _BASIC_H */
