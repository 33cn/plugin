

enum rtftp_status_t {
     RTFTP_OK = 0,
     RTFTP_NOENT = 1,
     RTFTP_CORRUPT = 2,
     RTFTP_EOF = 3,
     RTFTP_EEXISTS = 4,
     RTFTP_EFS = 5,
     RTFTP_BEGIN = 6,
     RTFTP_ERR = 7,
     RTFTP_OUT_OF_SEQ = 8,
     RTFTP_INCOMPLETE = 9
};

%#define RTFTP_HASHSZ 20
%#define CHUNKSZ 0x8000
%#define MAGIC 0xbeef4989

typedef opaque rtftp_hash_t[RTFTP_HASHSZ];

typedef string rtftp_id_t<>;
typedef opaque rtftp_data_t<>;
typedef hyper rtftp_xfer_id_t;

struct rtftp_file_t {
       unsigned magic;
       rtftp_id_t name;
       rtftp_hash_t hash;
       rtftp_data_t data;
};

struct rtftp_header_t {
       unsigned magic;
       rtftp_id_t name;
       rtftp_hash_t hash;
       unsigned size;
};

struct rtftp_xfer_header_t {
       rtftp_xfer_id_t xfer_id;
       unsigned size;
       rtftp_hash_t hash;
};

struct rtftp_chunkid_t {
       rtftp_xfer_id_t xfer_id;
       unsigned offset;
       unsigned size;
};

struct rtftp_chunk_t {
       rtftp_chunkid_t id;
       rtftp_data_t data;
};

struct rtftp_footer_t {
       rtftp_xfer_id_t xfer_id;
       unsigned size;
       unsigned n_chunks;
       rtftp_hash_t hash;
};

union rtftp_get_res_t switch (rtftp_status_t status) {
case RTFTP_OK:
       rtftp_file_t file;
default:
       void;
};

union rtftp_put2_res_t switch (rtftp_status_t status) {
case RTFTP_BEGIN:
     rtftp_xfer_id_t xfer_id;
default:
     void;
};

union rtftp_put2_arg_t switch (rtftp_status_t status) {
case RTFTP_BEGIN:
     rtftp_id_t name;
case RTFTP_OK:
     rtftp_chunk_t data;
case RTFTP_EOF:
     rtftp_footer_t footer;
default:
     void;
};

union rtftp_get2_arg_t switch (rtftp_status_t status) {
case RTFTP_BEGIN:
     rtftp_id_t name;
case RTFTP_OK:
     rtftp_chunkid_t chunk;
case RTFTP_EOF:
     rtftp_xfer_id_t id;
default:
     void;
};

union rtftp_get2_res_t switch (rtftp_status_t status) {
case RTFTP_BEGIN:
     rtftp_xfer_header_t header;
case RTFTP_OK:
     rtftp_chunk_t chunk;
default:
     void;

};

namespace RPC {

program RTFTP_PROGRAM { 
	version RTFTP_VERS {

	void
	RTFTP_NULL (void) = 0;

	rtftp_status_t
	RTFTP_CHECK(rtftp_id_t) = 1;

	rtftp_status_t
	RTFTP_PUT(rtftp_file_t) = 2;

	rtftp_get_res_t
	RTFTP_GET(rtftp_id_t) = 3;

	rtftp_put2_res_t
	RTFTP_PUT2(rtftp_put2_arg_t) = 4;

	rtftp_get2_res_t
	RTFTP_GET2(rtftp_get2_arg_t) = 5;

	} = 1;
} = 5401;	  

};

%#define RTFTP_TCP_PORT 5401
%#define RTFTP_UDP_PORT 5402

%#define MAX_PACKET_SIZE 0x8000000

