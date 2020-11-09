#!/bin/sh
# Connect to the rogue server

ssh -i ../creds/ssh/id_rsa \
    -p 6050 \
    azureuser@localhost
