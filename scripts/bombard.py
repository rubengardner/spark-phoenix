#!/usr/bin/env python3
import argparse
import math
import time
import json
import random
from concurrent.futures import ThreadPoolExecutor, as_completed
try:
    import requests
    _HAS_REQUESTS = True
except Exception:
    import urllib.request
    _HAS_REQUESTS = False

def hsl_from_params(t, x, y, max_coord):
    hue = int((math.sin(t*0.002 + x/max_coord*6.2831 - y/max_coord*3.1415) + 1) * 180) % 360
    sat = 80
    light = int((0.5 + 0.5*math.cos(t*0.001 + x/max_coord - y/max_coord)) * 50) + 25
    return [hue, sat, light]

def lissajous_coords(step, a=3, b=2, delta=math.pi/2, max_coord=511):
    t = step/100.0
    x = int((math.sin(a*t + delta) + 1) * 0.5 * max_coord)
    y = int((math.sin(b*t) + 1) * 0.5 * max_coord)
    return x, y

def spiral_coords(step, max_coord=511):
    theta = step * 0.15
    r = 0.5 * math.sqrt(step)
    cx = cy = max_coord/2
    x = int(cx + r * math.cos(theta) * max_coord/512)
    y = int(cy + r * math.sin(theta) * max_coord/512)
    x = max(0, min(max_coord, x))
    y = max(0, min(max_coord, y))
    return x, y

def golden_spiral(step, max_coord=511):
    phi = (1 + 5**0.5)/2
    theta = step * 2.4
    r = (theta / (2*math.pi))**0.5
    cx = cy = max_coord/2
    x = int(cx + r * math.cos(theta) * max_coord/6)
    y = int(cy + r * math.sin(theta) * max_coord/6)
    x = max(0, min(max_coord, x))
    y = max(0, min(max_coord, y))
    return x, y

def random_walk(step, max_coord=511):
    x = random.randint(0, max_coord)
    y = random.randint(0, max_coord)
    return x, y

PATTERNS = {
    "lissajous": lissajous_coords,
    "spiral": spiral_coords,
    "golden": golden_spiral,
    "random": random_walk
}

def make_payload(x, y, color, radius, transparency, time_to_grow):
    return {"color": color, "radius": radius, "transparency": transparency, "time_to_grow": time_to_grow}

def send_post(url, payload, timeout=2.0):
    data = json.dumps(payload).encode("utf-8")
    headers = {"Content-Type": "application/json"}
    if _HAS_REQUESTS:
        try:
            r = requests.post(url, json=payload, timeout=timeout)
            return r.status_code, r.text
        except Exception as e:
            return None, str(e)
    else:
        req = urllib.request.Request(url, data=data, headers=headers, method="POST")
        try:
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                return resp.getcode(), resp.read().decode("utf-8")
        except Exception as e:
            return None, str(e)

def run_bombard(host, port, pattern, rate, duration, count, radius, transparency, time_to_grow, dry_run, concurrency):
    max_coord = 511
    total = 0
    start = time.time()
    interval = 1.0 / max(rate, 1)
    executor = ThreadPoolExecutor(max_workers=concurrency)
    futures = []
    step = 0
    while True:
        now = time.time()
        if duration and (now - start) >= duration:
            break
        if count and total >= count:
            break
        x, y = PATTERNS.get(pattern, random_walk)(step, max_coord=max_coord)
        color = hsl_from_params(now*1000, x, y, max_coord)
        payload = make_payload(x, y, color, radius, transparency, time_to_grow)
        url = f"http://{host}:{port}/api/{x}/{y}"
        if dry_run:
            print(f"[DRY] POST {url} -> {payload}")
        else:
            futures.append(executor.submit(send_post, url, payload))
        total += 1
        step += 1
        elapsed = time.time() - now
        sleep_for = interval - elapsed
        if sleep_for > 0:
            time.sleep(sleep_for)
    if not dry_run:
        for f in as_completed(futures):
            status, resp = f.result()
            print("RESULT", status, (resp[:200] if isinstance(resp, str) else resp))
    executor.shutdown(wait=True)

def build_arg_parser():
    p = argparse.ArgumentParser()
    p.add_argument("--host", default="127.0.0.1")
    p.add_argument("--port", type=int, default=4000)
    p.add_argument("--pattern", choices=list(PATTERNS.keys()), default="lissajous")
    p.add_argument("--rate", type=int, default=200)
    p.add_argument("--duration", type=float, default=0)
    p.add_argument("--count", type=int, default=0)
    p.add_argument("--radius", type=float, default=30.0)
    p.add_argument("--transparency", type=float, default=0.9)
    p.add_argument("--time_to_grow", type=int, default=500)
    p.add_argument("--dry-run", action="store_true")
    p.add_argument("--concurrency", type=int, default=50)
    return p

if __name__ == "__main__":
    parser = build_arg_parser()
    args = parser.parse_args()
    run_bombard(args.host, args.port, args.pattern, args.rate, args.duration, args.count, args.radius, args.transparency, args.time_to_grow, args.dry_run, args.concurrency)
