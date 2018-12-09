// -*-c++-*-
/* $Id: sfscrypt.h 1754 2006-05-19 20:59:19Z max $ */

/*
 *
 * Copyright (C) 2002 Max Krohn <max@cs.nyu.edu>
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation; either version 2, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307
 * USA
 */

/* SFSCRYPT
 *
 * This library serves as an abstract interface between specific
 * cryptographic implementations and the generic XDR implementations
 * of keys, ciphertexts, and signatures.  In general, keys are
 * stored in files, given to SFS programs via RPC calls, or generated
 * by SFS. In all cases, the key should be expressed as ptr<sfspub>
 * or an ptr<sfspriv> object before using that key for encryption/
 * decryption or signing/vefication.  This library serves to 
 * facilitate (1) the importing of keys from file and 
 * XDR sturctures (2) the exporting of keys back to XDR or files 
 * (3) using keys to perform common cryptographic tasks and (4) generating
 * new keys.
 *
 * Note that sfspriv is a subclass of an sfspub. All 4 cryptographic
 * operations (sign, verify, decrypt, encrypt) can work on an sfspriv
 * object, but only verify and encrypt can be called on an sfspub.
 * Methods such as export_pubkey (), operator==, get_pubkey_hash ()
 * are defined on sfspub objects, and hence will work on sfspriv objects.
 *
 * Key Generation/Allocation 
 * --------------------------
 *  How to get an ptr<sfspriv> or ptr<sfspub>: In general, call
 *  alloc () or gen () on the global sfscrypt object.  In 
 *  particular:
 * 
 *    Generation
 *    ----------
 *
 *        ptr<sfsriv> k = sfscrypt.gen (SFS_RABIN, 0, SFS_SIGN | SFS_DECRYPT);
 *
 *    To generate a key, call the gen () function on the global object
 *    sfscrypt. The first argument is the type of key that is needed,
 *    which can at current be SFS_RABIN or 2SFS_SCHNORR. The second
 *    argument is a keysize; the value of this parameter cleary has
 *    different meanings based on the requested cryptosystem. Providing
 *    0 gives system-wide defaults as given in sfs_config. The third
 *    argument indicates what the key is used for; it can be a bitwise
 *    OR of 0 or more of the following 4 modes: SFS_SIGN, SFS_VERIFY,
 *    SFS_ENCRYPT and SFS_DECRYPT.  Using a key in a mode it wasn't
 *    allocated for will cause an assertion failure.
 *
 *    Allocation/Importing
 *    --------------------
 *    How to read a key from an XDR structure or a file and use it for
 *    signing, verifying, encrypting or decrypting:
 *
 *        // import a public key from an XDR structure for verifying
 *        // signatures.
 *        sfs_encryptarg2 *a = sbp->template getarg<sfs_encryptarg2> ();
 *        ptr<sfspub> p = sfscrypt.alloc (a->pubkey, SFS_VERIFY);
 *
 *        // Read an unencrypted private key from a file, for signing.
 *        str rawkey = file2str ("/tmp/srvkey");
 *        ptr<sfspriv> s = sfscrypt.alloc_priv (rawkey, SFS_SIGN);
 *
 *        // Read a possibly encrypted private key, prompting 
 *        // for password if it is encrypted.
 *        str rawkey2 = file2str ("/tmp/myprivkey");
 *        str keyname, pwd;
 *        u_int cost; 
 *        ptr<sfspriv> s2 = sfscrypt.alloc (rawkey2, &keyname, &pwd,
 *                                          &cost, SFS_SIGN);
 *       
 *    These are examples of the most commonly used allocators. In general,
 *    it's possible to load a key from XDR structure or file into a 
 *    ptr<sfspub>/ptr<sfspriv> object, for subsequent cryptographic use.
 *    As above, it's necessary to specify ahead of time what the key
 *    is going to be used for; using a key for undeclared purposes will
 *    again cause an assertion failure. Other useful ::alloc methods
 *    are defined in sfscrypt_t, and should be called on the global
 *    sfscrypt object.
 *
 * 
 * Exporting
 * -----------
 *  Once you have a ptr<sfspub>/ptr<sfspriv> object, you can write it 
 *  back out to XDR or to a string.
 *
 *      str s;
 *      sfs_encryptarg2 a;
 *      ptr<sfspriv> p = sfscrypt.alloc_priv (rawkey, SFS_SIGN);
 *
 *      // export to XDR to establish a session key in SFS
 *      bool b1 = p->export_pubkey (a.pubkey);
 *
 *      // export to a file
 *      bool b2 = p->export_privkey (&s, "my-key", "fv94++1ee", 10);
 *      str2file ("/tmp/privkey", s);
 *      
 *  These are common usages of the export functions. They return
 *  true on success, and false on failure. The given export_privkey ()
 *  takes as arguments a string to which to output the key, a keyname,
 *  a password, and password-guessing cost. The key will be encrypted using
 *  the given password, and will be output to base-64 encoded string that 
 *  can subsequently be written to a file. To export a private key to a file
 *  in the clear, use the same function but pass NULL as a password 
 *  and 0 for cost. 
 *
 *
 * Signing/Verifying Decrypting/Encrypting
 * ---------------------------------------
 *  Here are the important method signatures for cryptographic 
 *  primitives. All return true on success and false on failure.
 *  Some include optional last arguments for error reporting.
 *
 *    bool verify (const sfs_sig2 &sig, const str &msg, str *err = NULL);
 *    bool encrypt (sfs_ctext2 *ct, const str &msg);
 *    bool decrypt (const sfs_ctext2 &ct, str *msg, u_int ptsz);
 *    bool sign (sfs_sig2 *sig, const str &msg);
 *
 *  Note that some signature schemes such as 2-Schnorr make network
 *  calls and hence block.  Calling the above synchronous sign () in
 *  the context of the 2-Schnorr cryptosystem will trigger a call to acheck (),
 *  which, if called from within amain (), will present problems.  The 
 *  asynchronous version of sign will work for all signature systems,
 *  including Rabin:
 *  
 *    void sign (const sfsauth2_sigreq &sr, sfs_authinfo ainfo, cbsign cb);
 * 
 *  cb is called upon completion or error:
 * 
 *    typedef callback<void, str, ptr<sfs_sig2> >::ref cbsign;
 *
 *  The first argument is an error message (or NULL if success) and the
 *  second is the desired signature.
 *
 * Miscellaneous
 * -------------
 *  sfspub/sfspriv objects can help to facilitate key equality checks, and 
 *  also in-memory storage of keys through key hashes.  The relevant functions
 *  are as follows:
 * 
 *  class sfspub {
 *    bool operator== (const str &s) const;
 *    bool operator== (const sfspub &p) const;
 *    bool operator== (const sfs_pubkey2 &pk) const;
 *    bool get_pubkey_hash (sfs_hash *h) const;
 *    str get_pubkey_hash () const;
 *  }
 *
 *  Note that the corresponding "!=" operators are not currently defined.
 *  The get_pubkey_hash () function will return false or NULL on failure,
 *  and true or a 20-byte hash on success.
 *
 */

#ifndef _SFSMISC_SFSCRYPT_H
#define _SFSMISC_SFSCRYPT_H 1

#include "crypt.h"
#include "rabin.h"
#include "serial.h"
#include "sfs_prot.h"
#include "sfsauth_prot.h"
#include "sfsagent.h"
#include "sfsconnect.h"
#include "sfsmisc.h"
#include "esign.h"

#define SFS_ENCRYPT   (1 << 0)
#define SFS_VERIFY    (1 << 1)
#define SFS_DECRYPT   (1 << 2)
#define SFS_SIGN      (1 << 3)


class sfspub {
public:
  sfspub (sfs_keytype kt, u_char o = 0, const str &k = NULL) 
    : keylabel (k), ktype (kt), opts (o) {}
  virtual ~sfspub () {}
  sfs_keytype get_type () const { return ktype; }

  // Should be implemented by child classes
  virtual bool export_pubkey (sfs_pubkey2 *k) const = 0;
  virtual bool export_pubkey (strbuf &sb, bool prefix = true) const = 0;
  virtual bool check_keysize (str *s = NULL) const = 0;
  virtual bool encrypt (sfs_ctext2 *ct, const str &msg) const { return false; }
  virtual bool verify (const sfs_sig2 &sig, const str &msg, str *e = NULL) 
    const { return false; }

  // These have reasonable default behavior
  virtual bool operator== (const str &s) const;
  virtual bool operator== (const sfspub &p) const;
  virtual bool operator== (const sfs_pubkey2 &pk) const;
  virtual bool operator== (const sfsauth_keyhalf &kh) const { return false; }
  virtual void set_username (const str &s) {} 
  virtual void set_hostname (const str &s) {}
  virtual str  get_hostname () const { return NULL; }
  virtual bool is_proac () const { return false; }
  
  // Legacy Code -- To be Phased out
  virtual bool is_v2 () const { return true; }
  virtual bool export_pubkey (sfs_pubkey *k) const { return false; }
  virtual bool encrypt (sfs_ctext *ct, const str &msg) const { return false; }
  virtual bool verify_r (const bigint &n, size_t len, str &msg, str *e = NULL) 
    const { return false; }

  static bool check_keysize (size_t nbits, size_t ll, size_t ul,
			     const str &t, str *s = NULL);
  bool check_opts () const { return (!get_opt (get_bad_opts ())); }
  bool get_pubkey_hash (sfs_hash *h, int vers = 2) const ;
  str get_pubkey_hash (int vers = 2) const;

  const str keylabel;
protected:
  bool verify_init (const sfs_sig2 &sig, const str &msg, str *e) const;
  virtual u_char get_bad_opts () const { return (SFS_DECRYPT | SFS_SIGN); }
  bool get_opt (u_char o) const { return (opts & o); }
  const sfs_keytype ktype;
  const u_char opts;
};

class sfspriv : virtual public sfspub {
public:
  sfspriv (sfs_keytype kt, u_char o = 0, const str &k = NULL) 
    : sfspub (kt, o, k) {}
  virtual ~sfspriv () {}

  // The Following 3 functions should be implemented by child classes
  virtual bool decrypt (const sfs_ctext2 &ct, str *msg, u_int sz) const = 0;
  virtual bool sign (sfs_sig2 *sig, const str &msg) = 0;
protected: 
  virtual bool export_privkey (str *s) const = 0;

  // Asynchronous version ; reasonable default is given
public:
  virtual void sign (const sfsauth2_sigreq &sr, sfs_authinfo ainfo, 
		     cbsign cb);
  
  // Legacy Code -- to be phased out
  virtual bool decrypt (const sfs_ctext &ct, str *msg) const { return false; }
  virtual bool sign_r (sfs_sig *sig, const str &msg) const { return false; }

  // Implemented in terms of export_privkey (str *s)
  virtual bool get_privkey_hash (u_int8_t *buf, const sfs_hash &hostid) const;
  virtual bool export_privkey (str *s, const str &kn, str pwd, u_int cost) 
    const;
  virtual bool export_privkey (str *s, const eksblowfish *eksb) const;

  bool export_privkey (str *s, const str &salt, const str &ske,
		       const str &kn) const;
  virtual bool export_keyhalf (sfsauth_keyhalf *, bool *) const ;

  virtual bool export_privkey (sfs_privkey2_clear *k) const { return false; }
  virtual str get_desc () const { return ("generic private key"); }

  // Initialize a key by signing a null request
  void init (cbs cb);
  void initcb (cbs cb, str err, ptr<sfs_sig2> sig);

  // Functions needed for abstraction of multi-party signatures
  virtual ptr<sfspriv> regen () const { return NULL; }
  virtual ptr<sfspriv> update () const { return NULL; }
  virtual ptr<sfspriv> wholekey () const { return NULL; }
  virtual bool get_coninfo (ptr<sfscon> *c, str *h) const { return false; }

protected:
  u_char get_bad_modes () { return 0; }
  void signcb (str *errp, sfs_sig2 *sigp, str err, ptr<sfs_sig2> sig);
};

class sfsca { // SFS Crypt Allocator
public:
  sfsca () : ktype (SFS_NOKEY) {}
  sfsca (sfs_keytype x, const str &s) : ktype (x), skey (s) {} 

  // These 6 functions should be implemented by child classes
  virtual ptr<sfspub>  alloc (const sfs_pubkey2 &pk, u_char o = 0) const = 0;
  virtual ptr<sfspub>  alloc (const str &s, u_char o = 0) const = 0;
  virtual ptr<sfspriv> alloc (const sfs_privkey2_clear &pk, u_char o = 0) 
    const = 0;
  virtual ptr<sfspriv> alloc (const str &raw, ptr<sfscon> s,
			      u_char o = 0) const = 0;
  virtual ptr<sfspriv> gen (u_int nbits, u_char o = 0) const { return NULL; }
  virtual const sfs_keytype *get_private_keytypes (u_int *n) const 
  { return NULL; }

  // Legacy Code
  virtual ptr<sfspub> alloc (const sfs_pubkey &pk, u_char o = 0) const 
  { return NULL; }

  virtual ~sfsca () {}
  
  ihash_entry<sfsca> hlinks;
  ihash_entry<sfsca> hlinkx;
  sfs_keytype ktype;
  str skey;
};

#define ESIGN_SCALE(x) ((3 * (x)) >> 1) /* XXX: hack for now */
class sfs_esign_pub : virtual public sfspub {
public:
  sfs_esign_pub (ref<esign_pub> k, u_char opts = 0)
    : sfspub (SFS_ESIGN, opts, "esign"), pubk (k) {}
  bool verify (const sfs_sig2 &sig, const str &msg, str *e = NULL) const;
  bool export_pubkey (sfs_pubkey2 *k) const;
  bool export_pubkey (strbuf &b, bool prefix = true) const;
  bool check_keysize (str *s = NULL) const;
  static bool check_keysize (size_t nbits, str *s);

private:
  ref<esign_pub> pubk;
};

class sfs_esign_priv : public sfs_esign_pub, public sfspriv {
public:
  sfs_esign_priv (ref<esign_priv> k, u_char opts = 0)
    : sfspub (SFS_ESIGN, opts, "esign"), sfs_esign_pub (k),
      sfspriv (SFS_ESIGN, opts, "esign"), privk (k) {}

  bool sign (sfs_sig2 *s, const str &msg) ;
  bool export_privkey (sfs_privkey2_clear *k) const;
  bool decrypt (const sfs_ctext2 &ct, str *msg, u_int sz) const 
  { return false; }

  str get_desc () const { strbuf b; export_pubkey (b); return b;}

protected:
  bool export_privkey (str *s) const;

private:
  bool export_privkey (sfs_esign_priv_xdr *k) const;
  ref<esign_priv> privk;
};

class sfs_rabin_pub : virtual public sfspub {
public:
  sfs_rabin_pub (ref<rabin_pub> k, u_char o = 0) 
    : sfspub (SFS_RABIN, o, "rabin"), pubk (k) {}
  bool verify (const sfs_sig2 &sig, const str &msg, str *e = NULL) const;
  bool export_pubkey (sfs_pubkey2 *k) const;
  bool export_pubkey (strbuf &b, bool prefix = true) const;
  bool check_keysize (str *s = NULL) const;
  bool encrypt (sfs_ctext2 *ct, const str &msg) const;

  static bool check_keysize (size_t nbits, str *s = NULL);

  // Legacy Code
  bool verify_r (const bigint &n, size_t len, str &msg, str *e = NULL) const;
  bool export_pubkey (sfs_pubkey *k) const;
  bool encrypt (sfs_ctext *ct, const str &msg) const;
  bool is_v2 () const { return false; }
private:
  ref<rabin_pub> pubk;
};

class sfs_rabin_priv : public sfs_rabin_pub, public sfspriv {
public:
  sfs_rabin_priv (ref<rabin_priv> k, u_char o = 0)
    : sfspub (SFS_RABIN, o, "rabin"), sfs_rabin_pub (k), 
      sfspriv (SFS_RABIN, o, "rabin"), privk (k) {}

  bool sign (sfs_sig2 *s, const str &msg);
  bool export_privkey (sfs_privkey2_clear *k) const;
  bool decrypt (const sfs_ctext2 &ct, str *msg, u_int sz) const;
  bool get_privkey_hash (u_int8_t *buf, const sfs_hash &hostid) const;

  // Legacy Code
  bool sign_r (sfs_sig *sig, const str &msg) const;
  bool decrypt (const sfs_ctext &ct, str *msg) const;

  str get_desc () const { strbuf b; export_pubkey (b); return b;}

protected:
  bool export_privkey (str *s) const;

private:
  ref<rabin_priv> privk;
};

class sfs_rabin_alloc : public sfsca {
public:
  sfs_rabin_alloc (sfs_keytype kt = SFS_RABIN, const str &s = "rabin") 
    : sfsca (kt, s) {}

  ptr<sfspriv> alloc (const sfs_privkey2_clear &pk, u_char o = 0) const;
  ptr<sfspub>  alloc (const sfs_pubkey2 &pk, u_char o = 0) const;
  ptr<sfspub>  alloc (const str &s, u_char o = 0) const;
  ptr<sfspriv> gen (u_int nbits, u_char o = 0) const;
  ptr<sfspriv> alloc (const str &raw, ptr<sfscon> c, u_char o = 0) const;

  // Old protocol
  ptr<sfspub> alloc (const sfs_pubkey &pk, u_char o = 0) const;
};

class sfs_esign_alloc : public sfsca {
public:
  sfs_esign_alloc (sfs_keytype kt = SFS_ESIGN, const str &s = "esign")
    : sfsca (kt, s) {}

  ptr<sfspriv> alloc (const sfs_privkey2_clear &pk, u_char o = 0) const;
  ptr<sfspub>  alloc (const sfs_pubkey2 &pk, u_char o = 0) const;
  ptr<sfspub>  alloc (const str &s, u_char o = 0) const;
  ptr<sfspriv> gen (u_int nbits, u_char o = 0) const;
  ptr<sfspriv> alloc (const str &raw, ptr<sfscon> c, u_char o = 0) const;
private:
  ptr<sfspriv> alloc (const sfs_esign_priv_xdr &k, u_char o = 0) const;
};

class sfscrypt_t {
public:
  sfscrypt_t ();

  /*
   * the argument 'u_char o' is for "options", an OR of 0 or more of:
   *
   *    { SFS_SIGN, SFS_VERIFY, SFS_ENCRYPT, SFS_DECRYPT }
   *
   */
  ptr<sfspriv> alloc (const sfs_privkey2_clear &pk, u_char o = 0) const;
  ptr<sfspriv> alloc (const str &ske, str *kn, str *pwd, u_int *cost, 
		      u_char o = 0) const;
  ptr<sfspub>  alloc (const sfs_pubkey2 &pk, u_char o = 0) const;
  ptr<sfspub>  alloc (const str &s, u_char o = 0) const;
  ptr<sfspriv> alloc (sfs_keytype xt, const str &esk, const eksblowfish *eksb, 
		      ptr<sfscon> s = NULL, u_char o = 0) const;
  ptr<sfspriv> gen (sfs_keytype xt, u_int nbits, u_char o = 0) const;
  ptr<sfspriv> alloc_priv (const str &sk, u_char o = 0) const;
  ptr<sfspub>  alloc_from_priv (const str &sk) const;

  // Legacy Code
  ptr<sfspub> alloc (const sfs_pubkey &pk, u_char o = 0) const;

  bool verify (const sfs_pubkey2 &pk, const sfs_sig2 &sig, const str &msg,
	       str *e = NULL) const;

  void add (sfsca *c)
  {
    if (c->skey.len ()) { strtab.insert (c); }
    if (c->ktype && !xttab[c->ktype]) { xttab.insert (c); }
  }
private:
  ptr<sfspub>  alloc (sfs_keytype xt, const str &pk, u_char o = 0) const;
  bool parse (const str &raw, sfs_keytype *xt, str *salt, str *ske, str *pk, 
	      str *kn) const;
  bool verify_sk (ref<sfspriv> k, sfs_keytype xt, const str &pk) const ;

  ihash<str, sfsca, &sfsca::skey, &sfsca::hlinks> strtab; 
  ihash<sfs_keytype, sfsca, &sfsca::ktype, &sfsca::hlinkx> xttab; 
};


#define SALTBITS SK_RABIN_SALTBITS

extern sfscrypt_t sfscrypt;

str sk_decrypt (const str &epk, const eksblowfish *eksb);
str sk_encrypt (const str &pk, const eksblowfish *eksb);
str timestr ();

template<class T> str
xdr2str_pad (const T &t, bool scrub = false, u_int align = 8)
{
  xdrsuio x (XDR_ENCODE, scrub);
  XDR *xp = &x;
  if (!rpc_traverse (xp, const_cast<T &> (t)))
    return NULL;
  while (x.uio ()->resid () & (align >> 1)) 
    if (!xdr_putint (&x, 0))
      return NULL;
  mstr m (x.uio ()->resid ());
  x.uio ()->copyout (m);
  if (scrub)
    return str2wstr (m);
  return m;
}


#endif /* _SFSMISC_SFSCRYPT_H */
