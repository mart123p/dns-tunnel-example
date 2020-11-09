#!/bin/sh

# Setup the port forwarding from ssh connexion for remote server
ip=$(cat ../creds/ip)
ssh -L 6001:10.0.0.3:5986 \
    -L 6002:10.0.0.4:5986 \
    \
    -L 6011:10.0.0.3:3389 \
    -L 6012:10.0.0.4:3389 \
    \
    -L 6050:10.0.5.2:22 \
    -i ../creds/ssh/id_rsa \
    azureuser@$ip
