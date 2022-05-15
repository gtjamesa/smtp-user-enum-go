# SMTP User Enum (go)

This is a simple application to enumerate a target SMTP service. It was developed as a test project to learn Golang.

```bash
$ ./smtp-user-enum-go --help
NAME:
   SMTP User Enum - A simple SMTP user enumeration program

USAGE:
   smtp-user-enum [global options] command [command options] <target>

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --method method, -m method     Enumeration method to use (allowed: VRFY, EXPN, RCPT) (default: "VRFY")
   --port port, -p port           Set a non-standard SMTP port (default: 25)
   --threads threads, -t threads  Amount of threads to use (default: 8)
   --verbose, -v                  Set verbose output (default: false)
   --wordlist value, -w value     Wordlist containing usernames
   --help, -h                     show help (default: false)
```

## Usage

```bash
$ ./smtp-user-enum-go -w /usr/share/seclists/Usernames/xato-net-10-million-usernames.txt --verbose 192.168.0.20
220 mail.example.tld ESMTP FakeSmtpServer 1.0.0/1.0.0; Thu, 12 May 2022 10:09:46 +0000
VRFY method disallowed by server
james           250 2.1.5 james <james@mail.example.tld>
root            250 2.1.5 root <root@mail.example.tld>
spooky          250 2.1.5 spooky <spooky@mail.example.tld>
syclone         250 2.1.5 syclone <syclone@mail.example.tld>
taken           250 2.1.5 taken <taken@mail.example.tld>
```

The program supports the `VRFY`, `EXPN` and `RCPT TO` enumeration methods, and the method can be chosen by supplying the `--method/-m` parameter. The default is `VRFY`. If the chosen (or default) method is disallowed, the program will switch to the first available method.

*Multiple targets are not yet supported.*