# injustice-gaus-nex

A NEX server for **Injustice: Gods Among Us** on the Wii U, so the online mode
works again now that Nintendo's servers for it are gone. It's built on the
[Pretendo](https://github.com/PretendoNetwork) libraries (`nex-go`,
`nex-protocols-go`, `nex-protocols-common-go`) and written in Go.

The game talks to a single NEX endpoint for login, matchmaking (ranked and
player matches, King of the Hill, survivor), leaderboards and DataStore. All of
that is handled here, backed by PostgreSQL.

## Where it stands

Tested on real consoles (EUR and USA):

- login, ticket, secure connection
- the profile bootstrap the game does on first launch
- entering the lobby
- matchmaking: two consoles find each other and end up in the same session,
  including across regions

Still being worked on: what happens right after two players are matched (the
game reports "the game session is no longer available" in some cases). If you
have a capture of that moment, I'd love to see it.

Leaderboard writes (`UploadScore` / `UploadCommonData`) haven't been seen from
a real match yet, so treat ranking as untested.

## Numbers you'll want

| | |
|---|---|
| access key | `65e9f4d6` |
| NEX version | 3.4.13 (the binary's library tags say 3.2.1) |
| `LegacyConnectionSignature` | **false** — `true` leaves the console stuck on 106-0502 |
| game server IDs | `10111700` (USA), `10111a00` (EUR), `10140700` (JPN) |

The USA ID was read out of the executable. EUR and JPN come from the title
database and haven't been checked against a console of that region (a EUR
console did request `10111700`, so one ID may well cover everything). Map every
ID to the same auth port in whatever hands out your NEX tokens.

## Things that bit me

Written down because they cost real time, and they'd apply to other titles of
the same era:

- **An empty ranking list crashes the console.** The lobby asks for the
  player's own stats (`GetRanking`, mode 4, category 0). On success the game
  reads element 0 of the result without checking its size. For a player with no
  data it must get `Ranking::NotFound`, not an empty success. This is done in
  `internal/.../ranking/get_ranking.go`.
- **`PerpetuateObject` has to pin the object to the slot.** The game creates
  its profile objects with `PostMetaBinary`, pins each one with
  `PerpetuateObject(slot, dataID)`, then reads them back by `(owner, slot)`.
  Acknowledging the call without storing the slot leaves the client sitting on
  "connecting to lobby" forever.
- **Matchmaking attribute 4 is the console region** (EUR is 1, USA is 0).
  This server ignores it in session searches so regions can play together.
  Remove the `i == 4` case in
  `matchmake-extension/database/find_matchmake_session_by_search_criteria.go`
  if you want them split again.
- **Empty gatherings** (registered, no participants) used to make the search
  query fail on a NULL `array_length`. They are skipped now.

How these were found, and the addresses in the executable, are in
[RECON.md](RECON.md).

## Running it

```bash
cp .env.example .env
cp settings.example.json settings.json
docker compose up --build
```

That gives you the auth server on UDP 28900 and the secure server on UDP 28901,
plus a PostgreSQL container. Local mode (the default) reads accounts from
`settings.json` and doesn't validate tokens, so only use it on a network you
control.

For a shared setup, set `PN_INJ_LOCAL_MODE=0` and provide
`PN_INJ_NEX_TOKEN_AES_KEY` (64 hex chars) and `PN_INJ_NEX_PASSWORD_SECRET`
(32+ bytes, hex) matching your account server. DataStore uploads need an
S3-compatible bucket (`PN_S3_*`, MinIO works).

Without Docker: `go build -o inj-nex .` and export the `PN_INJ_*` variables from
`compose.yaml`, including `PN_INJ_POSTGRES_URI`.

Every RMC call is printed as `[INJ Secure] PID=… protocol=… method=…`, which is
the quickest way to see what the game wants next. `[MM-SEARCH]`, `[INJ NAT]`
and `[INJ DataStore] GetMeta` lines cover matchmaking, NAT traversal and
profile reads.

## Layout

```
nex/        auth + secure servers, protocol wiring
database/   ranking and DataStore queries (inj_* tables)
globals/    config, accounts, login validation
internal/   vendored fork of nex-protocols-common-go (matchmaking, DataStore/S3)
recon/      the scripts used to pull constants out of the retail .rpx
```

## Credits and license

Almost everything here stands on the Pretendo Network's work. kinnay's
[NintendoClients](https://github.com/kinnay/NintendoClients) documentation and
game database were what made this possible at all.

AGPL-3.0, see [LICENSE](LICENSE) and [NOTICE](NOTICE). No Nintendo or
Warner Bros. / NetherRealm code or assets are included, and the `.rpx` is not
part of this repository.
