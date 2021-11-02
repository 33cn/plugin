// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gossip

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc/credentials"

	pb "github.com/33cn/chain33/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
)

// NetAddress defines information about a peer on the network
// including its IP address, and port.
type NetAddress struct {
	IP   net.IP
	Port uint16
	str  string
}

// NewNetAddress returns a new NetAddress using the provided TCP
// address.
func NewNetAddress(addr net.Addr) *NetAddress {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return nil
	}
	ip := tcpAddr.IP
	port := uint16(tcpAddr.Port)
	return NewNetAddressIPPort(ip, port)
}

// NewNetAddressString returns a new NetAddress using the provided
// address in the form of "IP:Port". Also resolves the host if host
// is not an IP.
func NewNetAddressString(addr string) (*NetAddress, error) {

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(host)
	if ip == nil {
		if len(host) > 0 {
			ips, err := net.LookupIP(host)
			if err != nil {
				return nil, err
			}
			ip = ips[0]
		}
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, err
	}

	na := NewNetAddressIPPort(ip, uint16(port))
	return na, nil
}

// NewNetAddressStrings returns an array of NetAddress'es build using
// the provided strings.
func NewNetAddressStrings(addrs []string) ([]*NetAddress, error) {
	netAddrs := make([]*NetAddress, len(addrs))
	for i, addr := range addrs {
		netAddr, err := NewNetAddressString(addr)
		if err != nil {
			return nil, fmt.Errorf("error in address %s: %v", addr, err)
		}
		netAddrs[i] = netAddr
	}
	return netAddrs, nil
}

// NewNetAddressIPPort returns a new NetAddress using the provided IP
// and port number.
func NewNetAddressIPPort(ip net.IP, port uint16) *NetAddress {
	na := &NetAddress{
		IP:   ip,
		Port: port,
		str: net.JoinHostPort(
			ip.String(),
			strconv.FormatUint(uint64(port), 10),
		),
	}
	return na
}

// Equals reports whether na and other are the same addresses.
func (na *NetAddress) Equals(other interface{}) bool {
	if o, ok := other.(*NetAddress); ok {
		return na.String() == o.String()
	}

	return false
}

// Less reports whether na and other are the less addresses
func (na *NetAddress) Less(other interface{}) bool {
	if o, ok := other.(*NetAddress); ok {
		return na.String() < o.String()
	}

	log.Error("Cannot compare unequal types")
	return false
}

// String representation.
func (na *NetAddress) String() string {
	if na.str == "" {
		na.str = net.JoinHostPort(
			na.IP.String(),
			strconv.FormatUint(uint64(na.Port), 10),
		)
	}
	return na.str
}

// Copy na address
func (na *NetAddress) Copy() *NetAddress {
	copytmp := *na
	return &copytmp
}

// DialTimeout calls net.DialTimeout on the address.
func isCompressSupport(err error) bool {
	var errstr = `grpc: Decompressor is not installed for grpc-encoding "gzip"`
	if status.Code(err) == codes.Unimplemented && status.Convert(err).Message() == errstr {
		return false
	}
	return true
}

// DialTimeout dial timeout
func (na *NetAddress) DialTimeout(version int32, creds credentials.TransportCredentials, bList *BlackList) (*grpc.ClientConn, error) {
	ch := make(chan grpc.ServiceConfig, 1)
	ch <- P2pComm.GrpcConfig()

	var cliparm keepalive.ClientParameters
	cliparm.Time = 15 * time.Second    //keepalive ping 周期
	cliparm.Timeout = 10 * time.Second //ping后的获取ack消息超时时间
	cliparm.PermitWithoutStream = true //启动keepalive 进行检查
	keepaliveOp := grpc.WithKeepaliveParams(cliparm)
	log.Debug("DialTimeout", "Dial------------->", na.String())
	maxMsgSize := pb.MaxBlockSize + 1024*1024
	//配置SSL连接
	var secOpt grpc.DialOption
	if creds == nil {
		secOpt = grpc.WithInsecure()
	} else {
		secOpt = grpc.WithTransportCredentials(creds)
	}
	//接口拦截器
	interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 黑名单校验
		//checkAuth
		log.Debug("interceptor client", "remoteAddr", na.String())
		ip, _, err := net.SplitHostPort(na.String())
		if err != nil {
			return err
		}

		if bList != nil && bList.Has(ip) || bList != nil && bList.Has(na.String()) {
			return fmt.Errorf("interceptor blacklist peer  %v no authorized", na.String())
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}

	//流拦截器
	interceptorStream := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ip, _, err := net.SplitHostPort(na.String())
		if err != nil {
			return nil, err
		}

		if bList.Has(ip) {

			return nil, fmt.Errorf("blacklist peer %v  no authorized", ip)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	conn, err := grpc.DialContext(ctx, na.String(),
		grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxMsgSize)),
		grpc.WithServiceConfig(ch), keepaliveOp, secOpt,
		grpc.WithUnaryInterceptor(interceptor), grpc.WithStreamInterceptor(interceptorStream))
	if err != nil {
		log.Error("grpc DialCon", "did not connect", err, "addr", na.String())
		return nil, err
	}

	//p2p version check 通过版本协议，获取通信session

	//判断是否对方是否支持压缩
	cli := pb.NewP2PgserviceClient(conn)
	_, err = cli.GetHeaders(context.Background(), &pb.P2PGetHeaders{StartHeight: 0, EndHeight: 0, Version: version}, grpc.WaitForReady(false))
	if err != nil && !isCompressSupport(err) {
		//compress not support
		log.Error("compress not supprot , rollback to uncompress version", "addr", na.String())
		err = conn.Close()
		if err != nil {
			log.Error("conn", "close err", err)
		}
		ch2 := make(chan grpc.ServiceConfig, 1)
		ch2 <- P2pComm.GrpcConfig()
		log.Debug("NetAddress", "Dial with unCompressor", na.String())
		conn, err = grpc.DialContext(ctx, na.String(), secOpt, grpc.WithServiceConfig(ch2), keepaliveOp)

	}

	if err != nil {
		log.Debug("grpc DialCon Uncompressor", "connect err", err)
		if conn != nil {
			errs := conn.Close()
			if errs != nil {
				log.Error("conn", "close err", errs)
			}
		}
		return nil, err
	}

	return conn, nil
}
