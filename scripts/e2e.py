import json
import urllib.request
import urllib.error

B = "http://localhost:8080/api/v1"


def call(method, path, body=None, token=None):
    req = urllib.request.Request(B + path, method=method,
        data=json.dumps(body).encode() if body is not None else None,
        headers={"Content-Type": "application/json", **({"Authorization": f"Bearer {token}"} if token else {})})
    try:
        with urllib.request.urlopen(req) as r:
            return r.status, json.loads(r.read() or b"null")
    except urllib.error.HTTPError as e:
        try:
            return e.code, json.loads(e.read() or b"null")
        except Exception:
            return e.code, None


S = "e2e_visor"
s, reg = call("POST", "/auth/register", {"email": f"{S}@example.com", "username": S, "password": "trenchcoat99"})
assert s == 201, (s, reg); tMAIN = reg["token"]; print("1 register 201 ok")
s, me = call("GET", "/auth/me", token=tMAIN)
assert s == 200 and me["profile"]["xp"] == 0, (s, me); print("2 me xp=0 ok")
s, _ = call("POST", "/auth/logout", token=tMAIN)
assert s == 200
s, _ = call("GET", "/auth/me", token=tMAIN)
assert s == 401, s; print("3 logout + revoked token 401 ok")
s, log = call("POST", "/auth/login", {"email": f"{S}@example.com", "password": "trenchcoat99"})
assert s == 200; tA = log["token"]; print("4 login 200 ok")
s, st1 = call("POST", "/cases/the_last_train/start", token=tA)
assert s == 201, (s, st1)
s, st2 = call("POST", "/cases/the_last_train/start", token=tA)
assert s == 200; print("5 start idempotent 201/200 ok")
s, sea = call("POST", "/cases/the_last_train/actions", {"action": "search", "location_id": "berth_c"}, tA)
assert s == 200 and len(sea["found"]) == 3, (s, sea); print("6 search berth_c: 3 found")
s, sea2 = call("POST", "/cases/the_last_train/actions", {"action": "search", "location_id": "berth_c"}, tA)
assert s == 200 and sea2["found"] == [] and len(sea2["already_known"]) == 3; print("7 re-search idempotent ok")
s, inter = call("POST", "/cases/the_last_train/interrogate", {"suspect_id": "edwin_marsh", "topic_id": "t_where"}, tA)
assert s == 200 and "midnight" in inter["response"], (s, inter); print("8 interrogate ok")
s, bad = call("POST", "/cases/the_last_train/present-evidence", {"suspect_id": "tommy_beck", "evidence_id": "timetable"}, tA)
assert s == 422 and bad["error"]["code"] == "unproven", (s, bad); print("9 present undiscovered -> 422 unproven ok")
for loc, n in [("dining_car", 2), ("vestibule_car3", 2), ("observation_deck", 1)]:
    s, r = call("POST", "/cases/the_last_train/actions", {"action": "search", "location_id": loc}, tA)
    assert s == 200 and len(r["found"]) == n, (loc, s, r)
print("10 all locations swept: 8 exhibits")
s, p1 = call("POST", "/cases/the_last_train/present-evidence", {"suspect_id": "tommy_beck", "evidence_id": "bond_stub"}, tA)
assert s == 200 and p1["contradiction"] and p1["unlocked_testimony"]["id"] == "tommy_timeline", (s, p1)
s, p2 = call("POST", "/cases/the_last_train/present-evidence", {"suspect_id": "vivienne_hart", "evidence_id": "matchbook"}, tA)
assert s == 200 and p2["contradiction"] and p2["unlocked_testimony"]["id"] == "vivienne_sighting", (s, p2)
print("11 both contradictions unlock testimonies ok")
s, state = call("GET", "/cases/the_last_train/state", token=tA)
assert s == 200 and len(state["discovered_evidence"]) == 8 and len(state["testimonies"]) == 2
print("12 resume-after-close: state intact (8 exhibits, 2 statements) ok")
s, reg2 = call("POST", "/auth/register", {"email": "e2e_wrong@example.com", "username": "e2e_wrong", "password": "trenchcoat99"})
tB = reg2["token"]
call("POST", "/cases/the_last_train/start", token=tB)
call("POST", "/cases/the_last_train/actions", {"action": "search", "location_id": "berth_c"}, tB)
s, w = call("POST", "/cases/the_last_train/accusations", {"suspect_id": "tommy_beck"}, tB)
assert s == 201 and w["result"] == "wrong" and w["xp_awarded"] == 25, (s, w)
print("13 wrong accusation recorded, +25 XP ok")
s, acc = call("POST", "/cases/the_last_train/accusations", {"suspect_id": "edwin_marsh"}, tA)
assert s == 201 and acc["result"] == "guilty" and acc["score"] == 1600 and acc["xp_awarded"] == 160, (s, acc)
print("14 GUILTY: score 1600, +160 XP ok")
s, acc2 = call("POST", "/cases/the_last_train/accusations", {"suspect_id": "edwin_marsh"}, tA)
assert s == 200 and acc2["xp_awarded"] == 160
s, me2 = call("GET", "/auth/me", token=tA)
assert me2["profile"]["xp"] == 160, me2["profile"]; print("15 replay idempotent, XP still 160 ok")
s, log2 = call("POST", "/auth/login", {"email": f"{S}@example.com", "password": "trenchcoat99"})
s, cases = call("GET", "/cases", token=log2["token"])
row = [c for c in cases if c["id"] == "the_last_train"][0]
assert row["status"] == "completed" and row["correct"] is True, row
s, ach = call("GET", "/player/achievements", token=log2["token"])
unlocked = [a["id"] for a in ach if a["unlocked_at"]]
assert set(unlocked) == {"started_first_case", "found_first_evidence", "broke_alibi", "solved_last_train", "perfect_deduction"}, unlocked
s, lb = call("GET", "/leaderboard", token=log2["token"])
assert any(e["username"] == S for e in lb)
print("16 relogin: completed/correct persisted, 5/5 achievements, leaderboard ok")
print("E2E PASS")
