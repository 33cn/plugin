
#if defined (__OpenBSD__) || defined (__NetBSD__)
# ifndef USE_PCTR
#  define  USE_PCTR 1
# endif /* !USE_PCTR */
#endif /* __OpenBSD__ */

#if USE_PCTR
#include <machine/pctr.h>
#define get_time() rdtsc ()
#define TIME_LABEL "cycles"
#else /* !USE_PCTR */
inline u_int64_t
get_time ()
{
  timeval tv;
  gettimeofday (&tv, NULL);
  return (u_int64_t) tv.tv_sec * 1000000 + tv.tv_usec;
}
#define TIME_LABEL "usec"
#endif /* !USE_PCTR */


#define BENCH(iter, code)					\
{								\
  u_int64_t __v;						\
  { code; }							\
  { code; }							\
  __v = get_time ();						\
  for (u_int i = 0; i < iter; i++) {				\
    code;							\
  }								\
  __v = get_time () - __v;					\
  warn ("%s: %" U64F "d " TIME_LABEL " (%" U64F "d tot)\n",	\
        #code, __v / iter, __v);				\
}

#define TIME(code)					\
{							\
  u_int64_t __v;					\
  __v = get_time ();					\
  { code; }						\
  __v = get_time () - __v;				\
  warn ("%s: %" U64F "d " TIME_LABEL "\n", #code, __v);	\
}
