package dns

import (
	"context"
	"net/netip"

	"github.com/miekg/dns"
)

type edns0SubnetTransportWrapper struct {
	Transport
	clientSubnet netip.Addr
}

func (t *edns0SubnetTransportWrapper) Exchange(ctx context.Context, message *dns.Msg) (*dns.Msg, error) {
	SetClientSubnet(message, t.clientSubnet, false)
	return t.Transport.Exchange(ctx, message)
}

func SetClientSubnet(message *dns.Msg, clientSubnet netip.Addr, override bool) {
	var (
		optRecord    *dns.OPT
		subnetOption *dns.EDNS0_SUBNET
	)
findExists:
	for _, record := range message.Extra {
		var isOPTRecord bool
		if optRecord, isOPTRecord = record.(*dns.OPT); isOPTRecord {
			for _, option := range optRecord.Option {
				var isEDNS0Subnet bool
				subnetOption, isEDNS0Subnet = option.(*dns.EDNS0_SUBNET)
				if isEDNS0Subnet {
					if !override {
						return
					}
					break findExists
				}
			}
		}
	}
	if optRecord == nil {
		optRecord = &dns.OPT{
			Hdr: dns.RR_Header{
				Name:   ".",
				Rrtype: dns.TypeOPT,
			},
		}
		message.Extra = append(message.Extra, optRecord)
	}
	if subnetOption == nil {
		subnetOption = new(dns.EDNS0_SUBNET)
		optRecord.Option = append(optRecord.Option, subnetOption)
	}
	subnetOption.Code = dns.EDNS0SUBNET
	if clientSubnet.Is4() {
		subnetOption.Family = 1
	} else {
		subnetOption.Family = 2
	}
	subnetOption.SourceNetmask = uint8(clientSubnet.BitLen())
	subnetOption.Address = clientSubnet.AsSlice()
}
