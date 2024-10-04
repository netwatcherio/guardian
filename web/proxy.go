package web

import (
	"github.com/kataras/iris/v12"
	"net"
	"strings"
)

var trustedProxies = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"fc00::/7",
}

func isPrivateIP(ip net.IP) bool {
	for _, cidr := range trustedProxies {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

func ProxyIPMiddleware(ctx iris.Context) {
	remoteIP := net.ParseIP(ctx.RemoteAddr())
	if remoteIP == nil {
		ctx.Values().Set("client_ip", ctx.RemoteAddr())
		ctx.Next()
		return
	}

	if !isPrivateIP(remoteIP) {
		ctx.Values().Set("client_ip", remoteIP.String())
		ctx.Next()
		return
	}

	if forwardedFor := ctx.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		for _, ip := range ips {
			parsedIP := net.ParseIP(strings.TrimSpace(ip))
			if parsedIP != nil && !isPrivateIP(parsedIP) {
				ctx.Values().Set("client_ip", parsedIP.String())
				ctx.Next()
				return
			}
		}
	}

	if realIP := ctx.GetHeader("X-Real-IP"); realIP != "" {
		parsedIP := net.ParseIP(realIP)
		if parsedIP != nil && !isPrivateIP(parsedIP) {
			ctx.Values().Set("client_ip", parsedIP.String())
			ctx.Next()
			return
		}
	}

	// If we couldn't determine a public IP, fall back to the remote address
	ctx.Values().Set("client_ip", ctx.RemoteAddr())
	ctx.Next()
}
