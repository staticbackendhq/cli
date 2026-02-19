# StaticBackend Go Client Library

Go client wrapper for the StaticBackend API.

Module: `github.com/staticbackendhq/backend-go`

## Setup

```go
import backend "github.com/staticbackendhq/backend-go"

// Required before any call
backend.PublicKey = "your-public-key"
backend.Region = backend.RegionLocalDev // or RegionNorthAmerica1, or a custom URL

// For dev using the CLI you may use these hard coded keys
// public key: dev_memory_pk
// Root token: safe-to-use-in-dev-root-token

// Region constants
backend.RegionNorthAmerica1 // "na1" - managed hosting
backend.RegionLocalDev      // "dev" - localhost:8099
// For self-hosted: backend.Region = "https://your-server.com"

// Optional verbose HTTP logging
backend.Verbose = true
```

## Error Pattern

All functions return `error`. On success, error is `nil`. Results are populated via pointer arguments.

```go
token, err := backend.Register("user@example.com", "password")
if err != nil {
    log.Fatal(err)
}
// use token for all subsequent calls
```

## Authentication

```go
// Register - creates user, returns session token
token, err := backend.Register(email, password string) (string, error)

// Login - returns session token
token, err := backend.Login(email, password string) (string, error)

// Get current user
me, err := backend.Me(token string) (CurrentUser, error)
// CurrentUser fields: AccountID, UserID, Email, Role (int)

// Change password
err := backend.SetPassword(token, email, oldPassword, newPassword string) error

// Password reset flow
code, err := backend.GetPasswordResetCode(token, email string) (string, error)
err := backend.ResetPassword(email, code, password string) error
```

## Account Management

```go
// List all users in the account
users, err := backend.Users(token string) ([]CurrentUser, error)

// Add a user to the same account
user, err := backend.AddUser(token, email, password string) (CurrentUser, error)

// Remove a user (token must have higher role than removed user)
err := backend.RemoveUser(token, userID string) error

// Get a token for a specific accountID (root token required)
tok, err := backend.SudoGetToken(rootToken, accountID string) (string, error)
```

__This requires the root token__.

## Database Operations

All DB functions take `token` and `repo` (collection name). Documents always get `id` and `accountId` auto-populated.

### Filters

```go
// QueryOperator constants
backend.QueryEqual            // "=="
backend.QueryNotEqual         // "!="
backend.QueryLowerThan        // "<"
backend.QueryLowerThanEqual   // "<="
backend.QueryGreaterThan      // ">"
backend.QueryGreaterThanEqual // ">="
backend.QueryIn               // "in"
backend.QueryNotIn            // "!in"

// QueryItem struct
filter := backend.QueryItem{Field: "status", Op: backend.QueryEqual, Value: "active"}

// Sentinel errors
backend.ErrNoDocument       // FindOne returned 0 results
backend.ErrMultipleDocument // FindOne returned >1 results
```

### ListParams & ListResult

```go
params := &backend.ListParams{Page: 1, Size: 25, Descending: true}

// ListResult returned by List/Find/SudoList/SudoFind
type ListResult struct {
    Page     int
    PageSize int
    Total    int
    Results  interface{} // points to your slice
}
```

### Create

```go
// Create one document - v is populated with the created document (including id, accountId)
err := backend.Create(token, repo string, body interface{}, v interface{}) error

// Create multiple documents
ok, err := backend.CreateBulk(token, repo string, body interface{}) (bool, error)
```

### Read

```go
// List all documents
var tasks []Task
meta, err := backend.List(token, repo string, &tasks, params *ListParams) (ListResult, error)

// Get by ID
var task Task
err := backend.GetByID(token, repo, id string, &task) error

// Get multiple by IDs (avoids N+1 queries)
var tasks []Task
err := backend.GetByIDs(token, repo string, ids []string, &tasks) error

// Query with filters
filters := []backend.QueryItem{
    {Field: "done", Op: backend.QueryEqual, Value: false},
}
var tasks []Task
meta, err := backend.Find(token, repo string, filters, &tasks, params *ListParams) (ListResult, error)

// Find exactly one document
var task Task
err := backend.FindOne(token, repo string, filters []QueryItem, &task) error
// Returns ErrNoDocument or ErrMultipleDocument if not exactly one match

// Full-text search
var tasks []Task
err := backend.Search(token, repo, keywords string, &tasks) error

// Count documents matching filters (pass nil for all)
n, err := backend.Count(token, repo string, filters []QueryItem) (int64, error)
```

### Update

```go
// Update a document (can be partial - only changed fields needed)
var updated Task
err := backend.Update(token, repo, id string, body interface{}, &updated) error

// Bulk update matching filters
filters := []backend.QueryItem{{Field: "status", Op: backend.QueryEqual, Value: "pending"}}
n, err := backend.UpdateBulk(token, repo string, filters []QueryItem, body interface{}) (int, error)

// Increment/decrement a numeric field (n can be negative)
err := backend.Increase(token, repo, id, field string, n int) error
```

### Delete

```go
err := backend.Delete(token, repo, id string) error
err := backend.DeleteBulk(token, repo string, filters []QueryItem) error
```

### Sudo (Root Token) DB Operations

Sudo variants bypass account scoping - all documents across all accounts are accessible. Requires a root token.

```go
err := backend.SudoCreate(token, repo string, body interface{}, v interface{}) error
meta, err := backend.SudoList(token, repo string, v interface{}, params *ListParams) (ListResult, error)
err := backend.SudoGetByID(token, repo, id string, v interface{}) error
err := backend.SudoGetByIDs(token, repo string, ids []string, v interface{}) error
err := backend.SudoUpdate(token, repo, id string, body interface{}, v interface{}) error
meta, err := backend.SudoFind(token, repo string, filters []QueryItem, v interface{}, params *ListParams) (ListResult, error)
err := backend.SudoFindOne(token, repo string, filters []QueryItem, v interface{}) error
err := backend.SudoDelete(token, repo, id string) error

// List all collection names
names, err := backend.SudoListRepositories(token string) ([]string, error)

// Add a DB index on a field
err := backend.SudoAddIndex(token, repo, field string) error
```

## File Storage

```go
// Upload a file - returns id and public URL
file, err := os.Open("photo.jpg")
res, err := backend.StoreFile(token, filename string, file io.ReadSeeker) (StoreFileResult, error)
// StoreFileResult: {ID string, URL string}

// Download file content
buf, err := backend.DownloadFile(token, fileURL string) ([]byte, error)

// Delete a file by its ID
ok, err := backend.DeleteFile(token, id string) (bool, error)
```

## Email

```go
// Send a simple email
ok, err := backend.SendMail(token, from, fromName, to, subject, body, replyTo string) (bool, error)

// Send email with attachments
email := backend.EmailData{
    From:     "you@example.com",
    FromName: "Your Name",
    To:       "recipient@example.com",
    Subject:  "Hello",
    Body:     "<p>HTML body</p>",
    ReplyTo:  "reply@example.com",
    Attachments: []backend.Attachment{
        {URL: "https://example.com/file.pdf"},                           // fetched by server
        {Body: pdfBytes, ContentType: "application/pdf", Filename: "f.pdf"}, // inline bytes
    },
}
ok, err := backend.SendMailWithAttachments(token string, email EmailData) (bool, error)
```

__You specify either the URL or the (body, content type, filename) as attachement__.

## Cache & Work Queue

```go
// Cache (requires root token)
err := backend.CacheSet(token, key string, v interface{}) error
err := backend.CacheGet(token, key string, v interface{}) error

// Work queue - enqueue a string value
err := backend.QueueWork(token, key, value string) error

// Worker - polls every 5 seconds, call in a goroutine
go backend.WorkerQueue(token, key string, func(val string) {
    // process val
})
```

## Messaging

```go
// Publish a message to a channel
err := backend.Publish(token, channel, typ string, data interface{}) error
```

## Extra Features

```go
// Resize an image (PNG/JPG input → JPG output), returns StoreFileResult
res, err := backend.ResizeImage(token, filename string, file io.ReadSeeker, maxWidth float64) (StoreFileResult, error)

// Convert a public URL to PDF or PNG, returns StoreFileResult
res, err := backend.ConvertURLToX(token string, data ConvertParam) (StoreFileResult, error)
// ConvertParam: {ToPDF bool, URL string, FullPage bool}

// Send SMS via Twilio (root token required)
err := backend.SudoSendSMS(token string, data SMSData) error
// SMSData: {AccountSID, AuthToken, ToNumber, FromNumber, Body string}
```

## Server-side Functions

```go
// Add a function (triggered by a topic/channel message)
fn := backend.Function{FunctionName: "myFn", TriggerTopic: "chan", Code: "..."}
err := backend.AddFunction(token string, fn Function) error

// List all functions
fns, err := backend.ListFunctions(token string) ([]Function, error)

// Update a function
err := backend.UpdateFunction(token string, fn Function) error

// Delete a function by name
err := backend.DeleteFunction(token, name string) error

// Get function details including run history
fn, err := backend.FunctionInfo(token, name string) (Function, error)
```

## Forms

```go
// List form submissions (pass empty name for all forms)
data, err := backend.ListForm(token, name string) ([]map[string]interface{}, error)
```

## Low-level HTTP Helpers

Use these when calling custom or unlisted API endpoints:

```go
backend.Get(token, url string, v interface{}) error
backend.Post(token, url string, body interface{}, v interface{}) error
backend.Put(token, url string, body interface{}, v interface{}) error
backend.Del(token, url string) error
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"

    backend "github.com/staticbackendhq/backend-go"
)

type Task struct {
    ID        string `json:"id"`
    AccountID string `json:"accountId"`
    Title     string `json:"title"`
    Done      bool   `json:"done"`
}

func main() {
    backend.PublicKey = "your-public-key"
    backend.Region = backend.RegionLocalDev

    token, err := backend.Register("user@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }

    task := Task{Title: "Buy milk"}
    if err := backend.Create(token, "tasks", task, &task); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Created:", task.ID)

    filters := []backend.QueryItem{
        {Field: "done", Op: backend.QueryEqual, Value: false},
    }
    var tasks []Task
    meta, err := backend.Find(token, "tasks", filters, &tasks, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%d pending tasks (total: %d)\n", len(tasks), meta.Total)
}
```
