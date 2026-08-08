# Research: TrueNAS `user.update` partial-update semantics

## Summary

Primary sources support treating `user.update` as PATCH-like across the relevant releases: 24.10 uses the same v25.04 API model, and 25.04/25.10 define `UserUpdate` with `ForUpdateMetaclass`, which makes any subset of fields acceptable. Omitted fields preserve existing values because the implementation loads the current user, overlays only submitted keys with `user.update(data)`, and then writes the merged record.

## Findings

1. **Update fields are optional in the versioned middleware API models** — TrueNAS 24.10 tag `TS-24.10.0` imports the v25.04 API as current (`from .v25_04_0 import *`), and v25.04.0 defines `class UserUpdate(UserCreate, metaclass=ForUpdateMetaclass): ...`. The metaclass states: “Using this metaclass on a model will change all of its fields default values to `undefined`. Such a model might be instantiated with any subset of its fields, which can be useful to validate request bodies for requests with PATCH semantics.” [TS-24.10.0 current.py](https://raw.githubusercontent.com/truenas/middleware/TS-24.10.0/src/middlewared/middlewared/api/current.py), [TS-25.04.0 user.py](https://raw.githubusercontent.com/truenas/middleware/TS-25.04.0/src/middlewared/middlewared/api/v25_04_0/user.py), [TS-25.04.0 model.py](https://raw.githubusercontent.com/truenas/middleware/TS-25.04.0/src/middlewared/middlewared/api/base/model.py)

2. **Official API docs agree that `user_update` is a closed object, with nullable `email`/`sshpubkey`, and no required per-field list shown** — v25.04.0 documents parameter 2 as “`user_update` / Type: object / No Additional Properties”; field entries include booleans such as `Smb`, `Password Disabled`, `Ssh Password Enabled`, `Locked`, and nullable fields: `Sshpubkey ... Type: string ... Type: null` and `Email ... Type: string Format: email ... Type: null`. v25.10.5 repeats the same call shape and nullable field types, adding descriptions but not changing update optionality. [v25.04.0 docs](https://api.truenas.com/v25.04.0/api_methods_user.update.html), [v25.10 docs](https://api.truenas.com/v25.10/api_methods_user.update.html)

3. **Omitted fields preserve existing values** — the implementation loads the existing user (`user = self.middleware.call_sync('user.get_instance', pk)`), chooses existing group/groups when omitted (`else: group = user['group']; user['group'] = group['id']`; `else: group_ids.extend(user['groups'])`), computes `home = data.get('home') or user['home']`, then explicitly overlays submitted data only: “`# After this point user dict has values from data` / `user.update(data)`”. The merged record is compressed and written with `datastore.update`. This pattern is present in TS-24.10.0, TS-25.04.0, and TS-25.10.0 source. [TS-24.10.0 account.py](https://raw.githubusercontent.com/truenas/middleware/TS-24.10.0/src/middlewared/middlewared/plugins/account.py), [TS-25.04.0 account.py](https://raw.githubusercontent.com/truenas/middleware/TS-25.04.0/src/middlewared/middlewared/plugins/account.py), [TS-25.10.0 account.py](https://raw.githubusercontent.com/truenas/middleware/TS-25.10.0/src/middlewared/middlewared/plugins/account.py)

4. **Email clearing is `null`, not omission** — API schemas allow `email` as string email or `null`; v25.10 source defines update input via `UserCreate.email: EmailStr | None = None`, inherited by `UserUpdate`. Because update merges submitted `data`, omitting `email` leaves `user['email']` unchanged, while submitting `"email": null` overwrites the merged record and is persisted. Source also normalizes stored empty email on read: “`# Normalize email, empty is really null` / `if user['email'] == '': user['email'] = None`.” [v25.04.0 docs](https://api.truenas.com/v25.04.0/api_methods_user.update.html), [TS-25.10.0 user.py](https://raw.githubusercontent.com/truenas/middleware/TS-25.10.0/src/middlewared/middlewared/api/v25_10_0/user.py), [TS-25.10.0 account.py](https://raw.githubusercontent.com/truenas/middleware/TS-25.10.0/src/middlewared/middlewared/plugins/account.py)

5. **SSH public-key clearing is `null` or empty/whitespace string; omission preserves by reusing current key** — `update_sshpubkey` begins `if 'sshpubkey' not in user: return`, then `pubkey = user.get('sshpubkey') or ''`, `pubkey = pubkey.strip()`, and if `pubkey == ''` it unlinks `authorized_keys` and returns. If non-empty, it compares the current file contents and returns unchanged when equal. Because `do_update` starts from `get_instance` (which includes the current `sshpubkey`) and then overlays submitted `data`, omitting `sshpubkey` keeps the existing key value; submitting `null`, `""`, or whitespace clears/removes `authorized_keys`. [TS-25.10.0 account.py](https://raw.githubusercontent.com/truenas/middleware/TS-25.10.0/src/middlewared/middlewared/plugins/account.py), [v25.10 docs](https://api.truenas.com/v25.10/api_methods_user.update.html)

6. **Boolean `false` must be distinguishable from omission in `UpdateUserOpts`** — source branches use both presence checks and value checks. For example, disabling SMB is detected only by explicit false: `if user['smb'] is True and data.get('smb') is False:`; enabling uses `data.get('smb') is True`; immutable validation tests key presence: `if i in data and data[i] != user[i]`; and the final `user.update(data)` overwrites false values when present. Therefore a Go client must be able to encode `false` when explicitly requested and omit the field otherwise (for example, `*bool` with `omitempty`). [TS-25.04.0 account.py](https://raw.githubusercontent.com/truenas/middleware/TS-25.04.0/src/middlewared/middlewared/plugins/account.py), [TS-25.10.0 account.py](https://raw.githubusercontent.com/truenas/middleware/TS-25.10.0/src/middlewared/middlewared/plugins/account.py)

7. **No material version difference affects `UpdateUserOpts` partial semantics across 24.10, 25.04, and 25.10** — 24.10 uses v25.04 API models. The 25.04 and 25.10 models both use `UserUpdate(... ForUpdateMetaclass)` and the same merge implementation. The repository's embedded `api/25.04/methods.json` has no `required` list for the `user.update` object. It marks each update property with `_required_: false`. The differences between these releases affect descriptions and return shapes, not update payload semantics. [embedded schema](../../api/25.04/methods.json), [v25.04 docs](https://api.truenas.com/v25.04.0/api_methods_user.update.html), [v25.10 docs](https://api.truenas.com/v25.10/api_methods_user.update.html)

## Implications for `UpdateUserOpts`

- Model every update field as optional/omittable.
- Use pointer booleans (or equivalent tri-state encoding) for `smb`, `password_disabled`, `ssh_password_enabled`, `locked`, `home_create`, and `random_password`, because `false` is a meaningful submitted value distinct from omission.
- Use pointer/string-null-capable encoding for `email` and `sshpubkey`:
  - omit = preserve existing value;
  - `email: null` = clear email;
  - `sshpubkey: null`, `""`, or whitespace = remove authorized keys.

## Sources

- Kept: TrueNAS API v25.04.0 `user.update` docs (https://api.truenas.com/v25.04.0/api_methods_user.update.html) — official method schema for 25.04.0.
- Kept: TrueNAS API v25.10 `user.update` docs (https://api.truenas.com/v25.10/api_methods_user.update.html) — official method schema for 25.10.
- Kept: `truenas/middleware` TS-24.10.0 `current.py` and `account.py` — official source showing 24.10 API model selection and implementation.
- Kept: `truenas/middleware` TS-25.04.0 `user.py`, `model.py`, `account.py` — official source for optional/PATCH metaclass and merge implementation.
- Kept: `truenas/middleware` TS-25.10.0 `user.py`, `model.py`, `account.py` — official source for 25.10 schema and implementation.
- Kept: local `api/25.04/methods.json` — confirms that every `user.update` property is optional.
- Dropped: forums/blogs/SEO summaries — not primary sources.

## Gaps

- `https://api.truenas.com/v24.10/api_methods_user.update.html` currently returns 404. The official `TS-24.10.0` source establishes the 24.10 behavior.
