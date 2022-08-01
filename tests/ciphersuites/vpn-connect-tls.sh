#!/bin/sh
set -x
sudo openvpn \
    --verb 3 \
    --dev tun --client --tls-client \
    --cipher $CIPHER \
    --remote-cert-tls server --tls-version-min 1.2 \
    --ca /tmp/ca.crt --cert /tmp/cert.pem --key /tmp/cert.pem \
    --pull-filter ignore ifconfig-ipv6 \
    --pull-filter ignore route-ipv6 \
    --pull-filter ignore route \
    --tls-cipher $TLS \
    --remote $GW $PORT $PROTO
