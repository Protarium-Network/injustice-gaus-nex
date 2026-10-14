# Protocol coverage

What the server answers, and what the retail client has actually been seen to call.

| Protocol | Coverage |
|---|---|
| Ticket Granting | `Login` / `LoginEx` / `RequestTicket` (stock) |
| Secure Connection | stock, insecure `Register` enabled for Wii U |
| Utility | stock |
| Ranking | `UploadScore`, `UploadCommonData`, `GetRanking` (range / near / user), `GetCommonData`, `GetCachedTopX`; `inj_*` tables. `GetRanking` answers `Ranking::NotFound` for an empty result (the client crashes on an empty success) |
| DataStore | `PreparePostObject` / `CompletePostObject` / `PrepareGetObject`, `GetMeta`, `PostMetaBinary`, `PerpetuateObject`, `GetPersistenceInfo`, `ChangeMetaV1` / `ChangeMeta`, `SearchObject`; ratings are a stub; blobs in S3 |
| NAT Traversal | stock; results are logged as `[INJ NAT]` |
| MatchMaking / MatchMakingExt | stock common handlers |
| MatchmakeExtension | `AutoMatchmakeWithSearchCriteria_Postpone` (pre-3.4 decoder), `AutoMatchmakeWithGatheringID_Postpone` (friends only), `Create/Join/BrowseMatchmakeSession`, `UpdateMatchmakeSession`, `UpdateApplicationBuffer`, `OpenParticipation`, notification data calls |

Not implemented: `BrowseMatchmakeSessionWithHostUrls`, Notification, Monitoring,
Health, AccountManagement (the client does not need them answered).

Seen on a console so far: login and secure register, the DataStore profile
bootstrap, `GetRanking` (modes 0 and 4), `AutoMatchmakeWithSearchCriteria_Postpone`,
`UpdateMatchmakeSession`, `EndParticipation`, NAT traversal reports.
