
typedef string ex_str_t<255>;

struct ex_struct_t {
	ex_str_t 	s;
	unsigned 	u;
};

namespace RPC {
program EX_PROG {
	version EX_VERS {
		void
		EX_NULL(void) = 0;
	
		int
		EX_RANDOM(void) = 1;

		ex_str_t
		EX_REVERSE(ex_str_t) = 2;

		ex_struct_t
		EX_STRUCT(void) = 3;

		unsigned
		EX_RANDOM2 (unsigned) = 4;
	} = 1;
} = 31313;
};
