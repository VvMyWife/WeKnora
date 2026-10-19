#!/usr/bin/env python3
"""Run an explicit command on the authoritative WeKnora server using Paramiko."""
import argparse
import getpass
import sys
import paramiko

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("command")
args = parser.parse_args()
client = paramiko.SSHClient()
client.load_system_host_keys()
client.set_missing_host_key_policy(paramiko.RejectPolicy())
client.connect("10.20.0.15", username="ubuntu",
               password=getpass.getpass("Ubuntu SSH password: "),
               allow_agent=False, look_for_keys=False, timeout=15)
try:
    _, out, err = client.exec_command("export GIT_PAGER=cat PAGER=cat; " + args.command, timeout=300)
    for line in iter(out.readline, ""):
        print(line, end="", flush=True)
    print(err.read().decode("utf-8", "replace"), end="", file=sys.stderr)
    raise SystemExit(out.channel.recv_exit_status())
finally:
    client.close()
