package dnstap

/*
    Copyright (c) 2013 by Internet Systems Consortium, Inc. ("ISC")

    Permission to use, copy, modify, and/or distribute this software for any
    purpose with or without fee is hereby granted, provided that the above
    copyright notice and this permission notice appear in all copies.

    THE SOFTWARE IS PROVIDED "AS IS" AND ISC DISCLAIMS ALL WARRANTIES
    WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
    MERCHANTABILITY AND FITNESS.  IN NO EVENT SHALL ISC BE LIABLE FOR
    ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
    WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
    ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT
    OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/

import "bytes"
import "fmt"
import "net"
import "strconv"
import "strings"
import "time"

import "github.com/miekg/dns"

import dnstapProto "golang-dnstap/dnstap.pb"

const yamlTimeFormat = "2006-01-02 15:04:05.999999999"

func yamlConvertMessage(m *dnstapProto.Message, s *bytes.Buffer) {
    s.WriteString(fmt.Sprint("  type: ", m.Type, "\n"))

    if m.QueryTimeSec != nil && m.QueryTimeNsec != nil {
        t := time.Unix(int64(*m.QueryTimeSec), int64(*m.QueryTimeNsec)).UTC()
        s.WriteString(fmt.Sprint("  query_time: !!timestamp ", t.Format(yamlTimeFormat), "\n"))
    }

    if m.ResponseTimeSec != nil && m.ResponseTimeNsec != nil {
        t := time.Unix(int64(*m.ResponseTimeSec), int64(*m.ResponseTimeNsec)).UTC()
        s.WriteString(fmt.Sprint("  response_time: !!timestamp ", t.Format(yamlTimeFormat), "\n"))
    }

    if m.SocketFamily != nil {
        s.WriteString(fmt.Sprint("  socket_family: ", m.SocketFamily, "\n"))
    }

    if m.SocketProtocol != nil {
        s.WriteString(fmt.Sprint("  socket_protocol: ", m.SocketProtocol, "\n"))
    }

    if m.QueryAddress != nil {
        s.WriteString(fmt.Sprint("  query_address: ", net.IP(m.QueryAddress), "\n"))
    }

    if m.ResponseAddress != nil {
        s.WriteString(fmt.Sprint("  response_address: ", net.IP(m.ResponseAddress), "\n"))
    }

    if m.QueryPort != nil {
        s.WriteString(fmt.Sprint("  query_port: ", *m.QueryPort, "\n"))
    }

    if m.ResponsePort != nil {
        s.WriteString(fmt.Sprint("  response_port: ", *m.ResponsePort, "\n"))
    }

    if m.QueryZone != nil {
        name, _, err := dns.UnpackDomainName(m.QueryZone, 0)
        if err != nil {
            s.WriteString("  # query_zone: parse failed\n")
        } else {
            s.WriteString(fmt.Sprint("  query_zone: ", strconv.Quote(name), "\n"))
        }
    }

    if m.QueryName != nil {
        name, _, err := dns.UnpackDomainName(m.QueryName, 0)
        if err != nil {
            s.WriteString("  # query_name: parse failed\n")
        }
        s.WriteString(fmt.Sprint("  query_name: ", strconv.Quote(name), "\n"))
    }

    if m.QueryClass != nil {
        s.WriteString(fmt.Sprint("  query_class: ", dns.Class(*m.QueryClass), "\n"))
    }

    if m.QueryType != nil {
        s.WriteString(fmt.Sprint("  query_type: ", dns.Type(*m.QueryType), "\n"))
    }

    if m.QueryMessage != nil {
        msg := new(dns.Msg)
        err := msg.Unpack(m.QueryMessage)
        if err != nil {
            s.WriteString("  # query_message: parse failed\n")
        } else {
            s.WriteString("  query_message: |\n")
            s.WriteString("    " + strings.Replace(strings.TrimSpace(msg.String()), "\n", "\n    ", -1) + "\n")
        }
    }
    if m.ResponseMessage != nil {
        msg := new(dns.Msg)
        err := msg.Unpack(m.ResponseMessage)
        if err != nil {
            s.WriteString("  # response_message: parse failed\n")
        } else {
            s.WriteString("  response_message: |\n")
            s.WriteString("    " + strings.Replace(strings.TrimSpace(msg.String()), "\n", "\n    ", -1) + "\n")
        }
    }
    s.WriteString("---\n")
}

func yamlConvertPayload(dt *dnstapProto.Dnstap) (out []byte) {
    var s bytes.Buffer

    s.WriteString(fmt.Sprint("type: ", dt.Type, "\n"))
    if dt.Identity != nil {
        s.WriteString(fmt.Sprint("identity: ", strconv.Quote(string(dt.Identity)), "\n"))
    }
    if dt.Version != nil {
        s.WriteString(fmt.Sprint("version: ", strconv.Quote(string(dt.Version)), "\n"))
    }
    if *dt.Type == dnstapProto.Dnstap_MESSAGE {
        s.WriteString("message:\n")
        yamlConvertMessage(dt.Message, &s)
    }
    return s.Bytes()
}

func YamlConvert(buf []byte) (out []byte, ok bool) {
    dt, ok := Unpack(buf)
    if ok {
        return yamlConvertPayload(dt), true
    } else {
        return nil, false
    }
}