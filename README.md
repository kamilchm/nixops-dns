## NixOps-DNS

DNS server for resolving [NixOps](https://github.com/NixOS/nixops) machines.

### Why

When using NixOps you could easily spawn clusters of machines using VirtualBox or
any cloud provider. Machines from the same deployment use their names for network
connections.

It's useful to access same names locally from the node where NixOps was run.

### How

**NixOps-DNS** listens for DNS queries on `127.0.0.1:5300`. It looks into NixOps state
file and returns machine IP if there is corresponding entry.

#### Installation

**NixOps-DNS** is built with Go and you can install it by doing:
```
$ go get github.com/kamilchm/nixops-dns
```

You can then run it:

```
$ nixops-dns
```

#### Try it!

There's simple example of running multiple machines in [NixOps manual](https://nixos.org/nixops/manual/#idm140737319306144).

Run it, start `nixops-dns` and you should be able to resolve NixOps machine names:
```
$ dig +short proxy @127.0.0.1 -p 5300
192.168.56.103
$ dig +short backend1 @127.0.0.1 -p 5300
192.168.56.101
$ dig +short backend2 @127.0.0.1 -p 5300
192.168.56.102
```

#### System wide NixOps-DNS

The simplest way of using **NixOps-DNS** from any program (terminal, chrome, firefox ...)
is to add it to **DNSMASQ** config:
```
server=/./127.0.0.1#5300
```

