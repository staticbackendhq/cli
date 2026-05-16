# StaticBackend Go Client Library

Coding-agent reference for `github.com/staticbackendhq/backend-go` v1.7.

Use this file when generating Go code against StaticBackend. Keep the API shapes aligned with `../backend-go/*.go`.

## Setup

```go
import backend "github.com/staticbackendhq/backend-go"

backend.PublicKey = "your-public-key"
backend.Region = backend.RegionLocalDev // or backend.RegionNorthAmerica1
backend.Verbose = true                   // optional HTTP logging
```

Region values:

```go
backend.RegionNorthAmerica1 // "na1", managed hosting
backend.RegionLocalDev      // "dev", http://localhost:8099
backend.Region = "https://your-self-hosted-api.example.com"
```

For local development with the StaticBackend CLI, common dev credentials are:

```text
public key: dev_memory_pk
root token: safe-to-use-in-dev-root-token
```

## Error And Result Pattern

Go functions return `error`; success means `err == nil`. Functions that return data either return it directly or populate a pointer argument.

```go
token, err := backend.Login("user@example.com", "password")
if err != nil {
    return err
}
```

Entity documents returned by DB APIs are account-scoped and normally include server-populated `id` and `accountId` fields.

## Low-Level HTTP Helpers

These are exported and can be used for endpoints not wrapped by the client:

```go
err := backend.Get(token, "/path", &out)
err := backend.Post(token, "/path", body, &out)
err := backend.Put(token, "/path", body, &out)
err := backend.Del(token, "/path")
```

Prefer higher-level helpers when they exist.

## Authentication And Account APIs

```go
token, err := backend.Register(email, password)
token, err = backend.Login(email, password)
token, err = backend.LoginForAccount(email, password, accountID)

exists, err := backend.EmailExists(email)

me, err := backend.Me(token)
// backend.CurrentUser{AccountID, UserID, Email, Role}

err = backend.ChangeEmail(token, "new@example.com")
err = backend.SetPassword(token, email, oldPassword, newPassword)
err = backend.SetRole(rootOrAdminToken, accountID, email, role)

users, err := backend.Users(token)
user, err := backend.AddUser(token, email, password)
err = backend.RemoveUser(token, userID)

code, err := backend.GetPasswordResetCode(token, email)
err = backend.ResetPassword(email, code, newPassword)
```

Account association APIs:

```go
associations, err := backend.ListAssociations(token) // []backend.AccountUser
newToken, err := backend.PromoteUser(token)
```

Root-token account APIs:

```go
tok, err := backend.SudoGetToken(rootToken, accountID)
entries, err := backend.SudoGetUserAccounts(rootToken, email)
authTok, err := backend.SudoGetAuthTokenByUserID(rootToken, accountID, userID)
user, err := backend.SudoGetUserByID(rootToken, accountID, userID)
```

Exported account types:

```go
type AccountParams struct {
    Email     string `json:"email"`
    Password  string `json:"password"`
    AccountID string `json:"accountId,omitempty"`
}

type CurrentUser struct {
    AccountID string `json:"accountId"`
    UserID    string `json:"id"`
    Email     string `json:"email"`
    Role      int    `json:"role"`
}

type User struct {
    ID        string
    AccountID string
    Token     string
    Email     string
    Role      int
    Created   time.Time
}

type AccountUser struct {
    ID        string `json:"id"`
    UserID    string `json:"userId"`
    AccountID string `json:"accountId"`
    Email     string `json:"email"`
    Role      int    `json:"role"`
    Token     string `json:"token"`
}

type UserAccountEntry struct {
    AccountID string `json:"accountId"`
    Role      int    `json:"role"`
    Home      bool   `json:"home"`
    Token     string `json:"token,omitempty"`
}
```

## Database APIs

All normal DB functions take `token` and `repo`, where `repo` is the collection name.

Filters:

```go
type QueryItem struct {
    Field string
    Op    QueryOperator
    Value interface{}
}

filters := []backend.QueryItem{
    {Field: "status", Op: backend.QueryEqual, Value: "active"},
}
```

Operators: `backend.QueryEqual`, `backend.QueryNotEqual`, `backend.QueryLowerThan`, `backend.QueryLowerThanEqual`, `backend.QueryGreaterThan`, `backend.QueryGreaterThanEqual`, `backend.QueryIn`, and `backend.QueryNotIn`.

Sentinel errors: `backend.ErrNoDocument` and `backend.ErrMultipleDocument`.

Paging:

```go
params := &backend.ListParams{Page: 1, Size: 25, Descending: true}

type ListResult struct {
    Page     int         `json:"page"`
    PageSize int         `json:"size"`
    Total    int         `json:"total"`
    Results  interface{} `json:"results"`
}
```

Create:

```go
var task Task
err := backend.Create(token, "tasks", Task{Title: "Task 1"}, &task)

ok, err := backend.CreateBulk(token, "tasks", []Task{
    {Title: "Task 1"},
    {Title: "Task 2"},
})
```

Read:

```go
var tasks []Task
meta, err := backend.List(token, "tasks", &tasks, params)

var task Task
err = backend.GetByID(token, "tasks", id, &task)

var selected []Task
err = backend.GetByIDs(token, "tasks", []string{id1, id2}, &selected)

meta, err = backend.Find(token, "tasks", filters, &tasks, params)

err = backend.FindOne(token, "tasks", filters, &task)
// returns backend.ErrNoDocument or backend.ErrMultipleDocument when the match count is not exactly one

err = backend.Search(token, "tasks", "keywords", &tasks)

n, err := backend.Count(token, "tasks", filters)
```

Update:

```go
var updated Task
err := backend.Update(token, "tasks", id, map[string]any{"status": "done"}, &updated)

n, err := backend.UpdateBulk(token, "tasks", filters, map[string]any{"status": "done"})

err = backend.Increase(token, "tasks", id, "priority", 1)
err = backend.Increase(token, "tasks", id, "priority", -1)
```

Delete:

```go
err := backend.Delete(token, "tasks", id)
err = backend.DeleteBulk(token, "tasks", filters)
```

## Root-Token Database APIs

Sudo variants bypass normal account scoping and require a root token.

```go
err := backend.SudoCreate(rootToken, repo, body, &out)
meta, err := backend.SudoList(rootToken, repo, &items, params)
err = backend.SudoGetByID(rootToken, repo, id, &out)
err = backend.SudoGetByIDs(rootToken, repo, ids, &items)
err = backend.SudoUpdate(rootToken, repo, id, body, &out)
meta, err = backend.SudoFind(rootToken, repo, filters, &items, params)
err = backend.SudoFindOne(rootToken, repo, filters, &out)
err = backend.SudoDelete(rootToken, repo, id)

names, err := backend.SudoListRepositories(rootToken)
err = backend.SudoAddIndex(rootToken, repo, field)
```

If documents must be created under a specific account, prefer `SudoGetToken(rootToken, accountID)` and then call normal account-scoped DB functions with that returned token.

## File Storage

```go
file, err := os.Open("photo.jpg")
if err != nil {
    return err
}
defer file.Close()

uploaded, err := backend.StoreFile(token, "photo.jpg", file)
// backend.StoreFileResult{ID, URL}

buf, err := backend.DownloadFile(token, uploaded.URL)

usage, err := backend.StorageUsage(token)
// backend.FileUsage{Bytes, GB}

files, err := backend.ListFiles(token, &backend.ListFilesParams{
    Page:   1,
    SortBy: "size",
})
// backend.FileListResult{Page, Size, Total, Results}

ok, err := backend.DeleteFile(token, uploaded.ID)
```

Exported storage types:

```go
type StoreFileResult struct {
    ID  string `json:"id"`
    URL string `json:"url"`
}

type File struct {
    ID        string    `json:"id"`
    AccountID string    `json:"accountId"`
    Key       string    `json:"key"`
    URL       string    `json:"url"`
    Size      int64     `json:"size"`
    Uploaded  time.Time `json:"uploaded"`
}

type FileUsage struct {
    Bytes int64   `json:"bytes"`
    GB    float64 `json:"gb"`
}

type FileListResult struct {
    Page    int64
    Size    int64
    Total   int64
    Results []File
}

type ListFilesParams struct {
    Page   int
    SortBy string // currently supports "size"
}
```

## Email

```go
ok, err := backend.SendMail(token, from, fromName, to, subject, htmlBody, replyTo)

email := backend.EmailData{
    From:     "you@example.com",
    FromName: "Your Name",
    To:       "recipient@example.com",
    Subject:  "Hello",
    Body:     "<p>HTML body</p>",
    ReplyTo:  "reply@example.com",
    Attachments: []backend.Attachment{
        {URL: "https://example.com/file.pdf"},
        {Body: pdfBytes, ContentType: "application/pdf", Filename: "file.pdf"},
    },
}
ok, err = backend.SendMailWithAttachments(token, email)
```

For each attachment, use either `URL` or the inline `Body`, `ContentType`, and `Filename` fields.

Exported email types:

```go
type Attachment struct {
    URL         string `json:"url"`
    Body        []byte `json:"body"`
    ContentType string `json:"contentType"`
    Filename    string `json:"filename"`
}

type EmailData struct {
    FromName    string       `json:"fromName"`
    From        string       `json:"from"`
    To          string       `json:"to"`
    Subject     string       `json:"subject"`
    Body        string       `json:"body"`
    ReplyTo     string       `json:"replyTo"`
    Attachments []Attachment `json:"attachments"`
}
```

## Cache, Queue, And Publish

Root-token cache and queue helpers:

```go
err := backend.CacheSet(rootToken, "key", value)
err = backend.CacheGet(rootToken, "key", &value)

err = backend.QueueWork(rootToken, "queue-key", "string value")

go backend.WorkerQueue(rootToken, "queue-key", func(val string) {
    // process val
})
```

Publish a channel message, usually to trigger a server-side function:

```go
err := backend.Publish(token, "channel-name", "message-type", map[string]any{
    "id": id,
})
```

`WorkerTask` is the exported callback type:

```go
type WorkerTask func(val string)
```

## Server-Side Functions

```go
fn := backend.Function{
    FunctionName: "processTask",
    TriggerTopic: "tasks",
    Code: "module.exports = async function(payload) { return payload }",
}

err := backend.AddFunction(token, fn)
functions, err := backend.ListFunctions(token)
info, err := backend.FunctionInfo(token, "processTask")
err = backend.UpdateFunction(token, fn)
err = backend.DeleteFunction(token, "processTask")
```

Exported function types:

```go
type Function struct {
    ID           string
    FunctionName string
    TriggerTopic string
    Code         string
    Version      int
    LastUpdated  time.Time
    LastRun      time.Time
    History      []RunHistory
}

type RunHistory struct {
    ID        string
    Version   int
    Started   time.Time
    Completed time.Time
    Success   bool
    Output    []string
}
```

## Forms

```go
all, err := backend.ListForm(token, "")
contact, err := backend.ListForm(token, "contact")
```

`ListForm` returns `[]map[string]interface{}`.

## Extras

Image resize:

```go
file, err := os.Open("photo.png")
resized, err := backend.ResizeImage(token, "photo.png", file, 1200)
// returns backend.StoreFileResult
```

URL to PDF or PNG:

```go
out, err := backend.ConvertURLToX(token, backend.ConvertParam{
    ToPDF:    true,
    URL:      "https://example.com",
    FullPage: true,
})
// returns backend.StoreFileResult
```

Root-token SMS via Twilio:

```go
err := backend.SudoSendSMS(rootToken, backend.SMSData{
    AccountSID: "...",
    AuthToken:  "...",
    ToNumber:   "+15551234567",
    FromNumber: "+15557654321",
    Body:       "Hello",
})
```

Exported extra types:

```go
type ConvertParam struct {
    ToPDF    bool   `json:"toPDF"`
    URL      string `json:"url"`
    FullPage bool   `json:"fullpage"`
}

type SMSData struct {
    AccountSID string `json:"accountSID"`
    AuthToken  string `json:"authToken"`
    ToNumber   string `json:"toNumber"`
    FromNumber string `json:"fromNumber"`
    Body       string `json:"body"`
}
```

System-account creation:

```go
stripeURL, err := backend.NewSystemAccount(email)

data, err := backend.NewSystemAccountBypassStripe(email, bypassFlag)
// backend.NewSystemAccountData{PublicKey, RootToken, AdminPassword}
```

```go
type NewSystemAccountData struct {
    PublicKey     string `json:"pk"`
    RootToken     string `json:"rootToken"`
    AdminPassword string `json:"pw"`
}
```

## Coding-Agent Notes

- Prefer package helpers and exported types over hand-built HTTP calls.
- Pass pointers for output values in DB reads and writes, for example `&task` or `&tasks`.
- `FindOne` and `SudoFindOne` intentionally fail unless exactly one document matches.
- `ListResult.Results` is set to the pointer passed in `v`; read your typed slice variable after the call.
- Normal DB APIs are account-scoped. Sudo DB APIs require the root token and bypass account scoping.
- Go exposes more admin/server helpers than the JS browser client: sudo DB, account root lookups, function management, cache/queue, forms, email, SMS, and system-account setup.
