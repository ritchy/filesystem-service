# filesystem.io CLI

A command-line interface for [filesystem.io](https://filesystem.io) written in Go.

## Commands

| Command | Description |
|---|---|
| `fs login` | Authenticate with your filesystem.io account |
| `fs list` | List files and folders in the root directory |
| `fs list /folder` | List files in a specific folder |
| `fs list /folder/sub` | List files in a nested folder |
| `fs download /file.txt` | Download a file to the current directory |
| `fs download /folder/file.pdf -o ~/out.pdf` | Download a file to a specific path |

## Prerequisites

### 1. Enable USER_PASSWORD_AUTH in Cognito

`fs login` uses the `USER_PASSWORD_AUTH` flow. This must be explicitly enabled
on the Cognito User Pool App Client:

1. Open the [AWS Console](https://console.aws.amazon.com/)
2. Navigate to **Cognito → User Pools → `us-east-1_gI0gO1dL0`**
3. Go to **App clients** and select the client `7a469p47n22ru33dui20eekclj`
4. Click **Edit** → **Authentication flows**
5. Check **ALLOW_USER_PASSWORD_AUTH** and save

### 2. Go 1.21+

```sh
go version   # should be 1.21 or higher
```

## Build

```sh
cd cli

# Build as 'fs'
go build -o fs .

# Or build as 'filesystem'
go build -o filesystem .

# Optionally install to $GOPATH/bin
go install .
```

## Usage

### Login

```sh
./fs login
```

You will be prompted for your email and password. Credentials are saved to
`~/.filesystem/credentials.json` (permissions `0600`).

### List root directory

```sh
./fs list
```

Example output:
```
  Path: / (root)
  ────────────────────────────────────────────────────────────────────
  NAME                                      TYPE    SIZE
  ────────────────────────────────────────────────────────────────────
  Documents/                                folder           -
  Photos/                                   folder           -
  readme.txt                                file         1.2 KB

  3 item(s)
```

### List a specific folder

```sh
./fs list /Documents
./fs list /Documents/Work
```

## Project Structure

```
cli/
├── main.go                         # entry point
├── go.mod / go.sum                 # module and dependency lock
├── amplify_outputs.json            # Amplify / AWS configuration
├── cmd/
│   ├── root.go                     # cobra root command
│   ├── login.go                    # 'fs login' command
│   └── list.go                     # 'fs list' command
└── internal/
    ├── auth/
    │   ├── cognito.go              # Cognito USER_PASSWORD_AUTH login
    │   └── token.go                # JWT sub extraction (no external lib)
    ├── config/
    │   └── credentials.go          # credential persistence (~/.filesystem/)
    └── api/
        ├── client.go               # AppSync GraphQL HTTP client
        └── filesystem.go           # Member / FileFolder / File queries
```

## How it Works

1. **`fs login`** calls `InitiateAuth` (USER_PASSWORD_AUTH) on the Cognito User
   Pool and stores the resulting JWT tokens in `~/.filesystem/credentials.json`.

2. **`fs list [path]`**:
   - Decodes the stored ID token to extract the Cognito `sub` (user ID)
   - Queries AppSync GraphQL for the `Member` record matching that user ID
   - Reads the `rootFileId` from the associated `FileFolder`
   - If a path is given, navigates the tree by resolving each path component
     against the current folder's children until the target folder is reached
   - Prints the children of the target folder

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/aws/aws-sdk-go-v2/aws` | AWS SDK core types |
| `github.com/aws/aws-sdk-go-v2/config` | AWS SDK config loading |
| `github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider` | Cognito auth |
| `github.com/spf13/cobra` | CLI framework |
| `golang.org/x/term` | Secure password input (no echo) |
