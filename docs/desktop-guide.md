---
title: Desktop and offline synchronization
description: Wails HR/admin console, local encryption, queueing and conflict semantics.
section: docs
permalink: /desktop-guide/
---

# Desktop and offline synchronization

Advance HRIS includes a native desktop console for HR and admin users. The desktop is built with **Wails, React and Go** and shares the same authenticated Go API used by the web portal.

## What is stored locally?

- Employee payloads are encrypted with **AES-256-GCM** before storage in SQLite.
- The local data encryption key and cloud bearer token are held in the operating-system keychain.
- The frontend receives session metadata rather than the raw bearer token.
- Local employee records stay locked until an online HR/admin session has been validated.

Local storage is a cache and editing workspace, **not** an alternate source of authority for company identity, tenant scope or permissions.

## Installation for developers

1. Run the Go API and PostgreSQL from Docker or your local development stack.
2. Install compatible Go and Wails tooling, plus frontend Node.js dependencies.
3. Start the desktop client:

```sh
cd apps/desktop
wails dev
```

The default API address is `http://localhost:8080`. For a different environment:

```sh
ADVANCE_HRIS_API_URL=https://hr-api.example.com wails dev
```

The exact Wails production packaging and signing procedure must be validated per target operating system. Do not treat `wails dev` as an installer.

## Offline edits and the queue

1. HR/admin signs in online and the app validates the session.
2. Employee edits are saved to encrypted SQLite first.
3. A durable mutation queue records changes, coalescing repeated unsynced updates where possible.
4. On synchronization, the desktop sends queued changes to `POST /api/v1/sync/push`.
5. The server validates the user, tenant, role and base server version.
6. Accepted operations are removed from the queue; conflicts are reported separately.
7. The client pulls the cloud employee snapshot via `GET /api/v1/sync/pull`.

Pulls do **not** silently overwrite an employee record with a pending local edit.

## Resolving conflicts

If a server record changed after the desktop's base version:

- the server returns a conflict and the pending local mutation remains;
- review the remote and local versions with an authorized operator;
- do not delete the local queue or rewrite version numbers to suppress the conflict;
- use the HRIS web view and a controlled reconciliation procedure.

A complete in-app conflict resolution UI and efficient delta synchronization are **planned**, not shipped. The current pull endpoint can return a full employee snapshot.

## Security cautions

- Offline data is still sensitive; protect the OS account and disk.
- A keychain entry is not equivalent to device management or full-disk encryption.
- Logout, key rotation, incident response and lost-device handling require operational validation.
- The agent service is a web-only capability in this increment; native offline desktop AI is not implemented.

See [architecture]({{ '/architecture/' | relative_url }}), [security]({{ '/security-and-privacy/' | relative_url }}) and [troubleshooting]({{ '/troubleshooting/' | relative_url }}).
