---
name: filesystemio
title: Filesystem features to upload and manage files.
description: A cloud based filesystem skill with help to show usage patterns and the commands: list, upload, download, delete, rename and share files through filesystem.io
homepage: https://filesystem.io
version: 0.1
author: bytestream.io
license: MIT
dependencies: []
platforms: [linux, macos]
metadata:
{
    "openclaw":
      {
        "emoji": "📂",
        "requires": { "bins": ["fs"] },
        "install":
          [
            {
              "id": "go",
              "kind": "go",
              "module": "github.com/filesystem.io/cli/cmd/fs@latest",
              "bins": ["fs"],
              "label": "Install fs (go)",
            },
          ],
      },
  }
---

# FileSystem cli

When the user asks to 'fs' or 'filesystem' anything, like 'fs list' or 'filesystem list'. 
Use `fs` to manage filesystem.io files directly from the terminal. Create, view, edit, delete, search, move files between folders. There also a help command to list all available commands and their usage.

Setup

- Install (download): `https://filesystem.io/cli/download/macos/fs -o /usr/local/bin/fs && chmod +x /usr/local/bin/fs`
- macOS-only; if prompted, grant Automation access to fs app.

Help Commands

- General help: `fs help`
- Command-specific help: `fs <command> --help`  
  - Example: `fs list --help` for listing options.

List Files

- List files: `fs list`
   - Lists files in the root directory.
- List files in folder: `fs list /folder`
   - Lists files in a specific folder.
- List files in nested folder: `fs list /folder/sub`
   - Lists files in a nested folder.

Delete Files

- Delete a file: `fs delete /file`
  - Interactive selection of file to delete.

Move Files

- Move file to folder: `fs move /folder/file /new-folder`
- Move file to nested folder: `fs move /folder/file /folder/sub-folder`
- Move folder to new location: `fs move /folder/sub-folder /new-folder/sub-folder`

Rename Files

- Rename a file: `fs rename /folder/file /folder/new-file-name`

Download Files

- Download a file: `fs download /file`
  - Downloads selected file.
- Download with output path: `fs download /file -o /path/to/destination`
  - Downloads selected file to a specified location. 

Limitations

- Interactive prompts may require terminal access.

Notes

- Requires 'fs' or 'filesystem' to be in PATH.
- For automation, grant permissions in System Settings > Privacy & Security > Automation.

# metadata-hermes:
#   tags: [Files, filesystem, share]
#   category: files
#   related_skills: []
#   requires_toolsets: [terminal, fs]
