# The RPM will fail to build with some versions of gcc 2.96, which is
# default in the RH7.x distributions.  If "rpmbuild -ba sfs-0.6.spec"
# fails and you have gcc 2.96, install the following packages from the
# RH7.2 updates:
#   gcc3-3.0.4-1.i386.rpm      libstdc++3-3.0.4-1.i386.rpm
#   gcc3-c++-3.0.4-1.i386.rpm  libstdc++3-devel-3.0.4-1.i386.rpm
#   libgcc-3.0.4-1.i386.rpm
# and then run "rpmbuild -ba sfs-0.6.spec --with gcc3".  No gcc3
# packages are available in RH7.3, but the RH7.2 updates appear to work
# on RH7.3. 

Summary: SFS -- a secure, decentralized, global file system
Name: sfs
Version: 0.6
Release: 5
Copyright: GPL 
Group: Applications/File
Source: http://www.fs.net/download/sfs-%{version}.tar.gz 
URL: http://www.fs.net/
Packager: SFS Developers <http://www.fs.net/>
BuildRoot: %{_tmppath}/%{name}-%{version}-buildroot
Requires: kernel >= 2.2.14, gmp >= 2.0
Obsoletes: sfs-devel

%if %{?_with_gcc3:1}%{!?_with_gcc3:0}
BuildRequires: gcc3, gcc3-c++
%else
BuildRequires: gcc, gcc-c++
BuildConflicts: gcc = 2.96-85, gcc-c++ = 2.96-85, gcc = 2.96-98, gcc-c++ = 2.96-98,gcc = 2.96-108, gcc-c++ = 2.96-108, gcc = 2.96-110, gcc-c++ = 2.96-110
%endif


%package servers
Summary: SFS servers.
Group: Applications/File
PreReq: sfs = %{version}-%{release}
Requires: nfs-server


%description
The Self-Certifying File System (SFS) is a secure, global file system
with completely decentralized control. SFS lets you access your files
from anywhere and share them with anyone, anywhere. Anyone can set up
an SFS server, and any user can access any server from any client. SFS
lets you share files across administrative realms without involving
administrators or certification authorities.

This file includes the core files necessary for SFS clients.  Also
included are libraries and header files useful for development of
SFS-enabled tools.


%description servers
The Self-Certifying File System (SFS).

This package includes servers of SFS services such as the read-write
file server, the remote execution (REX) server, and the user
authentication server.


%prep
%setup -q

%build
if test -z "${DEBUG+set}"; then
    DEBUG=-O2
    export DEBUG
fi

%if %{?_with_gcc3:1}%{!?_with_gcc3:0}
env CC=gcc3 CXX=g++3 ./configure --prefix=/usr
%else
env CC=gcc CXX=g++ ./configure --prefix=/usr
%endif
make 

%install
rm -rf $RPM_BUILD_ROOT
make install-strip DESTDIR=$RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/var/sfs $RPM_BUILD_ROOT/etc/sfs $RPM_BUILD_ROOT/sfs
install -D -m 755 etc/sfscd.init $RPM_BUILD_ROOT/etc/rc.d/init.d/sfscd
install -D -m 755 etc/sfssd.init $RPM_BUILD_ROOT/etc/rc.d/init.d/sfssd

%clean
rm -rf $RPM_BUILD_ROOT

%pre
/bin/grep -q '^sfs:' /etc/passwd \
   || /usr/sbin/useradd -M -r -d /var/sfs -s /bin/true -c "SFS" sfs \
   || :
/bin/grep -q '^sfs:' /etc/group \
   || /usr/sbin/groupadd sfs 


%post
if [ $1 = 1 ]; then
  /sbin/chkconfig --add sfscd
#  /bin/chown sfs.sfs /var/sfs
# We cannot include /sfs in the %files list because of mount points
  /bin/mkdir -p /sfs
  /bin/chown sfs.sfs /sfs
fi
#/bin/chown root.sfs /usr/lib/sfs/suidconnect
#/bin/chmod g+s /usr/lib/sfs/suidconnect
#/bin/chmod u+s /usr/lib/sfs/newaid
/sbin/install-info --info-dir=/usr/info /usr/info/sfs.info || :

%post servers
if [ $1 = 1 ]; then
  /sbin/chkconfig --add sfssd
fi


%preun
if [ "$1" = "0" ]; then
   /sbin/service sfscd stop > /dev/null 2>&1 || :
   /sbin/chkconfig --del sfscd
   /sbin/install-info --info-dir=/usr/info --remove /usr/info/sfs.info || :
   /bin/rm -rf /var/sfs/random_seed /var/sfs/sockets
   /bin/rmdir /sfs || /bin/echo "Unable to remove /sfs -- remove manually when all file systems unmounted."  
fi

%preun servers
if [ "$1" = "0" ]; then
   /sbin/service sfssd stop > /dev/null 2>&1 || :
   /sbin/chkconfig --del sfssd
   /bin/rm -rf /var/sfs/authdb
fi


%postun
if [ "$1" = "0" ]; then
   if /bin/grep -q '^sfs:.*:/var/sfs:' /etc/passwd; then
      /usr/sbin/userdel sfs
      if /bin/grep -q \^sfs: /etc/group; then
         /usr/sbin/groupdel sfs
      fi
   fi
fi

if [ $1 -ge 1 ]; then
# race condition, no condrestart in SFS 0.6
  if /sbin/service sfscd status >/dev/null 2>&1; then
    /sbin/service sfscd restart >/dev/null 2>&1 || :

    /bin/sleep 3
    if ! /sbin/service sfscd status >/dev/null 2>&1; then
      /bin/echo "Restart sfscd manually after unmounting all remote SFS file systems."
    fi
  fi
fi

if [ $1 -ge 1 ]; then
# race condition, no condrestart in SFS 0.6
  if /sbin/service sfssd status >/dev/null 2>&1; then
    /sbin/service sfssd restart >/dev/null 2>&1 || :
  fi
fi


%files
%defattr(-,root,root)
%attr(0755,sfs,sfs) %dir /var/sfs
%attr(2555,root,sfs) /usr/lib/sfs-%{version}/suidconnect
%attr(4555,root,root) /usr/lib/sfs-%{version}/newaid
%config /etc/rc.d/init.d/sfscd
%config /usr/share/sfs/sfscd_config
%config /usr/share/sfs/sfs_config
%config /usr/share/sfs/sfs_srp_parms
%dir /etc/sfs
%dir /usr/lib/sfs-%{version}
%dir /usr/share/sfs
%doc AUTHORS COPYING ChangeLog NEWS README STANDARDS TODO
%ghost %attr(0755,sfs,sfs) %dir /sfs
/usr/bin/*
/usr/include/sfs
/usr/include/sfs.h
/usr/include/sfs-%{version}
/usr/info/sfs.info-1.gz
/usr/info/sfs.info-2.gz
/usr/info/sfs.info.gz
/usr/lib/libsfs.a
/usr/lib/sfs
/usr/lib/sfs-%{version}/aiod
/usr/lib/sfs-%{version}/connect
/usr/lib/sfs-%{version}/libarpc.a
/usr/lib/sfs-%{version}/libarpc.la
/usr/lib/sfs-%{version}/libasync.a
/usr/lib/sfs-%{version}/libasync.la
/usr/lib/sfs-%{version}/libsfscrypt.a
/usr/lib/sfs-%{version}/libsfscrypt.la
/usr/lib/sfs-%{version}/libsfsmisc.a
/usr/lib/sfs-%{version}/libsfsmisc.la
/usr/lib/sfs-%{version}/libsvc.a
/usr/lib/sfs-%{version}/libsvc.la
/usr/lib/sfs-%{version}/listen
/usr/lib/sfs-%{version}/mallock.o
/usr/lib/sfs-%{version}/moduled
/usr/lib/sfs-%{version}/nfsmounter
/usr/lib/sfs-%{version}/pathinfo
/usr/lib/sfs-%{version}/sfsrwcd
/usr/lib/sfs-%{version}/xfer
/usr/man/man1/rex.1.gz
/usr/man/man1/sfsagent.1.gz
/usr/man/man1/sfskey.1.gz
/usr/man/man1/ssu.1.gz
/usr/man/man5/sfscd_config.5.gz
/usr/man/man5/sfs_config.5.gz
/usr/man/man5/sfs_srp_params.5.gz
/usr/man/man8/sfscd.8.gz
/usr/sbin/funmount
/usr/sbin/sfscd


%files servers
%defattr(-,root,root)
%config /etc/rc.d/init.d/sfssd
%config /usr/share/sfs/sfsauthd_config
%config /usr/share/sfs/sfssd_config
# fake dir to clean up if one upgrades sfs before sfs-servers
%dir /usr/lib/sfs-%{version}
/usr/lib/sfs-%{version}/proxy
/usr/lib/sfs-%{version}/ptyd
/usr/lib/sfs-%{version}/rexd
/usr/lib/sfs-%{version}/sfsauthd
/usr/lib/sfs-%{version}/sfsrwsd
/usr/lib/sfs-%{version}/ttyd
/usr/man/man5/sfsauthd_config.5.gz
/usr/man/man5/sfsrwsd_config.5.gz
/usr/man/man5/sfssd_config.5.gz
/usr/man/man5/sfs_users.5.gz
/usr/man/man8/sfsauthd.8.gz
/usr/man/man8/sfsrwsd.8.gz
/usr/man/man8/sfssd.8.gz
/usr/sbin/sfssd


%changelog
* Tue Dec 03 2002 Kevin Fu <fubob@mit.edu>
- Merge the sfs-devel package into the core sfs package.  Move some
  misplaced files between sfs and sfs-servers packages.  Clean up the
  upgrade process.  Make /sfs a %%ghost directory for easier removal.

* Sun Dec 01 2002 Kevin Fu <fubob@mit.edu>
- Add conditional magic to avoid compiling with gcc 2.96

* Sat Nov 30 2002 Kevin Fu <fubob@mit.edu>
- Removed unnecessary db3 dependency to allow clean
  installation on both RH 7.3 and 8.0. 

* Thu Jul 11 2002 Kevin Fu <fubob@mit.edu>
- Cleaned up for the recently released 0.6.  Added man pages.

* Sat May 04 2002 Kevin Fu <fubob@mit.edu>
- Cleaned up, synced with actual files in SFS 0.6 candidate

* Wed Jan 09 2002 Kevin Fu <fubob@mit.edu>
- Updated for new files in SFS 0.6 with SFSRO

* Wed Feb 09 2000 Michael Kaminsky <kaminsky@lcs.mit.edu>
- generalized to a generated autoconf file to be version independent 

* Tue Feb 08 2000 Michael Kaminsky <kaminsky@lcs.mit.edu>
- first build
