/* $Id: pmap_prot.x 1693 2006-04-28 23:17:35Z max $ */

/* This is from RFC 1057 */

const PMAP_PORT = 111;      /* portmapper port number */

struct mapping {
   unsigned int prog;
   unsigned int vers;
   unsigned int prot;
   unsigned int port;
};

struct pmaplist {
   mapping map;
   pmaplist *next;
};
typedef pmaplist *pmaplist_ptr;

struct call_args {
   unsigned int prog;
   unsigned int vers;
   unsigned int proc;
   opaque args<>;
};

struct call_result {
   unsigned int port;
   opaque res<>;
};

program PMAP_PROG {
   version PMAP_VERS {
      void
      PMAPPROC_NULL(void)         = 0;

      bool
      PMAPPROC_SET(mapping)       = 1;

      bool
      PMAPPROC_UNSET(mapping)     = 2;

      unsigned int
      PMAPPROC_GETPORT(mapping)   = 3;

      pmaplist_ptr
      PMAPPROC_DUMP(void)         = 4;

      call_result
      PMAPPROC_CALLIT(call_args)  = 5;
   } = 2;
} = 100000;
