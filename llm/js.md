# StaticBackend JavaScript Library

A lightweight client library for StaticBackend API.

## Installation & Setup

### Installation

```sh
npm install @staticbackend/js@latest
```

### Setup

Usually in something like `src/sb.ts`

```typescript
import { Backend, type Filter, type Operator } from "@staticbackend/js";

const pubKey = import.meta.env.VITE_BACKEND_PK;
const region = import.meta.env.VITE_BACKEND_REGION;

const bkn = new Backend(pubKey || "public key required", region || "dev");

export const backend = bkn;
export type { Filter, Operator };
```

Region options: `'na1'`, `'dev'` (localhost:8099), or custom URL if self-hosting.

## Response Pattern

All async methods return: `{ok: boolean, content: any}`

The `content` take the shape of what's requested, can be a string, a JSON object.

**Standard usage pattern:**
```typescript
import { backend } from "./sb";

const res = await backend.someMethod(token);
if (!res.ok) {
  // Handle error - res.content contains error message as a string
  console.error(res.content);
  return;
}
// Happy path - res.content contains the result
const data = res.content;
```

When creating any entities an `id` and the `accountId` are automatically populated and returned. Same goes for fetching, `accountId` is always added automatically by the BaaS.

The `token` is the authentication token received after a successful `register` or `login`. All BaaS function requires an authentication token.

A good way to know that the authentication token has expired / isn't valid anymore is by calling the `backend.me(token)` which fails for invalid token.

## Examples / schenario

Note that the checks for success `if (!res.ok)`  is omited to save spaces, in production you always need to check the `ok`  field.

### Authentication

```typescript
// Register - returns token string
const res = await backend.register(email, password);
const token = res.content; // string token

// Login - returns token string
const res = await backend.login(email, password);
const token = res.content; // string token

// Get current user
const res = await backend.me(token);
const user = res.content;
// content is a JSON object
{
  id: string;
  accountId: string;
  email: string;
  role: number;
}
```

### Account Management

```typescript
// List all users in account
const res = await backend.users(token);

// Add user to account
const res = await backend.addUser(token, email, password);
// res.content is the new user's id as a string

// Remove user from account
await backend.removeUser(token, userId);
```

### Database Operations

All database operations require a `token` and `repo` (collection name).

The `filters` is an array of `Filter` defiend as follow:

```ts
export type Operator = "==" | "!=" | "<" | "<=" | ">" | ">=" | "in" | "!in";

export type Filter = [string, Operator, any];
```

#### Create
```typescript
const res = await backend.create(token, 'tasks', { title: 'Task 1' });
// res.content is the new entity created a JSON object
{ id: "new-id-here", accountId: "auth-acct-id", title: "Task 1" }

const res = await backend.createBulk(token, 'tasks', [{ title: 'Task 1' }, { title: 'Task 2' 
  }]);
  // res.content is an array of created entities
```

#### Read
```typescript
// List all - returns {page: number, size: number, total: number, results: Array<T>}
const res = await backend.list(token, 'tasks');
const { page, size, total, results } = res.content;

// Get by ID
const res = await backend.getById(token, 'tasks', 'doc-id');
// res.content is the JSON object of the entity

// Query with filters - returns {page: number, size: number, total: number, results: Array<T>}
const filters = [
  ["field", "==", value]
]
const res = await backend.query(token, 'tasks', filters);
const { page, size, total, results } = res.content;

// Optional ListParam for the .list and .query
// { page?: number, size?: number, desc?: boolean }
// Allow to specify page and page size and the desc order by descending based on creation dates.

// Search
const res await backend.search(token, 'tasks', 'keywords');
const { page, size, total, results } = res.content;

// Count
const res = await backend.count(token, 'tasks', filters);
// res.content is the total number of entities matching the criteria```
```

#### Update
```typescript
const res = await backend.update(token, 'tasks', 'doc-id', { title: 'Updated' });
// res.content is the updated entity

// Note that you may pass only the field you want to update, no need to pass then entire object.

// Bulk update
const bulkData = {
  update: { status: 'completed' },
  clauses: [['field', '==', 'value']]
};
await backend.updateBulk(token, 'tasks', bulkData);
```

#### Delete
```typescript
await backend.delete(token, 'tasks', 'doc-id');
await backend.deleteBulk(token, 'tasks', filters);
// where filters is usual [["field 1", "op1", "value1"]]
```

### File Storage

```typescript
// Upload field from <input> element
const res = await backend.uploadFile(token, document.getElementbyId("file-input-upload"));
const { id, url } = res.content;
```

### Real-time & Messaging

```typescript
// Publish message to channel
await backend.publish(token, 'channel-name', 'message-type', { data: 'value' });

// WebSocket connection
backend.connectWS(token,
  (tok) => { /* onAuth callback */ },
  (payload) => { /* onMessage callback */ }
);
backend.sendWS('message-type', 'data', 'channel');

// SSE connection
backend.connect(token,
  (tok) => { /* onAuth callback */ },
  (payload) => { /* onMessage callback */ }
);
backend.send('message-type', 'data', 'channel');
```

## Other Features

```typescript
// Convert URL to PDF
await backend.convertURLToX(token, {
  toPDF: true,
  url: 'https://example.com',
  fullpage: true
});

// Social login (opens popup)
const user = await backend.socialLogin('google'); // 'twitter' | 'google' | 'facebook'
// Returns ExternalUser: {token, email, name, first, last, avatarUrl}
```

