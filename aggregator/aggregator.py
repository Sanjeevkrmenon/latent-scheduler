#!/usr/bin/env python3
"""
Aggregator for latency mesh: merges all /latency/*.json (from DaemonSet nodes)
into a cluster-wide /latency/cluster-latency.json file for the custom scheduler.
"""

import os
import json
import glob
import time

LATENCY_DIR = "/latency"
OUTPUT = os.path.join(LATENCY_DIR, "cluster-latency.json")

def main():
    mesh = {}
    files = [f for f in glob.glob(f"{LATENCY_DIR}/*.json") if not f.endswith("cluster-latency.json")]
    for path in files:
        try:
            with open(path) as f:
                data = json.load(f)
            mesh.update(data)
        except Exception as e:
            print(f"Warning: Could not read {path}: {e}", flush=True)
    # Write atomically (write to .tmp, then move)
    tmp = OUTPUT + ".tmp"
    with open(tmp, "w") as f:
        json.dump(mesh, f, indent=2)
    os.replace(tmp, OUTPUT)
    print(f"[{time.strftime('%F %T')}] Aggregated {len(files)} files into {OUTPUT} ({len(mesh)} nodes)", flush=True)

if __name__ == "__main__":
    main()