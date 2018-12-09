/* umac.h */
#ifndef _UMAC_H
#define _UMAC_H

#ifdef __cplusplus
    extern "C" {
#endif

typedef struct umac_ctx *umac_ctx_t;

umac_ctx_t umac_new(char key[]);
/* Dynamically allocate a umac_ctx struct, initialize variables, 
 * generate subkeys from key.
 */

int umac_reset(umac_ctx_t ctx);
/* Reset a umac_ctx to begin authenicating a new message */

int umac_update(umac_ctx_t ctx, char *input, long len);
/* Incorporate len bytes pointed to by input into context ctx */

int umac_final(umac_ctx_t ctx, char tag[], char nonce[8]);
/* Incorporate any pending data and the ctr value, and return tag. 
 * This function returns error code if ctr < 0. 
 */

int umac_delete(umac_ctx_t ctx);
/* Deallocate the context structure */

int umac(umac_ctx_t ctx, char *input, 
         long len, char tag[],
         char nonce[8]);
/* All-in-one implementation of the functions Reset, Update and Final */


/* uhash.h */


typedef struct uhash_ctx *uhash_ctx_t;
  /* The uhash_ctx structure is defined by the implementation of the    */
  /* UHASH functions.                                                   */
 
uhash_ctx_t uhash_alloc(char key[16]);
  /* Dynamically allocate a uhash_ctx struct and generate subkeys using */
  /* the kdf and kdf_key passed in. If kdf_key_len is 0 then RC6 is     */
  /* used to generate key with a fixed key. If kdf_key_len > 0 but kdf  */
  /* is NULL then the first 16 bytes pointed at by kdf_key is used as a */
  /* key for an RC6 based KDF.                                          */
  
int uhash_free(uhash_ctx_t ctx);

int uhash_set_params(uhash_ctx_t ctx,
                   void       *params);

int uhash_reset(uhash_ctx_t ctx);

int uhash_update(uhash_ctx_t ctx,
               char       *input,
               long        len);

int uhash_final(uhash_ctx_t ctx,
              char        ouput[]);

int uhash(uhash_ctx_t ctx,
        char       *input,
        long        len,
        char        output[]);

#ifdef __cplusplus
    }
#endif 

#endif
