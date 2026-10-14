# Recon — Injustice: Gods Among Us (Wii U)

Source: `DCFGame-Retail.rpx` (NetherRealm internal name "DCF", USA build).
Everything below was read from the executable; reproduce with `recon/`
(`python -I unrpx.py DCFGame-Retail.rpx dcf.bin`, then `findgid.py`,
`keyrefs.py`, `lits.py`, `dis.py` — see `recon/README.md`).

## Constants

| Field | Value | Evidence |
|---|---|---|
| Access key | `65e9f4d6` | UTF-16BE literal at `.rodata 0x10107100` (also a copy at `0x10138ba8`); returned by a one-line getter at `0x290e72c` (`lis r3,0x1010 ; addi r3,r3,0x7100 ; blr`) and another at `0x2bd6314`. kinnay.github.io `nexwiiu.json`: id 269555456, key `65e9f4d6`. |
| Game server ID | `0x10111700` | Getter at `0x290e720` / `0x2bd6308`: `lis r3,0x1011 ; addi r3,r3,0x1700`. Equals the unique-ID half of the USA title `0005000010111700`; kinnay lists `aid 0005000010111700`. |
| NEX build | `3.4.13` (`ngs/3.4.x.3`) | kinnay: `build:3_4_13_3_0`. The binary's `NEX_3_2_1` / `NEX_MM_3_2_1` / `NEX_DS_3_2_1` / `NEX_RK_3_2_1` strings are per-library tags, as on the other 3.4.x titles. |

### Regions

| Region | Product code | Title ID | Game server ID | Source |
|---|---|---|---|---|
| USA | WUP-P-AJSE | `0005000010111700` | `10111700` | read from the binary; wiiubrew title DB |
| EUR | WUP-P-AJSP | `0005000010111A00` | `10111a00` | wiiubrew title DB (derived: unique-ID half) |
| JPN | WUP-P-AJSJ | `0005000010140700` | `10140700` | wiiubrew title DB (derived) |

The binary's ID equals its own title-ID half, so the same rule is assumed for
the other regions. **EUR/JPN IDs stay unconfirmed until a console of that
region connects** (DKC:TF is a counter-example where one ID covers all
regions, so registering all three is the safe choice). A EUR console (title `0005000010111A00`) was observed requesting a nex_token for `10111700`, so in practice one ID covers EUR and USA. Only the nex_token
front-end depends on them.

## What the game links

`nn::nex` client classes present: `RendezVous` login, `SecureConnection`,
`NATTraversal` (+Relay, ReportInternal), `MatchMakingProtocol`,
`MatchMakingProtocolExt`, `MatchmakeExtension`, `Ranking`, `DataStore`
(+Hpp / LogicServer / FileServer clients, notification URLs),
`Notification` / `NintendoNotificationEvent`, `UtilitySubsystem`.

Game-side strings show what each is for:

- **Ranking** — `OnlineWiiUWriteRankingJob` (StepUpdateScore →
  StepUploadCommonData), `OnlineWiiUSearchRankingJob` (SearchRanking +
  GetStats), `OnlineLeaderboardReadStatsJob`, leaderboards Ranked / Player /
  KOTH / Survivor / Tag, columns Wins / Losses / Streak / Best Streak /
  Kills / Deaths / Games Played / Respect / Skill.
- **Matchmaking** — `AutoMatchmake`, `AutoMatchmakeWithSearchCriteria`,
  `AutoMatchmakeWithGatheringId`, `BrowseMatchmakeSession(+WithHostUrls)`,
  `CreateMatchmakeSession`, `JoinMatchmakeSession(+Ex)`,
  `UpdateMatchmakeSessionAttribute`, `UpdateGatheringHost`,
  `MigrateGatheringOwnership`, host migration, game invites
  (`OnlineGatheringMatchInviteJob`).
- **DataStore** — full client incl. `PrepareGetObject`, `PreparePostObject`,
  ratings with slots, notifications.
- **Friends** — Wii U friend list based; `GetFriendNotificationData`.

## Server consequences

Auth + Secure endpoints; secure registers SecureConnection, Utility, NAT
Traversal, MatchMaking, MatchMakingExt, MatchmakeExtension, Ranking,
DataStore. Which exact RMC methods the retail code calls was **not** mapped
(the executable is stripped); the server enables the full common protocol set
and logs every call (`[INJ Secure] PID=… protocol=… method=…`) so unhandled
methods show up on the first console session. `BrowseMatchmakeSessionWithHostUrls`
is not implemented by the common library.

## Console findings (2026-10-12)

- **LegacyConnectionSignature must be `false`.** The SYN/CONNECT HMACs captured
  from the console reproduce exactly with access key `65e9f4d6`, empty
  connection signature for SYN and the server's SYN-ACK signature for CONNECT
  (standard scheme). `true` leaves the console retransmitting CONNECT forever
  (106-0502).
- **Login bootstrap:** `LoginEx` → `RequestTicket` → `Register`, then
  DataStore `GetMeta(owner, slot 0)`; when not found the client creates six
  profile objects with `PostMetaBinary` (slots 0, 1, 2, 3, 14, 15) pinning each
  with `PerpetuateObject(slot, dataID, deleteLast=true)`, then re-reads them by
  `(owner, slot)` through `GetMeta`. `PerpetuateObject` must bind the object to
  the slot or the client stays on "Connexion au lobby".
- **`Ranking::GetRanking` mode 4 (User), category 0, length 1** is the lobby's
  own-stats query (function at `0x293ec94`, result handler at `0x293ee3c`).
  On success it reads element 0 of the result list with no size check — an empty
  success crashes the console (null dereference). The handler has a separate
  branch for error `0x80670005` = `Ranking::NotFound`, which is what a player
  with no data must receive.
