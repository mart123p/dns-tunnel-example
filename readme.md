# DNS Tunneling

For a Network security class, we needed to choose a topic relating to network security.
We have chosen DNS Tunneling. The goal of this repo is to illustrate how DNS Tunneling
work and how it can be used by a command and control server.

## Command and control server
The command and control server is used to control a single comprimsed machine. It will 
provide a shell once the client connects to it by DNS. An aditional command `upload` is
provided to upload documents to the server. It is used to exfiltrate data during our
demonstration.

While there are similar tools out there, we wanted to have a simple tool that demonstrate
DNS tunneling.