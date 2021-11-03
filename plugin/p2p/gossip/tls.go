package gossip

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"syscall"

	"google.golang.org/grpc/credentials"
)

//Tls defines the specific interface for all the live gRPC wire
// protocols and supported transport security protocols (e.g., TLS, SSL).
type Tls struct {
	config *tls.Config
}

type certInfo struct {
	revoke bool
	ip     string
	serial string
}

var (
	serials       = make(map[string]*certInfo)
	revokeLock    sync.Mutex
	latestSerials sync.Map
)

//serialNum -->ip
func addCertSerial(serial string, ip string) {
	revokeLock.Lock()
	defer revokeLock.Unlock()
	serials[serial] = &certInfo{false, ip, serial}

}
func updateCertSerial(serial string, revoke bool) certInfo {
	revokeLock.Lock()
	defer revokeLock.Unlock()
	v, ok := serials[serial]
	if ok {
		v.revoke = revoke
		return *v
	}

	return certInfo{}
}

func isRevoke(serial string) bool {
	revokeLock.Lock()
	defer revokeLock.Unlock()
	if r, ok := serials[serial]; ok {
		return r.revoke
	}
	return false
}

func removeCertSerial(serial string) {
	revokeLock.Lock()
	defer revokeLock.Unlock()
	delete(serials, serial)
}

func getSerialNums() []string {
	revokeLock.Lock()
	defer revokeLock.Unlock()
	var certs []string
	for s := range serials {
		certs = append(certs, s)
	}
	return certs
}

func getSerials() map[string]*certInfo {
	revokeLock.Lock()
	defer revokeLock.Unlock()
	var certs = make(map[string]*certInfo)
	for k, v := range serials {
		certs[k] = v
	}
	return certs
}

func (c Tls) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{
		SecurityProtocol: "tls",
		SecurityVersion:  "1.2",
		ServerName:       c.config.ServerName,
	}
}

func CloneTLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		return &tls.Config{}
	}

	return cfg.Clone()
}
func (c *Tls) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (_ net.Conn, _ credentials.AuthInfo, err error) {

	// use local cfg to avoid clobbering ServerName if using multiple endpoints
	cfg := CloneTLSConfig(c.config)
	if cfg.ServerName == "" {
		serverName, _, err := net.SplitHostPort(authority)
		if err != nil {
			// If the authority had no host port or if the authority cannot be parsed, use it as-is.
			serverName = authority
		}
		cfg.ServerName = serverName
	}
	conn := tls.Client(rawConn, cfg)
	errChannel := make(chan error, 1)
	go func() {
		errChannel <- conn.Handshake()
		close(errChannel)
	}()
	select {
	case err := <-errChannel:
		if err != nil {
			conn.Close()
			return nil, nil, err
		}
	case <-ctx.Done():
		conn.Close()
		return nil, nil, ctx.Err()
	}
	tlsInfo := credentials.TLSInfo{
		State: conn.ConnectionState(),
		CommonAuthInfo: credentials.CommonAuthInfo{
			SecurityLevel: credentials.PrivacyAndIntegrity,
		},
	}
	peerCert := tlsInfo.State.PeerCertificates
	//校验CERT
	certNum := len(peerCert)
	if certNum > 0 {
		peerSerialNum := peerCert[0].SerialNumber
		log.Debug("ClientHandshake", "Certificate SerialNumber", peerSerialNum, "Certificate Number", certNum, "RemoteAddr", rawConn.RemoteAddr(), "tlsInfo", tlsInfo)
		addrSplites := strings.Split(rawConn.RemoteAddr().String(), ":")
		//检查证书是否被吊销
		if isRevoke(peerSerialNum.String()) {
			conn.Close()
			return nil, nil, errors.New(fmt.Sprintf("transport: authentication handshake failed: ClientHandshake Certificate SerialNumber %v revoked", peerSerialNum.String()))
		}

		if len(addrSplites) > 0 { //服务端证书的序列号，已经其IP地址
			addCertSerial(peerSerialNum.String(), addrSplites[0])
			latestSerials.Store(addrSplites[0], peerSerialNum.String()) //ip --->serialNum
		}
	}

	id := SPIFFEIDFromState(conn.ConnectionState())
	if id != nil {
		tlsInfo.SPIFFEID = id
	}
	return WrapSyscallConn(rawConn, conn), tlsInfo, nil
}

//ServerHandshake check cert
func (c *Tls) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn := tls.Server(rawConn, c.config)
	if err := conn.Handshake(); err != nil {
		conn.Close()
		return nil, nil, err
	}
	tlsInfo := credentials.TLSInfo{
		State: conn.ConnectionState(),
		CommonAuthInfo: credentials.CommonAuthInfo{
			SecurityLevel: credentials.PrivacyAndIntegrity,
		},
	}
	peerCert := tlsInfo.State.PeerCertificates
	//校验CERT
	certNum := len(peerCert)
	if certNum != 0 {
		peerSerialNum := peerCert[0].SerialNumber
		log.Debug("ServerHandshake", "peerSerialNum", peerSerialNum, "Certificate Number", certNum, "RemoteAddr", rawConn.RemoteAddr(), "tlsinfo", tlsInfo, "remoteAddr", conn.RemoteAddr())

		if isRevoke(peerSerialNum.String()) {
			rawConn.Close()
			return nil, nil, errors.New(fmt.Sprintf("transport: authentication handshake failed: ServerHandshake  %s  revoked", peerSerialNum.String()))
		}
		addrSplites := strings.Split(rawConn.RemoteAddr().String(), ":")
		if len(addrSplites) > 0 {
			addCertSerial(peerSerialNum.String(), addrSplites[0])
			latestSerials.Store(addrSplites[0], peerSerialNum.String()) //ip --->serialNum
		}

	} else {
		log.Debug("ServerHandshake", "info", tlsInfo)
	}

	id := SPIFFEIDFromState(conn.ConnectionState())
	if id != nil {
		tlsInfo.SPIFFEID = id
	}

	return WrapSyscallConn(rawConn, conn), tlsInfo, nil
}

//  uses c to construct a TransportCredentials based on TLS.
func newTLS(c *tls.Config) credentials.TransportCredentials {
	tc := &Tls{}
	tc.config = CloneTLSConfig(c)
	//tc.serials=make(map[*big.Int]*certInfo)
	tc.config.NextProtos = AppendH2ToNextProtos(tc.config.NextProtos)
	return tc
}

func (c *Tls) Clone() credentials.TransportCredentials {
	return newTLS(c.config)
}

func (c *Tls) OverrideServerName(serverNameOverride string) error {
	c.config.ServerName = serverNameOverride
	return nil
}

func SPIFFEIDFromState(state tls.ConnectionState) *url.URL {
	if len(state.PeerCertificates) == 0 || len(state.PeerCertificates[0].URIs) == 0 {
		return nil
	}
	return SPIFFEIDFromCert(state.PeerCertificates[0])
}

// SPIFFEIDFromCert parses the SPIFFE ID from x509.Certificate. If the SPIFFE
// ID format is invalid, return nil with warning.
func SPIFFEIDFromCert(cert *x509.Certificate) *url.URL {
	if cert == nil || cert.URIs == nil {
		return nil
	}
	var spiffeID *url.URL
	for _, uri := range cert.URIs {
		if uri == nil || uri.Scheme != "spiffe" || uri.Opaque != "" || (uri.User != nil && uri.User.Username() != "") {
			continue
		}
		// From this point, we assume the uri is intended for a SPIFFE ID.
		if len(uri.String()) > 2048 {
			//logger.Warning("invalid SPIFFE ID: total ID length larger than 2048 bytes")
			return nil
		}
		if len(uri.Host) == 0 || len(uri.Path) == 0 {
			//logger.Warning("invalid SPIFFE ID: domain or workload ID is empty")
			return nil
		}
		if len(uri.Host) > 255 {
			//logger.Warning("invalid SPIFFE ID: domain length larger than 255 characters")
			return nil
		}
		// A valid SPIFFE certificate can only have exactly one URI SAN field.
		if len(cert.URIs) > 1 {
			//logger.Warning("invalid SPIFFE ID: multiple URI SANs")
			return nil
		}
		spiffeID = uri
	}
	return spiffeID
}

type sysConn = syscall.Conn

type syscallConn struct {
	net.Conn
	// sysConn is a type alias of syscall.Conn. It's necessary because the name
	// `Conn` collides with `net.Conn`.
	sysConn
}

func WrapSyscallConn(rawConn, newConn net.Conn) net.Conn {
	sysConn, ok := rawConn.(syscall.Conn)
	if !ok {
		return newConn
	}
	return &syscallConn{
		Conn:    newConn,
		sysConn: sysConn,
	}
}

const alpnProtoStrH2 = "h2"

// AppendH2ToNextProtos appends h2 to next protos.
func AppendH2ToNextProtos(ps []string) []string {
	for _, p := range ps {
		if p == alpnProtoStrH2 {
			return ps
		}
	}
	ret := make([]string, 0, len(ps)+1)
	ret = append(ret, ps...)
	return append(ret, alpnProtoStrH2)
}
