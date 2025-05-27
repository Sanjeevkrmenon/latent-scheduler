#!/usr/bin/env python3

import subprocess
import time
import re
import datetime as dt
import traceback
import os
import socket
import json

# Configuration (use env vars or defaults)
PING_COUNT   = int(os.getenv("PING_COUNT", "4"))
SLEEP_CYCLE  = int(os.getenv("SLEEP_INTERVAL", "60"))
PING_TIMEOUT = int(os.getenv("PING_TIMEOUT", "15"))
KUBECTL_TIMEOUT = int(os.getenv("KUBECTL_TIMEOUT", "10"))

# Use true node name (from DaemonSet env)
NODE_NAME = os.environ.get("NODE_NAME", socket.gethostname())

def ts(msg: str) -> None:
    print(f"[{dt.datetime.now():%Y-%m-%d %H:%M:%S}][{NODE_NAME}] {msg}", flush=True)

def get_node_ips() -> dict:
    ts("Fetching node IPs with kubectl…")
    try:
        out = subprocess.run(
            [
                "kubectl", "get", "nodes",
                "-o",
                'jsonpath={range .items[*]}{.metadata.name}={.status.addresses[?(@.type=="InternalIP")].address}{" "}{end}'
            ],
            capture_output=True,
            text=True,
            check=True,
            timeout=KUBECTL_TIMEOUT,
        )
        ip_pairs = out.stdout.strip().split()
        nodes_ips = {}
        for pair in ip_pairs:
            if "=" in pair:
                node, ip = pair.split("=")
                nodes_ips[node] = ip
        if not nodes_ips:
            ts("WARNING: No node IPs found by kubectl!")
        else:
            ts(f"Node IPs: {nodes_ips}")
        return nodes_ips
    except FileNotFoundError:
        ts("ERROR: kubectl not found in PATH")
        return {}
    except subprocess.TimeoutExpired:
        ts("ERROR: kubectl timed out")
        return {}
    except subprocess.CalledProcessError as e:
        ts(f"ERROR: kubectl failed rc={e.returncode}  stderr={e.stderr.strip()}")
        return {}
    except Exception:
        ts("Unexpected error while fetching IPs")
        traceback.print_exc()
        return {}

rtt_re = re.compile(r"rtt min/avg/max/\S+\s*=\s*([0-9.]+)/([0-9.]+)/([0-9.]+)")

def ping_one(ip: str) -> tuple:
    try:
        out = subprocess.run(
            ["ping", "-c", str(PING_COUNT), "-W", str(PING_TIMEOUT), ip],
            capture_output=True,
            text=True,
            timeout=PING_TIMEOUT + 2,
            check=False,
        )
        all_text = out.stdout + out.stderr
        loss_m = re.search(r"(\d+)% packet loss", all_text)
        rtt_m  = rtt_re.search(out.stdout)
        loss = int(loss_m.group(1)) if loss_m else None
        if rtt_m:
            min_, avg_, max_ = map(float, rtt_m.groups())
            return (loss, min_, avg_, max_)
        else:
            return (loss, None, None, None)
    except FileNotFoundError:
        ts("ERROR: ping binary not found")
        return (None, None, None, None)
    except subprocess.TimeoutExpired:
        ts(f"ERROR: ping to {ip} timed-out after {PING_TIMEOUT}s")
        return (None, None, None, None)
    except Exception:
        ts(f"Unexpected exception while pinging {ip}")
        traceback.print_exc()
        return (None, None, None, None)

def main() -> None:
    ts("=== SCRIPT EXECUTION STARTED ===")
    while True:
        try:
            cycle_start = dt.datetime.now()
            nodes_ips = get_node_ips()
            if not nodes_ips:
                ts("No nodes found, skipping ping cycle.")
                time.sleep(SLEEP_CYCLE)
                continue

            my_row = {}
            for dst_node, dst_ip in nodes_ips.items():
                if NODE_NAME == dst_node:
                    my_row[dst_node] = 0.0
                else:
                    loss, min_rtt, avg_rtt, max_rtt = ping_one(dst_ip)
                    if avg_rtt is not None:
                        my_row[dst_node] = avg_rtt
                        ts(f"Ping {NODE_NAME} -> {dst_node} ({dst_ip}): loss={loss}% avg RTT={avg_rtt} ms")
                    else:
                        my_row[dst_node] = None
                        ts(f"Ping {NODE_NAME} -> {dst_node} ({dst_ip}): loss={loss}%, RTT not available")

            # Write just this node's row (by node name)
            try:
                os.makedirs("/latency", exist_ok=True)
                path = f"/latency/{NODE_NAME}.json"
                with open(path, "w") as f:
                    json.dump({NODE_NAME: my_row}, f, indent=2)
                ts(f"Latency row saved to {path}")
            except Exception as e:
                ts(f"ERROR writing latency JSON: {e}")

            elapsed = (dt.datetime.now() - cycle_start).total_seconds()
            ts(f"Cycle finished – sleeping {SLEEP_CYCLE}s\n")
            time.sleep(max(0, SLEEP_CYCLE - elapsed))

        except Exception as e:
            ts(f"ERROR in main loop: {e}")
            traceback.print_exc()
            time.sleep(30)

if __name__ == "__main__":
    main()