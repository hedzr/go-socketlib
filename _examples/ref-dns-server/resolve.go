package main

import (
	"fmt"

	"github.com/miekg/dns"
)

func resolve(domain string, qtype uint16) []dns.RR {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	c := new(dns.Client)
	in, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, ans := range in.Answer {
		fmt.Println(ans)
	}

	return in.Answer
}

// func main() {
// 	resolve("example.com", dns.TypeA)
// }
