import json, subprocess, sys, time
bin_ = sys.argv[1]
only = sys.argv[2:]
def run(*a, timeout=1200):
    try:
        p = subprocess.run([bin_, *a, "--json"], capture_output=True, text=True, timeout=timeout, stdin=subprocess.DEVNULL)
    except subprocess.TimeoutExpired:
        return None, "TIMEOUT"
    try:
        return json.loads(p.stdout), p.stderr
    except ValueError:
        return None, p.stdout[-300:] + p.stderr[-300:]
def installed():
    d, _ = run("tools", "list")
    return {x["key"] for x in d if x["installed"]}
before = installed()
all_ = [x["key"] for x in run("tools", "list")[0]]
targets = only or [k for k in all_ if k not in before]
print(f"{len(all_)} no catálogo, {len(before)} já instaladas, {len(targets)} a testar", flush=True)
ok, bad = [], []
for k in targets:
    t = time.time()
    d, err = run("tools", "install", k)
    r = d["results"][0] if d else {"action": "failed", "error": err}
    if r["action"] != "installed":
        msg = (r.get("error") or "")[-260:].replace("\n", " | ")
        print(f"✘ {k:15} install: {msg}", flush=True)
        bad.append(k)
        run("tools", "uninstall", k, "--yes")
        continue
    ver = r.get("version", "")[:50]
    d, err = run("tools", "uninstall", k, "--yes")
    u = d["results"][0] if d else {"action": "failed", "error": err}
    if u["action"] not in ("removed",):
        print(f"✘ {k:15} uninstall [{u['action']}]: {(u.get('error') or u.get('version') or '')[-200:]}", flush=True)
        bad.append(k)
        continue
    print(f"✔ {k:15} {int(time.time()-t):4}s  {ver}", flush=True)
    ok.append(k)
after = installed()
print(f"\nOK {len(ok)}, FALHAS {len(bad)}: {' '.join(bad)}")
print("sumiram:", sorted(before - after), " sobraram:", sorted(after - before))
