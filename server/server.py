from dnslib import RR
from dnslib.server import DNSServer, BaseResolver, DNSLogger
import argparse
import queue

qCommands = queue.Queue()


class TunnelResolver(BaseResolver):
    def resolve(self, request, handler):
        reply = request.reply()
        qname = request.q.qname
        try:
            command_txt = qCommands.get(block=False, timeout=None)
        except queue.Empty:
            command_txt = ""

        reply.add_answer(*RR.fromZone("{} 60 IN TXT \"{}\"".format(qname, command_txt)))
        return reply


p = argparse.ArgumentParser(description="Fixed DNS Resolver")
p.add_argument("--port", "-p", type=int, default=53,
               metavar="<port>",
               help="Server port (default:53)")
p.add_argument("--address", "-a", default="",
               metavar="<address>",
               help="Listen address (default:all)")
p.add_argument("--udplen", "-u", type=int, default=0,
               metavar="<udplen>",
               help="Max UDP packet length (default:0)")
p.add_argument("--tcp", action='store_true', default=False,
               help="TCP server (default: UDP only)")
p.add_argument("--log", default="truncated,error",
               help="Log hooks to enable (default: +request,+reply,+truncated,+error,-recv,-send,-data)")
p.add_argument("--log-prefix", action='store_true', default=False,
               help="Log prefix (timestamp/handler/resolver) (default: False)")

args = p.parse_args()
resolver = TunnelResolver()
logger = DNSLogger(args.log, args.log_prefix)

print("Starting Tunnel Resolver (%s:%d) [%s]" % (
    args.address or "*",
    args.port,
    "UDP/TCP" if args.tcp else "UDP"))

udp_server = DNSServer(resolver,
                       port=args.port,
                       address=args.address,
                       logger=logger)
udp_server.start_thread()

if args.tcp:
    tcp_server = DNSServer(resolver,
                           port=args.port,
                           address=args.address,
                           tcp=True,
                           logger=logger)
    tcp_server.start_thread()

while udp_server.isAlive():
    command = input("$ ")
    if len(command) <= 255:
        qCommands.put(command)
