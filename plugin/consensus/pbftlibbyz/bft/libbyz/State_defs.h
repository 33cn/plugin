#ifndef _State_defs_h
#define _State_defs_h

// uncomment to use BASE instead of BFT
//#define BASE

#ifdef BASE
	#define OBJ_REP
#else
	#define NO_STATE_TRANSLATION
#endif

#endif
