package httpheader

import (
	"net"
	"net/http"
	"strconv"
	"strings"
)

// Forwarded parses the Forwarded header from h (RFC 7239).
//
// Any valid elements at the end of the header are guaranteed to be parsed, even if
// they are preceded by malformed elements. This ensures that any information
// appended by a trusted gateway is recovered regardless of what was received from
// the (untrusted) client. Establishing trust is outside the scope of this package.
func Forwarded(h http.Header) []ForwardedElem {
	var elems []ForwardedElem
	for v, vs := iterElems("", h["Forwarded"]); vs != nil; v, vs = iterElems(v, vs) {
		var elem ForwardedElem
		for {
			var name, value string
			name, v = consumeItem(v)
			name = strings.ToLower(name)
			if name == "" { // no forwarded-pair
				if peek(v) == ';' {
					v = v[1:]
					continue
				}
				break
			}
			if peek(v) != '=' {
				break
			}
			v = v[1:]
			if peek(v) == '"' {
				var ok bool
				value, v, ok = consumeQuoted(v, true)
				if !ok { // unterminated string
					break
				}
			} else {
				value, v = consumeItem(v)
			}
			switch name {
			case "for":
				elem.For = value
			case "by":
				elem.By = value
			case "host":
				elem.Host = value
			case "proto":
				elem.Proto = strings.ToLower(value)
			default:
				if elem.Ext == nil {
					elem.Ext = make(map[string]string)
				}
				elem.Ext[name] = value
			}
		}
		elems = append(elems, elem)
	}
	return elems
}

// SetForwarded replaces the Forwarded header in h (RFC 7239).
func SetForwarded(h http.Header, elems []ForwardedElem) {
	h.Set("Forwarded", buildForwarded(elems))
}

// AddForwarded is like SetForwarded but appends instead of replacing.
func AddForwarded(h http.Header, elems ...ForwardedElem) {
	h.Add("Forwarded", buildForwarded(elems))
}

func buildForwarded(elems []ForwardedElem) string {
	b := &strings.Builder{}
	for i, elem := range elems {
		if i > 0 {
			b.WriteString(", ")
		}
		var wrote bool
		if elem.For != "" {
			wrote = writeParam(b, wrote, "for", elem.For)
		}
		if elem.By != "" {
			wrote = writeParam(b, wrote, "by", elem.By)
		}
		if elem.Host != "" {
			wrote = writeParam(b, wrote, "host", elem.Host)
		}
		if elem.Proto != "" {
			wrote = writeParam(b, wrote, "proto", elem.Proto)
		}
		for name, value := range elem.Ext {
			wrote = writeParam(b, wrote, name, value)
		}
	}
	return b.String()
}

// A ForwardedElem represents one element of the Forwarded header (RFC 7239).
// Standard parameters are stored in the corresponding fields;
// any extension parameters are stored in Ext.
type ForwardedElem struct {
	By    string
	For   string
	Host  string
	Proto string            // lowercased
	Ext   map[string]string // keys lowercased
}

// ByAddr returns the IP address and port from the By field of elem.
// If either is missing or cannot be parsed, the respective zero value is returned.
func (elem ForwardedElem) ByAddr() (net.IP, int) {
	return nodeAddr(elem.By)
}

// ForAddr returns the IP address and port from the For field of elem.
// If either is missing or cannot be parsed, the respective zero value is returned.
func (elem ForwardedElem) ForAddr() (net.IP, int) {
	return nodeAddr(elem.For)
}

func nodeAddr(node string) (net.IP, int) {
	rawIP, rawPort := node, ""
	portPos := strings.LastIndexByte(node, ':')
	if portPos < strings.IndexByte(node, ']') {
		// That's not a port, that's part of the IPv6 address.
		portPos = -1
	}
	if portPos != -1 {
		rawIP, rawPort = node[:portPos], node[portPos+1:]
	}
	rawIP = strings.TrimPrefix(rawIP, "[")
	rawIP = strings.TrimSuffix(rawIP, "]")
	ip := net.ParseIP(rawIP)
	port, _ := strconv.Atoi(rawPort)
	return ip, port
}