# Case file: The Tunis Ledger (v2, second case)

Fictional 1940s noir set in Tunis. All persons, companies, and places are
invented; any resemblance is coincidence. Requires closing The Last Train
first (`requires_case` in JSON). Same mechanics, harder read: testimonies
contradict each other until the ledger page is found.

## Setup

Hôtel Majestic, Tunis marina, November 1947. At 7:40 AM the chambermaid finds
**Henri Aubert** (60), owner of the hotel, dead at his office desk, cold coffee
at his elbow. The house doctor writes "heart" and reaches for his hat. But the
harbor ledger — the book where Aubert recorded every crate that crossed his
quay — lies open with a page razored out, and forty thousand francs of
"harbor money" are gone from the office safe. Captain Kowalski loans you to
the harbor police. You get one day before the shipping line buries it.

## Culprit

**Marco Silvestri** (45), Italian cargo clerk for the fictional Ligne
Sfaxienne. For a year he has forged bills of lading, walking condemned cargo
through Aubert's quay for a cut. Aubert finally compared the manifests against
the ledger and summoned Silvestri to the office at midnight to be dismissed —
and exposed. Silvestri dosed Aubert's anisette with sleeping drops, waited
until the old man folded over the desk, razored the ledger page that named
him, emptied the safe to look like robbery, and walked out through the
service stair with the night watchman's spare key, which he had waxed an
impression of a week before.

## Motive

Exposure + money. The forged manifests (in his own hand, misspelled the same
way twice) plus the missing ledger page would have ended him; the safe money
was meant to buy passage to Palermo.

## Timeline (all times stated in-world)

- 9:00 PM — Silvestri signs the hotel dispensary book for sleeping drops
  "for the crossing" (he sails — or claims to — on Friday).
- 11:30 PM — Hotel bar. While Aubert argues with Yasmine about her contract,
  Silvestri tops up Aubert's anisette from his own flask and doses it.
- Midnight — Aubert dismisses everyone, retires to the office with the glass.
- 12:40 AM — Yasmine, sleepless at the Cinéma L'Étoile's late show let-out
  across the square, sees Silvestri on the service stair — jacket on, ledger
  folio under his arm. She tells no one: a singer seen loitering near a dead
  man's stair loses her contract.
- 1:10 AM — Silvestri lets himself into the office with the waxed spare key,
  razors the page, empties the safe.
- 5:00 AM — Watchman Bilal Haddad makes his rounds, finds the office door
  ajar and Aubert cold. He pockets banknotes scattered on the floor, straightens
  the desk to hide his shaking, and reports "found it locked" at 7:40 with the maid.
- 7:40 AM — Body reported. Bilal has already spent one banknote at the quay café.

## Suspects

### Marco Silvestri, 45 — cargo clerk, Ligne Sfaxienne. THE KILLER.
Voice: elaborate courtesy, long sentences, quotes regulations; goes flat and
short when the manifests come up. Never raises his voice. Never answers about
Friday's sailing.
Alibi (stated): "In my room above the chandlery from eleven. Seasick in
advance — I am a poor sailor for a sailing man."
Truth: dosed the anisette at 11:30 PM, entered at 1:10 AM with the waxed key.
Breaks on: dispensary book (had the drops), forged manifests (motive + hand),
service-stair testimony (seen), waxed key (entry).

### Yasmine Ben Ammar, 34 — hotel bar singer. RED HERRING.
Voice: crisp patter, half French, half Tunis Arabic courtesy; proverbs when
nervous, steel when cornered. Calls the detective "monsieur l'inspecteur"
until she is angry, then just "you".
Alibi (stated): "In bed by midnight, alone, asleep. Ask my pillow."
Truth: sleepless, at the cinema's late show, then smoking in the square till
1 AM; saw Silvestri on the stair at 12:40 with the folio. She argued with
Aubert at 11 PM about money he "lent" her contract — she hated him a little
and feared looking guilty a lot.
Breaks on: the cinema ticket stub (places her out) → admits the square and,
pressed, gives up the stair sighting.

### Bilal Haddad, 28 — night watchman. WITNESS / petty thief.
Voice: short formal sentences, "sir" to everyone, repeats the rulebook when
frightened. Lies once, badly, about the door; confesses everything else twice.
Alibi (stated): "Rounds at five, sir. Door locked, sir. Reported at seven-forty."
Truth: door AJAR at 5 AM; took scattered banknotes; straightened the desk;
spent one note at the café. His defense: "A dead man needs no francs, sir —
my sister does."
Breaks on: the café banknote (serial matches the safe's recorded run) →
confesses and gives the timeline anchor: door ajar (not locked) at 5 AM, safe
already empty, desk already straightened — so entry was by key and death
preceded 5 AM.

## Evidence (8 items)

| ID | Name | Found | Proves / contradicts |
|----|------|-------|----------------------|
| `anisette` | Dosed anisette glass | Office | Sleeping drops in the dregs → drugged before death; contradicts heart failure. |
| `doctor_slip` | Doctor's slip | Office | Pupil constriction noted, then crossed out → the doctor saw drugging signs and buried them. |
| `ledger` | Razored ledger | Office | Page naming Silvestri's cargo cut out with a keen blade → robbery of proof, not money-first. |
| `manifests` | Forged manifests | Quay (harbormaster's hut) | Two bills of lading in Silvestri's hand, same misspelling twice → motive + authorship. |
| `dispensary` | Dispensary book | Bar (hotel dispensary drawer) | Silvestri signed out sleeping drops at 9 PM → had the drug. |
| `stub` | Cinema ticket stub | Cinéma L'Étoile | Torn stub, Yasmine's seat row, lipstick print → she was out; presenting it unlocks her testimony. |
| `banknote` | Café banknote | Quay (café till) | Safe-run serial on a note Bilal spent at dawn → Bilal stole post-mortem; unlocks his timeline. |
| `wax_key` | Waxed key impression | Service stair (loose step) | Key impression in sealing wax + Silvestri's seal ring mark → entry by copied key. |

## Contradiction chain

1. `anisette` + `doctor_slip` ⇒ drugged, staged as heart failure. Murder.
2. `ledger` ⇒ the missing page names someone — robbery of proof. `manifests`
   ⇒ Silvestri forged cargo papers; the page would have hanged him. Motive.
3. `dispensary` ⇒ Silvestri held the exact drops in Aubert's glass.
4. `wax_key` + Bilal's testimony (unlocked by `banknote`) ⇒ entry by copied
   service-stair key; door ajar at 5 AM; safe already empty. Death before 5 AM.
5. `stub` ⇒ Yasmine lied about bed ⇒ her testimony puts Silvestri on the
   service stair at 12:40 AM with the folio.
6. Conclusion: only Silvestri had motive (exposure + passage money), means
   (drops + copied key), opportunity (awake, on the stair in the death
   window), and a lied alibi. **Accuse Marco Silvestri.**

## Scoring
Same table as The Last Train (base 1000 + 50/exhibit + 100/testimony;
unproven 100 XP, wrong 25 XP). New achievement: `solved_tunis_ledger`.
