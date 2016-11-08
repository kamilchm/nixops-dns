package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os/user"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/miekg/dns"
)

var nixopsStateDb *sql.DB

func openNixopsStateDb() *sql.DB {
	usr, _ := user.Current()
	db, err := sql.Open("sqlite3", filepath.Join(
		usr.HomeDir, ".nixops/deployments.nixops")+"?mode=ro")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func nixopsHostIp(hostname string) (net.IP, error) {
	var ip string
	row := nixopsStateDb.QueryRow(`
	  select ra.value from Resources r, ResourceAttrs ra
	    where r.name = ? and r.id = ra.machine and ra.name = 'privateIpv4'
    `, hostname)
	if err := row.Scan(&ip); err != nil {
		return nil, fmt.Errorf("Error while trying to find host '%s' in NixOps: %q",
			hostname, err)
	}
	return net.ParseIP(ip), nil
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	q := r.Question[0]

	log.Printf("Question: Type=%s Class=%s Name=%s\n", dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass], q.Name)

	if q.Qtype != dns.TypeA || q.Qclass != dns.ClassINET {
		handleNotFound(w, r)
		return
	}

	ip, err := nixopsHostIp(strings.TrimSuffix(q.Name, "."))
	if err != nil {
		log.Println(err)
		handleNotFound(w, r)
		return
	}

	m := new(dns.Msg)
	m.SetReply(r)
	a := new(dns.A)
	a.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 30}
	a.A = ip
	m.Answer = []dns.RR{a}
	w.WriteMsg(m)
}

func handleNotFound(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Rcode = dns.RcodeNameError // NXDOMAIN
	w.WriteMsg(m)
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:5300", "listen address")

	flag.Parse()

	nixopsStateDb = openNixopsStateDb()
	defer nixopsStateDb.Close()

	server := &dns.Server{Addr: *addr, Net: "udp"}
	server.Handler = dns.HandlerFunc(handleRequest)

	log.Printf("Listening on %s\n", *addr)
	log.Fatal(server.ListenAndServe())
}
