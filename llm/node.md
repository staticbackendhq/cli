# StaticBackend Node Client Library

Coding-agent reference for `@staticbackend/backend` v1.7.

Use this file when generating Node.js or TypeScript code against StaticBackend. Keep the API shapes aligned with `../backend-node/src/backend.ts`.

The Node client is intended for trusted server-side code and mirrors the Go client surface, including root-token helpers. For browser code, use the JavaScript client guide instead.

## Install And Setup

```sh
npm install @staticbackend/backend@latest
```

```ts
import {
  Backend,
  type Filter,
  type ListParam,
  type BulkUpdate,
} from "@staticbackend/backend";

const backend = new Backend(process.env.SB_PUBLIC_KEY || "public key required", process.env.SB_REGION || "dev");
```

Region values:

- `"na1"`: managed North America API, `https://na1.staticbackend.dev`
- `"dev"`: local CLI API, `http://localhost:8099`
- Any value with length > 3: custom self-hosted base URL

## Response Pattern

All public async methods return:

```ts
type BackendResponse<T = unknown> = { ok: boolean; content: T };
```

Always check `ok` before using `content`.

```ts
const res = await backend.login(email, password);
if (!res.ok) {
  console.error(res.content);
  return;
}

const token = res.content as string;
```

On failures, `content` is usually an API error string. Network/runtime failures may return an Error-like object.

## Exported Types

```ts
export type Operator = "==" | "!=" | "<" | "<=" | ">" | ">=" | "in" | "!in";
export type Filter = [string, Operator, any];

export interface ListParam {
  page?: number;
  size?: number;
  descending?: boolean;
  desc?: boolean;
}

export interface BulkUpdate {
  update: any;
  clauses: Array<Array<any>>;
}

export interface Attachment {
  url?: string;
  body?: Buffer | string;
  contentType?: string;
  filename?: string;
}

export interface EmailData {
  fromName: string;
  from: string;
  to: string;
  subject: string;
  body: string;
  replyTo: string;
  attachments?: Attachment[];
}

export interface ConvertData {
  toPDF: boolean;
  url: string;
  fullpage: boolean;
}

export interface SMSData {
  accountSID: string;
  authToken: string;
  toNumber: string;
  fromNumber: string;
  body: string;
}

export interface MagicLinkData {
  fromEmail: string;
  fromName: string;
  email: string;
  subject: string;
  body: string;
  link: string;
}

export interface AccountUser {
  id: string;
  userId: string;
  accountId: string;
  email: string;
  role: number;
  token: string;
}

export interface UserAccountEntry {
  accountId: string;
  role: number;
  home: boolean;
  token?: string;
}

export interface User {
  id: string;
  accountId: string;
  token: string;
  email: string;
  role: number;
  created: string;
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

export interface StorageListParam {
  page?: number;
  sort?: "size";
}

export interface FileListResult {
  page: number;
  size: number;
  total: number;
  results: StoredFile[];
}

export interface RunHistory {
  id: string;
  functionId?: string;
  version: number;
  started: string;
  completed: string;
  success: boolean;
  output: string[];
}

export interface FunctionData {
  id?: string;
  accountId?: string;
  name: string;
  trigger: string;
  code: string;
  version?: number;
  lastUpdated?: string;
  lastRun?: string;
  history?: RunHistory[];
}
```

## Authentication And Account APIs

```ts
const reg = await backend.register(email, password, accountId); // accountId optional
const login = await backend.login(email, password, accountId); // accountId optional
const exists = await backend.emailExists(email);

const magic = await backend.requestMagicLink({
  fromEmail: "noreply@example.com",
  fromName: "Example",
  email,
  subject: "Your sign-in link",
  body: "<p>Click {{link}}</p>",
  link: "https://app.example.com/login/magic",
});
const magicLogin = await backend.loginWithMagicLink(email, code);

const me = await backend.me(token);
const changedEmail = await backend.changeEmail(token, "new@example.com");
const changedPassword = await backend.changePassword(token, email, oldPassword, newPassword);
const resetCode = await backend.getPasswordResetCode(rootToken, email);
const reset = await backend.resetPassword(email, code, newPassword);

const users = await backend.users(token);
const added = await backend.addUser(token, email, password);
const removed = await backend.removeUser(token, userId);
const role = await backend.setRole(rootOrAdminToken, email, roleNumber, accountId);

const associations = await backend.listAssociations(token); // content: AccountUser[]
const promoted = await backend.promoteUser(token); // content: new session token
```

Root-token account APIs:

```ts
const accountToken = await backend.sudoGetToken(rootToken, accountId);
const userAccounts = await backend.sudoGetUserAccounts(rootToken, email);
const userToken = await backend.sudoGetAuthTokenByUserID(rootToken, accountId, userId);
const user = await backend.sudoGetUserByID(rootToken, accountId, userId);
```

## Database APIs

All normal DB methods take `token` and `repo`, where `repo` is the collection name.

```ts
const filters: Filter[] = [["status", "==", "active"]];
const params: ListParam = { page: 1, size: 25, descending: true };
```

Create and read:

```ts
const created = await backend.create(token, "tasks", { title: "Task 1" });
const createdMany = await backend.createBulk(token, "tasks", [
  { title: "Task 1" },
  { title: "Task 2" },
]);

const list = await backend.list(token, "tasks", params);
const one = await backend.getById(token, "tasks", id);
const many = await backend.getByIds(token, "tasks", [id1, id2]);
const queried = await backend.query(token, "tasks", filters, params);
const count = await backend.count(token, "tasks", filters); // content: { count: number }
const searched = await backend.search(token, "tasks", "keywords");
```

Update and delete:

```ts
const updated = await backend.update(token, "tasks", id, { status: "done" });

const bulk: BulkUpdate = {
  update: { status: "done" },
  clauses: [["status", "==", "pending"]],
};
const bulkUpdated = await backend.updateBulk(token, "tasks", bulk);

await backend.increase(token, "tasks", id, "priority", 1);
await backend.increase(token, "tasks", id, "priority", -1);

await backend.delete(token, "tasks", id);
await backend.deleteBulk(token, "tasks", filters);
```

## Root-Token Database APIs

Sudo variants bypass normal account scoping and require a root token.

```ts
const sudoCreated = await backend.sudoCreate(rootToken, repo, doc);
const sudoCreatedMany = await backend.sudoCreateBulk(rootToken, repo, docs);
const sudoList = await backend.sudoList(rootToken, repo, params);
const sudoOne = await backend.sudoGetById(rootToken, repo, id);
const sudoMany = await backend.sudoGetByIds(rootToken, repo, ids);
const sudoUpdated = await backend.sudoUpdate(rootToken, repo, id, patch);
const sudoBulkUpdated = await backend.sudoUpdateBulk(rootToken, repo, bulk);
const sudoQuery = await backend.sudoQuery(rootToken, repo, filters, params);
const sudoFind = await backend.sudoFind(rootToken, repo, filters, params); // alias of sudoQuery
const sudoDeleted = await backend.sudoDelete(rootToken, repo, id);
const sudoBulkDeleted = await backend.sudoDeleteBulk(rootToken, repo, filters);

const repositories = await backend.sudoListRepositories(rootToken);
const index = await backend.sudoAddIndex(rootToken, repo, field, "text");
```

`sudoAddIndex` type is optional and may be `"text"`, `"number"`, `"boolean"`, or `"date"`.

## File Storage

```ts
const buf = await fs.promises.readFile("photo.jpg");

const uploaded = await backend.storeFile(token, buf, "photo.jpg");
// content: UploadedFile

const usage = await backend.storageUsage(token);
// content: FileUsage

const files = await backend.listFiles(token, { page: 1, sort: "size" });
// content: FileListResult

const deleted = await backend.deleteFile(rootToken, uploaded.content.id);
```

`storeFile` accepts `Buffer | ArrayBuffer`. `deleteFile` is a root-token API.

## Email

```ts
const sent = await backend.sendMail(rootToken, {
  fromName: "Example",
  from: "noreply@example.com",
  to: "user@example.com",
  subject: "Hello",
  body: "<p>Hello</p>",
  replyTo: "support@example.com",
});

const withAttachments = await backend.sendMailWithAttachments(rootToken, {
  fromName: "Example",
  from: "noreply@example.com",
  to: "user@example.com",
  subject: "Report",
  body: "<p>Attached</p>",
  replyTo: "support@example.com",
  attachments: [
    { url: "https://example.com/report.pdf" },
    { body: pdfBuffer, contentType: "application/pdf", filename: "report.pdf" },
  ],
});
```

For each attachment, use either `url` or the inline `body`, `contentType`, and `filename` fields.

## Cache, Queue, And Publish

Root-token cache and queue helpers:

```ts
await backend.cacheSet(rootToken, "key", { value: 1 });
const cached = await backend.cacheGet(rootToken, "key");

await backend.queueWork(rootToken, "queue-key", "string value");
const next = await backend.dequeueWork(rootToken, "queue-key");
```

Publish a channel message, usually to trigger a server-side function:

```ts
await backend.publish(token, "channel-name", "message-type", { id });
```

`publish` JSON-stringifies non-string payloads before sending them.

## Server-Side Functions

```ts
const fn: FunctionData = {
  name: "processTask",
  trigger: "tasks",
  code: "module.exports = async function(payload) { return payload }",
};

await backend.addFunction(rootToken, fn);
const functions = await backend.listFunctions(rootToken);
const info = await backend.functionInfo(rootToken, "processTask");
await backend.updateFunction(rootToken, fn);
await backend.deleteFunction(rootToken, "processTask");

const result = await backend.execFunction(token, "processTask", { id });
const sudoResult = await backend.sudoExecFunction(rootToken, "processTask", { id });
```

## Forms

```ts
const all = await backend.listForm(rootToken);
const contact = await backend.listForm(rootToken, "contact");
```

## Extras

Image resize:

```ts
const image = await fs.promises.readFile("photo.png");
const resized = await backend.resizeImage(token, 1200, image.buffer);
```

URL to PDF or PNG:

```ts
const pdf = await backend.convertURLToX(token, {
  toPDF: true,
  url: "https://example.com",
  fullpage: true,
});
```

Root-token SMS via Twilio:

```ts
await backend.sudoSendSMS(rootToken, {
  accountSID: "...",
  authToken: "...",
  toNumber: "+15551234567",
  fromNumber: "+15557654321",
  body: "Hello",
});
```

## Coding-Agent Notes

- Use this guide for server-side Node code. Do not use root-token APIs in browser code.
- The package name is `@staticbackend/backend`; the browser package is `@staticbackend/js`.
- Every method returns `{ ok, content }`, including uploads and root-token helpers.
- Prefer `ListParam.descending` in new Node code; `desc` is accepted for compatibility.
- Normal DB APIs are account-scoped. Sudo DB APIs require the root token and bypass account scoping.
- `deleteBulk` and `sudoDeleteBulk` base64-encode filters internally; pass the normal `Filter[]`.
- `storeFile` and `resizeImage` use Node buffers, not DOM `File` or `HTMLFormElement`.
