# Case file: The Last Train (v1, single case)

Fictional 1940s noir. All persons, companies, and places are invented. This doc
is the source of truth for the deduction; `backend/content/cases/the_last_train.json`
is the machine-readable twin. A Go test enforces that every ID referenced here
exists there and that the correct accusation is deducible from the evidence.

## Setup

Night mail train, the *Meridian Limited*, out of the capital bound for Graywater.
October 1947. At 6:10 AM, as the train brakes for Graywater Station, sleeping-car
porter Tommy Beck finds **Silas Vane** (58), paymaster of the Meridian Freight
Company, dead in his bunk in Car 3, Berth C. The company doctor rides along and
calls it heart failure. The railroad dicks want it buried before the papers get
it. The player — a Graywater police detective loaned to the railway — gets one
morning aboard the held train to find the truth.

Vane carried the quarterly payroll satchel: $40,000 in Graywater rail bonds.
The satchel is found slit open under his bunk. Empty.

## Culprit

**Edwin Marsh** (51), Vane's junior partner. He has been skimming freight receipts
to cover cotton-margin losses. Vane finished a private audit on the trip and
carried the audit letter in his coat: exposure at the Graywater board meeting,
ruin, prison. Marsh dosed Vane's flask, waited, entered with his company master
key during the 3:40 AM crew-inspection stop, smothered the drugged man with his
own pillow, took the bonds to cover the shortfall, and went back to bed.

## Motive

Money + exposure. The audit letter (briefing exhibit) proves Vane knew. Marsh's
debts are documented in his dossier. No crime of passion, no hired hand.

## Timeline (all times stated in-world)

- 11:00 PM — Marsh signs the dining-car medicine chest's chloral hydrate bottle
  out "for his nerves" (steward's log).
- Midnight — Dining car. Marsh pours Vane's rye from his own flask top-up and
  doses it with the chloral while Vane argues with the steward about the audit.
- 1:00 AM — Vane retires to Berth C, flask in hand, already drowsy.
- ~2:00 AM — Vivienne Hart steps onto the observation deck for air (and a
  cigarette she won't admit to); the drummer from the dining-car trio keeps her
  company for ten minutes, then leaves.
- 3:30 AM — Vivienne, still on the deck, sees Marsh pass the Car 3 vestibule
  window toward the berths, fully dressed, carrying his overcoat.
- 3:40 AM — Crew-inspection stop (timetable). Under cover of the brake noise,
  Marsh latches Berth C open with master key No. 6, smothers Vane, slits the
  satchel, takes the bonds.
- 5:50 AM — Tommy brings Vane's morning coffee. Door latched, no answer. Thinks
  him asleep; leaves the cup.
- 6:10 AM — Tommy returns with the conductor. Body found. Tommy, panicking at
  the sight of the slit satchel, pockets one loose bond from the floor — and
  tears it on the vestibule latch fleeing.

## Suspects

### Edwin Marsh, 51 — junior partner, Meridian Freight. THE KILLER.
Voice: clipped boardroom sentences. Five words where ten would do. Dodges
anything about money or the night; answers questions about the company like a
man reading minutes.
Alibi (stated): "Retired at midnight. Slept straight through. Woke at six."
Truth: dosed the flask at midnight, killed at 3:40 AM.
Breaks on: timetable (awake), medicine log (had the drug), key registry (had
entry), Vivienne's testimony (seen at the scene).

### Vivienne Hart, 38 — lounge singer, dining-car trio's vocalist. RED HERRING.
Voice: smoky, singsong, calls everyone "sugar". Deflects with flirtation; goes
cold and precise when cornered.
Alibi (stated): "Never left the dining car all night, sugar. Ask the boys."
Truth: was on the observation deck ~2–3:30 AM (ashamed: she was crying over a
letter from her estranged sister, and she saw Marsh but feared getting involved).
She hated Vane — fifteen years ago Vane's testimony bankrupted her father's
orchard in a freight-rate suit — but she never touched him.
Breaks on: the observation-deck matchbook (places her off the dining car) →
she admits the deck and, pressed, gives up what she saw: Marsh at Car 3 at 3:30.

### Tommy Beck, 29 — sleeping-car porter. WITNESS / petty thief.
Voice: run-on nervous sentences, "sir" every other word, apologizes to furniture.
Alibi (stated): "Polishing brass in Car 2 at four, sir, then coffee rounds."
Truth: mostly true — but at 6:10 he stole one bond from the scene and lied about
"never touching nothing".
Breaks on: the torn bond stub (matches his uniform tear) → confesses the theft
and gives the timeline anchor: Berth C latched (not bolted) at 5:50, satchel
already slit — so entry was by key, and death preceded 5:50.

## Evidence (8 items)

| ID | Name | Found | Proves / contradicts |
|----|------|-------|----------------------|
| `flask` | Vane's silver flask | Berth C | Rye laced with chloral hydrate → drugged before death; contradicts natural death. |
| `pillow_note` | Coroner's slip | Berth C | Down fibers in the airways, pillow misshapen → smothered; contradicts heart failure. |
| `satchel` | Slit payroll satchel | Berth C | $40,000 in bonds gone → robbery motive; contradicts "nothing taken". |
| `timetable` | Circled timetable | Dining car (Marsh's overcoat on his chair) | 3:40 AM inspection stop circled in Marsh's hand → awake and planning; contradicts "slept straight through". |
| `medicine_log` | Steward's medicine log | Dining car | Marsh signed out the chloral at 11 PM → had the drug. |
| `bond_stub` | Torn bond half | Car 3 vestibule latch | Matches Tommy's uniform tear → Tommy stole post-mortem (theft, not murder); proves satchel opened before 6:10. |
| `key_registry` | Conductor's key log | Car 3 vestibule (conductor's cubby) | Marsh signed out master key No. 6 in the capital → means of key entry; with Tommy's "latched, not bolted", entry is explained. |
| `matchbook` | Observation-deck matchbook | Observation deck | Lipstick-marked, Vivienne's shade → she left the dining car; presenting it unlocks her testimony (Marsh at Car 3, 3:30 AM). |

## Contradiction chain (the deduction)

1. `flask` + `pillow_note` ⇒ Vane was drugged, then smothered. Not heart failure. Murder.
2. `satchel` ⇒ robbery: $40,000 motive. Briefing's audit letter ⇒ Marsh is the one Vane could ruin.
3. `timetable` ⇒ Marsh lied about sleeping through the night.
4. `medicine_log` ⇒ Marsh held the exact drug in Vane's flask.
5. `key_registry` + Tommy's testimony (unlocked by `bond_stub`) ⇒ entry by company master key; Marsh held one; door latched-not-bolted fits key entry, not force.
6. `matchbook` ⇒ Vivienne lied about the dining car ⇒ her testimony puts Marsh at Car 3 at 3:30 AM, fully dressed.
7. Conclusion: only Marsh had motive (audit ruin + debts), means (drug + key), opportunity (awake, at the berth in the death window), and a lied alibi. **Accuse Edwin Marsh.**

Fair-play check: every link above is discoverable in-world; the Go test asserts the
required-evidence set `{flask, pillow_note, satchel, timetable, medicine_log,
key_registry, bond_stub}` plus unlocked testimonies `{tommy_timeline, vivienne_sighting}`
is reachable, and that accusing `edwin_marsh` with that set returns guilty while
any other suspect returns not-guilty.

## Scoring (implemented server-side from `accusation_rules` in JSON)

- Correct accusation with full required set: 1000 base + 50 per evidence item
  discovered (max 8) + 100 per unlocked testimony (max 2) = up to 1600.
- Correct suspect but incomplete proof: recorded as failed ("unproven") — 100 XP
  consolation, case closed.
- Wrong suspect: recorded as failed — 25 XP, case closed.
- XP = score / 10, integer. Awarded exactly once via unique `(user_id, case_id)`.
