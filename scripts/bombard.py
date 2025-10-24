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
    """Generate vibrant random colors"""
    hue = int((t * 0.05 + x / max_coord * 360 + y / max_coord * 180) % 360)
    sat = 85 + int(15 * math.sin(t * 0.01 + x / max_coord * 3.14))
    light = 45 + int(20 * math.cos(t * 0.008 - y / max_coord * 2))
    return [hue, sat, light]


def get_dynamic_properties(step, x, y, max_coord):
    """Generate varying radius and transparency"""
    t = time.time()

    # Distance from center for radial effects
    cx, cy = max_coord / 2, max_coord / 2
    dist_from_center = math.sqrt((x - cx) ** 2 + (y - cy) ** 2) / (max_coord / 2)

    # Pulsing size
    radius = 20 + 15 * math.sin(t * 2 + step * 0.05)
    transparency = 0.4 + 0.3 * math.cos(t * 1.5)
    time_to_grow = int(400 + 200 * math.sin(step * 0.1))

    # Clamp values
    radius = max(10, min(60, radius))
    transparency = max(0.2, min(0.95, transparency))
    time_to_grow = max(200, min(1000, time_to_grow))

    return radius, transparency, time_to_grow


def make_payload(x, y, color, radius, transparency, time_to_grow):
    return {
        "color": color,
        "radius": radius,
        "transparency": transparency,
        "time_to_grow": time_to_grow,
    }


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


def run_bombard(host, port, duration, dry_run, concurrency):
    max_coord = 511
    total = 0
    start = time.time()

    rate = 1000  # 1000 requests/second base
    interval = 1.0 / rate
    points_per_iteration = 5  # 5 points per iteration = ~5000 req/s

    executor = ThreadPoolExecutor(max_workers=concurrency)
    futures = []
    step = 0

    print(f"🔥 Starting random color bombardment")
    print(f"⚡ Target rate: ~{rate * points_per_iteration} req/s")
    print(f"🌈 Dynamic properties enabled")

    while True:
        now = time.time()
        if duration and (now - start) >= duration:
            break

        # Send multiple random points per iteration
        for _ in range(points_per_iteration):
            x = random.randint(0, max_coord)
            y = random.randint(0, max_coord)

            color = hsl_from_params(now * 1000, x, y, max_coord)
            radius, transparency, time_to_grow = get_dynamic_properties(
                step, x, y, max_coord
            )

            payload = make_payload(x, y, color, radius, transparency, time_to_grow)
            url = f"http://{host}:{port}/api/{x}/{y}"

            if dry_run:
                print(
                    f"[DRY] POST {url} -> r:{radius:.1f} t:{transparency:.2f} color:{color}"
                )
            else:
                futures.append(executor.submit(send_post, url, payload))

            total += 1
            step += 1

        if not dry_run and total % 500 == 0:
            print(f"📊 Sent {total} points... ({total/(time.time()-start):.0f} req/s)")

        elapsed = time.time() - now
        sleep_for = interval - elapsed
        if sleep_for > 0:
            time.sleep(sleep_for)

    if not dry_run:
        print(f"⏳ Waiting for requests to complete...")
        for f in as_completed(futures):
            status, resp = f.result()
            if status is None and total <= 10:
                print(f"❌ Error: {resp[:200] if isinstance(resp, str) else resp}")

    executor.shutdown(wait=True)
    print(f"✨ Complete! Sent {total} points in {time.time()-start:.1f}s")


def build_arg_parser():
    p = argparse.ArgumentParser(
        description="🔥 Random Color Bombardment - Stress test your Elixir API",
    )
    p.add_argument("--host", default="127.0.0.1", help="Target host")
    p.add_argument("--port", type=int, default=4000, help="Target port")
    p.add_argument(
        "--duration", type=float, default=0, help="Duration in seconds (0 = infinite)"
    )
    p.add_argument(
        "--dry-run", action="store_true", help="Print requests without sending"
    )
    p.add_argument(
        "--concurrency",
        type=int,
        default=200,
        help="Number of concurrent requests (default 200 for MAXIMUM throughput)",
    )
    return p


if __name__ == "__main__":
    parser = build_arg_parser()
    args = parser.parse_args()
    run_bombard(args.host, args.port, args.duration, args.dry_run, args.concurrency)
