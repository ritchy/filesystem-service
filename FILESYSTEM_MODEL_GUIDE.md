# FileSystem Data Model Guide

## Overview
This guide explains how to use the FileFolder and File models in your AWS Amplify application.

## Data Structure

### FileType Enum
- `file` - Represents a file
- `folder` - Represents a folder that can contain other files/folders

### FileFolder Model
The root container for a filesystem structure.

**Properties:**
- `id` (ID, required) - Unique identifier
- `name` (String, required) - Name of the root folder
- `createdDate` (DateTime, required) - When the folder was created
- `lastUpdatedDate` (DateTime, required) - When the folder was last modified
- `files` (HasMany File) - Direct children files/folders

### File Model
Represents both files and folders in a hierarchical structure.

**Properties:**
- `id` (ID, required) - Unique identifier
- `name` (String, required) - Name of the file/folder
- `type` (FileType, required) - Either 'file' or 'folder'
- `size` (Integer, default: 0) - Size in bytes (for files)
- `text` (String, optional) - Text content for text files
- `fileReference` (String, optional) - S3 reference for binary files
- `createdDate` (DateTime, required) - When created
- `lastUpdatedDate` (DateTime, required) - When last modified
- `fileFolderId` (ID, optional) - Reference to parent FileFolder
- `parentFileId` (ID, optional) - Reference to parent File (for nested structure)

**Relationships:**
- `fileFolder` (BelongsTo FileFolder) - The root folder this file belongs to
- `parentFile` (BelongsTo File) - The parent folder if nested
- `childFiles` (HasMany File) - Children if this is a folder

## Usage Examples

### Creating a Root FileFolder

```typescript
import { generateClient } from "aws-amplify/data";
import type { Schema } from "@/amplify/data/resource";

const client = generateClient<Schema>();

// Create a root folder
const rootFolder = await client.models.FileFolder.create({
  name: "My Documents",
  createdDate: new Date().toISOString(),
  lastUpdatedDate: new Date().toISOString(),
});
```

### Adding Files to the Root Folder

```typescript
// Add a folder at root level
const codeFolder = await client.models.File.create({
  name: "Code",
  type: "folder",
  fileFolderId: rootFolder.data.id,
  createdDate: new Date().toISOString(),
  lastUpdatedDate: new Date().toISOString(),
});

// Add a text file at root level
const textFile = await client.models.File.create({
  name: "readme.txt",
  type: "file",
  size: 1024,
  text: "This is the content of my text file",
  fileFolderId: rootFolder.data.id,
  createdDate: new Date().toISOString(),
  lastUpdatedDate: new Date().toISOString(),
});

// Add a binary file at root level (stored in S3)
const imageFile = await client.models.File.create({
  name: "photo.jpg",
  type: "file",
  size: 2048000,
  fileReference: "s3://my-bucket/photos/photo.jpg",
  fileFolderId: rootFolder.data.id,
  createdDate: new Date().toISOString(),
  lastUpdatedDate: new Date().toISOString(),
});
```

### Creating Nested Folders and Files

```typescript
// Add a file inside the Code folder
const jsFile = await client.models.File.create({
  name: "app.js",
  type: "file",
  size: 5120,
  text: "console.log('Hello World');",
  parentFileId: codeFolder.data.id, // This file is inside the Code folder
  fileFolderId: rootFolder.data.id,
  createdDate: new Date().toISOString(),
  lastUpdatedDate: new Date().toISOString(),
});

// Create a nested folder
const srcFolder = await client.models.File.create({
  name: "src",
  type: "folder",
  parentFileId: codeFolder.data.id, // Nested inside Code folder
  fileFolderId: rootFolder.data.id,
  createdDate: new Date().toISOString(),
  lastUpdatedDate: new Date().toISOString(),
});

// Add a file in the nested folder
const componentFile = await client.models.File.create({
  name: "Component.tsx",
  type: "file",
  size: 3072,
  text: "export const Component = () => { return <div>Hello</div>; }",
  parentFileId: srcFolder.data.id, // Inside src folder
  fileFolderId: rootFolder.data.id,
  createdDate: new Date().toISOString(),
  lastUpdatedDate: new Date().toISOString(),
});
```

### Querying the File Structure

```typescript
// Get a FileFolder with all its direct children
const { data: folder } = await client.models.FileFolder.get({
  id: rootFolder.data.id,
});

const { data: rootFiles } = await folder.files(); // Get all root-level files

// Get a folder and its children
const { data: codeFiles } = await codeFolder.childFiles(); // Get all files in Code folder

// Get a file's parent
const { data: parent } = await jsFile.parentFile();
```

### Filtering Files

```typescript
// Get all folders at root level
const { data: folders } = await client.models.File.list({
  filter: {
    fileFolderId: { eq: rootFolder.data.id },
    type: { eq: "folder" },
  },
});

// Get all text files
const { data: textFiles } = await client.models.File.list({
  filter: {
    type: { eq: "file" },
    text: { attributeExists: true },
  },
});

// Get all binary files (with fileReference)
const { data: binaryFiles } = await client.models.File.list({
  filter: {
    type: { eq: "file" },
    fileReference: { attributeExists: true },
  },
});
```

## Best Practices

1. **Text vs Binary Files**: 
   - Use the `text` field for text files (up to DynamoDB's item size limit of 400KB)
   - Use `fileReference` with S3 URLs for larger or binary files

2. **Maintaining Hierarchy**:
   - Always set `fileFolderId` to reference the root FileFolder
   - Set `parentFileId` for nested files/folders

3. **Updating Dates**:
   - Always update `lastUpdatedDate` when modifying a file
   - Consider cascading updates to parent folders

4. **Size Tracking**:
   - Set `size` to 0 for folders
   - Update folder sizes by aggregating children sizes

## Schema Diagram

```
FileFolder (Root Container)
├── files (HasMany File)
    │
    └── File
        ├── type: 'file' or 'folder'
        ├── text: (for text files)
        ├── fileReference: (for binary files)
        ├── parentFile (BelongsTo File) - for nesting
        └── childFiles (HasMany File) - if type is 'folder'
```
