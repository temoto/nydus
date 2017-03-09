# What
Nydus is minimalistic SOCKS5 proxy server application able to limit incoming/outgoing connections.

# Use case
Your project uses external service such as payment gate which only allows requests from set of whitelisted IPs.
But you have no control on source IP (autoscale, serverless).
Solution:

* create separate well secured system with static IP and ask external to whitelist only that
* run nydus proxy on secure machine
* configure application to run sensitive requests via nydus proxy

# Install
`go get https://github.com/temoto/nydus`

# Usage
* By default, no connections allowed. You must specify allow and deny rules explicitly. Check order: allow, deny.
* Systemd `Type=notify` and watchdog is supported.
* IPv6 is supported.
* Filtering by host names is not supported. You have to specify IP range in CIDR format.

Examples:

* `nydus -listen=10.0.0.4:8891 -allow='0.0.0.0/0:*->1.2.3.4/32:80'`  
Allow IPv4 connections from any address:port to single address 1.2.3.4:80.
* `nydus -listen=10.0.0.4:8891 -allow='10.0.0.0/16:*->77.88.0.0/16:443' -deny='0.0.0.0/0:*->77.88.7.0/24:*'`  
Allow IPv4 connections from any address:port to CIDR 77.88.0.0/16 port 443, except CIDR 77.88.7.0/24.

# Contact
* [https://github.com/temoto/nydus](https://github.com/temoto/nydus)
* [temotor@gmail.com](mailto:temotor@gmail.com)

# Flair
* [![Build Status](https://travis-ci.org/temoto/nydus.svg?branch=master)](https://travis-ci.org/temoto/nydus)
* [![Coverage](https://codecov.io/gh/temoto/nydus/branch/master/graph/badge.svg)](https://codecov.io/gh/temoto/nydus)
