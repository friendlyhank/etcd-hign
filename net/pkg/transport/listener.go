package transport

import "net"

type TLSInfo struct{
}

func newListener(addr string,scheme string)(net.Listener,error){
	if scheme == "unix" || scheme == "unixs"{
		return NewUnixListener(addr)
	}
	return net.Listen("tcp",addr)
}