#!/usr/bin/env python3

# Multi Connection Server
# https://realpython.com/python-sockets/#multi-connection-server

import selectors
import socket
import sys
import types

sel = selectors.DefaultSelector()


class SmtpServer():
    def __init__(self, host, port, domain):
        self.host = host
        self.port = port
        self.domain = domain
        self.clients = {}
        self.allow_vrfy = True
        self.allow_expn = True
        self.allow_rcpt_to = True
        self.users = ['root', 'james']

    # Register a client connection
    def register(self, addr, conn: socket.socket):
        conn.send(f'220 {self.domain} ESMTP FakeSmtpServer 1.0.0/1.0.0; Thu, 12 May 2022 10:09:46 +0000\n'.encode())

        self.clients[self._get_client_id(addr=addr)] = {
            'banner': True,
            'conn': conn
        }

    def close(self, addr):
        self.clients[self._get_client_id(addr=addr)] = None

    def _get_client_id(self, addr):
        return f'{addr[0]}:{addr[1]}'

    def get_client(self, addr):
        return self.clients[self._get_client_id(addr=addr)]

    def get_client_sock(self, addr) -> socket.socket:
        return self.clients[self._get_client_id(addr=addr)]['conn']

    def accept_wrapper(self, sock):
        conn, addr = sock.accept()  # Should be ready to read
        print(f"Accepted connection from {addr}")
        conn.setblocking(False)  # Set socket to non-blocking
        data = types.SimpleNamespace(addr=addr, inb=b"", outb=b"")
        events = selectors.EVENT_READ | selectors.EVENT_WRITE
        sel.register(conn, events, data=data)
        self.register(addr=addr, conn=conn)

    def _vrfy(self, username, sock):
        if not self.allow_vrfy:
            sock.send(b'502 VRFY disallowed.\n')
            return

        if username in self.users:
            sock.send(f'250 2.1.5 {username} <{username}@{self.domain}>\n'.encode())
        else:
            sock.send(f'550 5.1.1 {username}... User unknown\n'.encode())

    def _expn(self, username, sock):
        if not self.allow_expn:
            sock.send(b'502 Unimplemented command.\n')
            return

        if username in self.users:
            sock.send(f'250 2.1.5 {username} <{username}@{self.domain}>\n'.encode())
        else:
            sock.send(f'550 5.1.1 {username}... User unknown\n'.encode())

    # def _rcpt_to(self, username, sock):
    #     if not self.allow_rcpt_to:
    #         sock.send(b'502 Unimplemented command.\n')
    #         return
    #
    #     if username in self.users:
    #         sock.send(f'250 2.1.5 {username} <{username}@{self.domain}>\n'.encode())
    #     else:
    #         sock.send(f'550 5.1.1 {username}... User unknown\n'.encode())

    def handle_revc(self, data):
        # Fetch client socket
        sock = self.get_client_sock(addr=data.addr)
        print(f'[{self._get_client_id(data.addr)}]: {data.outb.decode().strip()}')

        if data.outb[:5] == b'VRFY ':
            self._vrfy(data.outb[5:].decode().strip(), sock=sock)

        if data.outb[:5] == b'EXPN ':
            self._expn(data.outb[5:].decode().strip(), sock=sock)

    def service_connection(self, key, mask):
        sock = key.fileobj  # type: socket.socket
        data = key.data

        if mask & selectors.EVENT_READ:  # Ready to read
            recv_data = sock.recv(1024)
            if recv_data:
                data.outb += recv_data
            else:
                print(f"Closing connection to {data.addr}")
                sel.unregister(sock)
                sock.close()
        if mask & selectors.EVENT_WRITE:  # Ready to write
            if data.outb:
                self.handle_revc(data=data)
                sent = len(data.outb)
                data.outb = data.outb[sent:]


def main(host, port):
    # Bind to socket
    lsock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    lsock.bind((host, port))
    lsock.listen()
    print(f"Listening on {(host, port)}")
    lsock.setblocking(False)  # Configure socket in non-blocking mode
    sel.register(lsock, selectors.EVENT_READ, data=None)

    try:
        # Event loop
        while True:
            events = sel.select(timeout=None)  # Block until there are sockets ready for I/O
            for key, mask in events:
                if key.data is None:  # Accept a socket connection and set in non-blocking mode
                    smtp.accept_wrapper(key.fileobj)
                else:  # Handle client connection
                    smtp.service_connection(key, mask)
    except KeyboardInterrupt:
        print("Caught keyboard interrupt, exiting")
    finally:
        sel.close()


if __name__ == "__main__":
    host, port = sys.argv[1], int(sys.argv[2])

    smtp = SmtpServer(host=host, port=port, domain='mail.example.tld')
    main(host=host, port=port)
