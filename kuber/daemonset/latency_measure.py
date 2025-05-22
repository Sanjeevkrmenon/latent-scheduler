#!/usr/bin/env python3
#
#  Kubernetes node-to-node latency reporter
#

import subprocess
import time
import re
import datetime as dt
import traceback
import os
import sys
import socket

# Configuration (use env vars or defaults)
PING_COUNT   = int(os.getenv("PING_COUNT", "4"))
SLEEP_CYCLE  = int(os.getenv("SLEEP_INTERVAL", "60"))
PING_TIMEOUT = int(os.getenv("PING_TIMEOUT", "15"))
KUBECTL_TIMEOUT = int(os.getenv("KUBECTL_TIMEOUT", "10"))

# For node/pod identity tagging in logs (K8s sets HOSTNAME to pod name, but we go for node name for clarity)
NODE_NAME = os.environ.get("NODE_NAME", socket.gethostname())

def ts(msg: str) -> None:
    print(f"[{dt.datetime.now():%Y-%m-%d %H:%M:%S}][{NODE_NAME}] {msg}", flush=True)

def get_node_ips() -> list:
    ts("Fetching node IPs with kubectl…")
    try:
        out = subprocess.run(
            [
                "kubectl", "get", "nodes",
                "-o",
                'jsonpath={.items[*].status.addresses[?(@.type=="InternalIP")].address}'
            ],
            capture_output=True,
            text=True,
            check=True,
            timeout=KUBECTL_TIMEOUT,
        )
        ips = [ip for ip in out.stdout.split() if ip]
        if not ips:
            ts("WARNING: No node IPs found by kubectl!")
        else:
            ts(f"Node IPs: {ips}")
        return ips
    except FileNotFoundError:
        ts("ERROR: kubectl not found in PATH")
        return []
    except subprocess.TimeoutExpired:
        ts("ERROR: kubectl timed out")
        return []
    except subprocess.CalledProcessError as e:
        ts(f"ERROR: kubectl failed rc={e.returncode}  stderr={e.stderr.strip()}")
        return []
    except Exception:
        ts("Unexpected error while fetching IPs")
        traceback.print_exc()
        return []

# Regular expression for RTT parsing
rtt_re = re.compile(r"rtt min/avg/max/\S+\s*=\s*([0-9.]+)/([0-9.]+)/([0-9.]+)")

def ping_one(ip: str) -> None:
    ts(f"Pinging {ip} …")
    try:
        out = subprocess.run(
            ["ping", "-c", str(PING_COUNT), "-W", str(PING_TIMEOUT), ip],
            capture_output=True,
            text=True,
            timeout=PING_TIMEOUT+2,  # a little extra time for process
            check=False,             # so we capture packet loss properly
        )
        all_text = out.stdout + out.stderr
        loss_m = re.search(r"(\d+)% packet loss", all_text)
        rtt_m  = rtt_re.search(out.stdout)

        loss = loss_m.group(1) if loss_m else "?"

        if rtt_m:
            min_, avg_, max_ = rtt_m.groups()
            ts(f"Result {ip}: loss={loss}%  min/avg/max={min_}/{avg_}/{max_} ms")
        else:
            ts(f"Result {ip}: loss={loss}%  (RTT not available)")
    except FileNotFoundError:
        ts("ERROR: ping binary not found")
    except subprocess.TimeoutExpired:
        ts(f"ERROR: ping to {ip} timed-out after {PING_TIMEOUT}s")
    except Exception:
        ts(f"Unexpected exception while pinging {ip}")
        traceback.print_exc()

def main() -> None:
    ts("=== SCRIPT EXECUTION STARTED ===")
    while True:
        try:
            cycle_start = dt.datetime.now()
            ips = get_node_ips()
            if not ips:
                ts("No IPs found, skipping ping cycle.")
            else:
                for ip in ips:
                    ping_one(ip)
            elapsed = (dt.datetime.now() - cycle_start).total_seconds()
            ts(f"Cycle finished – sleeping {SLEEP_CYCLE}s\n")
            time.sleep(max(0, SLEEP_CYCLE - elapsed))
        except Exception as e:
            ts(f"ERROR in main loop: {e}")
            traceback.print_exc()
            time.sleep(30)  # Sleep and retry on error

if __name__ == "__main__":
    main()