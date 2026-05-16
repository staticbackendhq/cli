# StaticBackend JavaScript Client Library

Coding-agent reference for `@staticbackend/js` v1.7.

Use this file when generating JavaScript or TypeScript code against StaticBackend. Keep the API shapes aligned with `../backend-js/src/backend.ts`.

## Install And Setup

```sh
npm install @staticbackend/js@latest
```

```ts
import { Backend, type Filter, type Operator } from "@staticbackend/js";

const pubKey = import.meta.env.VITE_BACKEND_PK;
const region = import.meta.env.VITE_BACKEND_REGION;

export const backend = new Backend(pubKey || "public key required", region || "dev");
export type { Filter, Operator };
```

Region values:

- `"na1"`: managed North America API, `https://na1.staticbackend.dev`
- `"dev"`: local CLI API, `http://localhost:8099`
- Any URL with length >= 10: custom self-hosted base URL. The WebSocket URL is derived by replacing `https` with `wss`.

## Response Pattern

Most async methods return:

```ts
type BackendResponse<T = unknown> = { ok: boolean; content: T };
```

Always check `ok` before using `content`. On failures, `content` is usually an error string from the API, but network/runtime errors may be an Error-like object.

```ts
const res = await backend.me(token);
if (!res.ok) {
  console.error(res.content);
  return;
}

const user = res.content;
```

Exceptions:

- `uploadFile(token, file)` returns `Promise<UploadedFile>` and throws if the upload fails.
- `socialLogin(provider)` returns `Promise<ExternalUser>`.
- Realtime `send`/`sendWS` return `boolean`.

Entity documents returned by DB APIs are account-scoped and normally include server-populated `id` and `accountId` fields.

## Exported Types

```ts
export type Operator = "==" | "!=" | "<" | "<=" | ">" | ">=" | "in" | "!in";
export type Filter = [string, Operator, any];

export interface ListParams {
  page?: number;
  size?: number;
  desc?: boolean;
}

export interface Payload {
  sid: string;
  type: string;
  data: string;
  channel: string;
  token: string;
}

export interface ConvertData {
  toPDF: boolean;
  url: string;
  fullpage: boolean;
}

export interface ExternalUser {
  token: string;
  email: string;
  name: string;
  first: string;
  last: string;
  avatarUrl: string;
}

export interface BulkUpdate {
  update: any;
  clauses: Array<Array<any>>;
}

export interface UploadedFile {
  id: string;
  url: string;
}

export interface FileUsage {
  bytes: number;
  gb: number;
}

export interface StoredFile {
  id: string;
  accountId: string;
  key: string;
  url: string;
  size: number;
  uploaded: string;
}

export interface StorageListParams extends ListParams {
  sort?: "size";
}

export interface FileListResult {
  page: number;
  size: number;
  total: number;
  results: StoredFile[];
}

export interface AccountUser {
  id: string;
  userId: string;
  accountId: string;
  email: string;
  role: number;
  token: string;
}
```

## Authentication And Account APIs

```ts
const registerRes = await backend.register(email, password); // content: string token
const loginRes = await backend.login(email, password); // content: string token
const accountLoginRes = await backend.login(email, password, accountId); // optional accountId

const meRes = await backend.me(token);
// content: { id: string; accountId: string; email: string; role: number }

await backend.changeEmail(token, "new@example.com");
await backend.setRole(token, email, role, accountId); // accountId is optional

const usersRes = await backend.users(token); // content: account users
const addUserRes = await backend.addUser(token, email, password); // content: created user
await backend.removeUser(token, userId);

const associationsRes = await backend.listAssociations(token); // content: AccountUser[]
const promoteRes = await backend.promoteUser(token); // content: new session token
```

Use `backend.me(token)` as a lightweight way to verify that a session token is still valid.

## Database APIs

All normal DB methods require a user `token` and `repo`, where `repo` is the collection name.

```ts
const filters: Filter[] = [["status", "==", "active"]];
const params: ListParams = { page: 1, size: 25, desc: true };
```

Create:

```ts
const created = await backend.create(token, "tasks", { title: "Task 1" });
const createdMany = await backend.createBulk(token, "tasks", [
  { title: "Task 1" },
  { title: "Task 2" },
]);
```

Read:

```ts
const listRes = await backend.list(token, "tasks", params);
// content: { page: number; size: number; total: number; results: T[] }

const oneRes = await backend.getById(token, "tasks", id);
const manyRes = await backend.getByIds(token, "tasks", [id1, id2]);

const queryRes = await backend.query(token, "tasks", filters, params);
const countRes = await backend.count(token, "tasks", filters); // content: { count: number }
const searchRes = await backend.search(token, "tasks", "keywords");

const idRes = await backend.newId(token); // content: generated id string
```

Update:

```ts
const updated = await backend.update(token, "tasks", id, { title: "Updated" });

const bulk: BulkUpdate = {
  update: { status: "done" },
  clauses: [["status", "==", "pending"]],
};
const updateCount = await backend.updateBulk(token, "tasks", bulk);

await backend.increase(token, "tasks", id, "priority", 1);
await backend.increase(token, "tasks", id, "priority", -1);
```

Delete:

```ts
await backend.delete(token, "tasks", id);
await backend.deleteBulk(token, "tasks", filters);
```

## File Storage

```ts
// Use a File from an <input>, drag/drop, or File API.
const file = input.files?.[0];
if (file) {
  const uploaded = await backend.uploadFile(token, file);
  // uploaded: { id: string; url: string }
}

// Legacy form upload API. Returns { ok, content }.
const formUpload = await backend.storeFile(token, formElement);

const usage = await backend.storageUsage(token);
// content: FileUsage

const files = await backend.listFiles(token, { page: 1, sort: "size" });
// content: FileListResult
```

`resizeImage(token, maxWidth, form)` uploads an image form, resizes it server-side, and returns `{ ok, content }` with a stored file result.

```ts
const resized = await backend.resizeImage(token, 1200, formElement);
```

## Server-Side Functions And Messaging

```ts
await backend.execFunction(token, "functionName", { any: "payload" });
```

The JS client can execute a named server-side function directly. It does not expose the Go client's `Publish` helper; for realtime channel messages use `send` or `sendWS`.

## Realtime

WebSocket:

```ts
backend.connectWS(
  token,
  (realtimeToken) => {},
  (payload) => {},
);

const sent = backend.sendWS("message-type", "string data", "channel-name");
```

Server-Sent Events:

```ts
backend.connect(
  token,
  (realtimeToken) => {},
  (payload) => {},
);

const sent = backend.send("message-type", "string data", "channel-name");
```

Known payload types are exposed under `backend.types`, including `init`, `token`, `auth`, `join`, `chan_in`, `chan_out`, `presence`, `db_created`, `db_updated`, and `db_deleted`.

## Extras And Social Login

```ts
const pdf = await backend.convertURLToX(token, {
  toPDF: true,
  url: "https://example.com",
  fullpage: true,
});

const screenshot = await backend.convertURLToX(token, {
  toPDF: false,
  url: "https://example.com",
  fullpage: true,
});

const externalUser = await backend.socialLogin("google");
// providers: "twitter" | "google" | "facebook"
```

`socialLogin` opens a popup and polls until the provider flow completes. On timeout it returns an empty `ExternalUser`.

## Coding-Agent Notes

- Prefer the exported types instead of duplicating local shapes.
- `uploadFile` is not the same response pattern as the other storage methods; wrap it in `try/catch`.
- `storeFile` and `resizeImage` expect an `HTMLFormElement`; `uploadFile` expects a `File`.
- `count` returns the API payload from `/db/count/{repo}`. In Go this is unwrapped to `int64`; in JS read `res.content.count` if the API returns `{ count }`.
- `deleteBulk` base64-encodes filters internally; pass the normal `Filter[]`.
- The JS client does not expose Go-only sudo DB helpers, system-account helpers, function management CRUD, cache helpers, queue helpers, `Publish`, form listing, SMS, email, or file deletion.
