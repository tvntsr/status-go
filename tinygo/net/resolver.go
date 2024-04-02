package net

import (
    "context"
)

type Resolver struct {
    // PreferGo controls whether Go's built-in DNS resolver is preferred
    // on platforms where it's available. It is equivalent to setting
    // GODEBUG=netdns=go, but scoped to just this resolver.
    PreferGo bool

    // StrictErrors controls the behavior of temporary errors
    // (including timeout, socket errors, and SERVFAIL) when using
    // Go's built-in resolver. For a query composed of multiple
    // sub-queries (such as an A+AAAA address lookup, or walking the
    // DNS search list), this option causes such errors to abort the
    // whole query instead of returning a partial result. This is
    // not enabled by default because it may affect compatibility
    // with resolvers that process AAAA queries incorrectly.
    StrictErrors bool

    // Dial optionally specifies an alternate dialer for use by
    // Go's built-in DNS resolver to make TCP and UDP connections
    // to DNS services. The host in the address parameter will
    // always be a literal IP address and not a host name, and the
    // port in the address parameter will be a literal port number
    // and not a service name.
    // If the Conn returned is also a PacketConn, sent and received DNS
    // messages must adhere to RFC 1035 section 4.2.1, "UDP usage".
    // Otherwise, DNS messages transmitted over Conn must adhere
    // to RFC 7766 section 5, "Transport Protocol Selection".
    // If nil, the default dialer is used.
    Dial func(ctx context.Context, network, address string) (Conn, error)
    // contains filtered or unexported fields
}

var DefaultResolver = &Resolver{}

func (r *Resolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
    panic("Not implemented")
}
func (r *Resolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
    panic("Not implemented")
}
func (r *Resolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
    panic("Not implemented")
}
func (r *Resolver) LookupCNAME(ctx context.Context, host string) (string, error) {
    panic("Not implemented")
}


func ResolveIPAddr(network, address string) (*IPAddr, error) {
    panic("Not implemented")
}

func ListenTCP(network string, laddr *TCPAddr) (*TCPListener, error) {
    panic("Not implemented")
}

func ListenUDP(network string, laddr *UDPAddr) (*UDPConn, error) {
    panic("Not implemented")
}

func ResolveUnixAddr(network, address string) (*UnixAddr, error) {
    panic("Not implemented")
}

func (c *UDPConn) ReadFromUDP(b []byte) (n int, addr *UDPAddr, err error) {
    panic("Not implemented")
}
func (c *UDPConn) WriteToUDP(b []byte, addr *UDPAddr) (int, error)  {
    panic("Not implemented")
}

func ListenMulticastUDP(network string, ifi *Interface, gaddr *UDPAddr) (*UDPConn, error) {
    panic("Not implemented")
}

func ListenPacket(network, address string) (PacketConn, error) {
    panic("Not implemented")
}

func LookupIP(host string) ([]IP, error){
    panic("Not implemented")
}

type DNSError struct {
    Err         string // description of the error
    Name        string // name looked for
    Server      string // server used
    IsTimeout   bool   // if true, timed out; not all timeouts set this
    IsTemporary bool   // if true, error is temporary; not all errors set this
    IsNotFound  bool   // if true, host could not be found
}
func (e *DNSError) Error() string {
    return "DNS error"
}

func (ifi *Interface) Addrs() ([]Addr, error){
    panic("Not implemented")
}