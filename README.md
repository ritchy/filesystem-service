# FileSystem.io Service

# Setup

## Amplify/React/Vite quick start
https://docs.amplify.aws/react/start/quickstart/


## Manual Process 

mkdir filesystem-service;cd filesystem-service
npm create amplify@latest
npx ampx sandbox
node errors -> 'nvm use 22'
npx ampx sandbox

## Deploy React App

        - mkdir dist
        - cp amplify_outputs.json dist/amplify_outputs.json
        - cp amplify_outputs.json client/src/amplify_outputs.json
        - cd client
        - npm install
        - BUILD_PATH=../dist npm run build

## Amplify.yml
version: 1
backend:
  phases:
    build:
      commands:
        - npm ci --cache .npm --prefer-offline
        - npx ampx pipeline-deploy --branch $AWS_BRANCH --app-id $AWS_APP_ID
frontend:
  phases:
    build:
      commands:
        - mkdir dist
        - cp amplify_outputs.json dist/amplify_outputs.json
        - cd client
        - npm install
        - BUILD_PATH=../dist npm run build
  artifacts:
    baseDirectory: dist
    files:
      - '**/*'
  cache:
    paths:
      - .npm/**/*

## Start tracking via git

git init
git add README.md
git add package.json
git add .gitignore
git add amplify

## Create hard-coded API/Gateway/Lambda for handling files

- AI generated prompt:

This is an AWS Amplify v2 project. I'd like to add an AWS API Gateway deployment entry for a GET request of '/files' and associated lambda definition to handle the request. I'd like the lambda to be in typescript or javascript language that returns the following hard-coded JSON document: 
[
    {
      id: '/Code',
      date: new Date(2023, 11, 2, 17, 25),
      type: 'folder',
    },
    {
      id: '/Music',
      date: new Date(2023, 11, 1, 14, 45),
      type: 'folder',
    },
    {
      id: '/Music/Animal_sounds.mp3',
      size: 1457296,
      date: new Date(2023, 11, 1, 14, 45),
      type: 'file',
    }
]

- Once AI completed, deploy your changes: `npx ampx sandbox`

- Once deployed, you'll get the API URL in the outputs which you can use to call: `GET {API_URL}/files`

## Amplify configuration during development will differ from production.

### Development 

import outputs from "../../../amplify_outputs.json";
Amplify.configure(outputs);

### Production

Amplify.configure(
  {
    API: {
      GraphQL: {
        endpoint: process.env.AMPLIFY_DATA_GRAPHQL_ENDPOINT || '',
        region: process.env.AWS_REGION || 'us-east-1',
        defaultAuthMode: 'identityPool',
      },
    },
  },
  {
    Auth: {
      credentialsProvider: {
        getCredentialsAndIdentityId: async () => ({
          credentials: {
            accessKeyId: process.env.AWS_ACCESS_KEY_ID || '',
            secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY || '',
            sessionToken: process.env.AWS_SESSION_TOKEN,
          },
        }),
        clearCredentialsAndIdentityId: () => {
          /* noop */
        },
      },
    },
  }
);

## Create Dynamo model entries representing the root folder and subsequent filesystem

add a filesystem folder model with nested file structure to this AWS Amplify data schema.

in this current aws amplify project, I'd like to add a new model that represents a filesystem folder; similar to following: 

  FileFolder: a.model({
    id: a.id().required(),
    createdDate: a.datetime().required(),
    lastUpdatedDate: a.datetime().required(),
  }),

 In addition to above properties, I'd like to add the ability to contain file meta-data where this "FileFolder" contains children of 'Files'. Each child can be of enum type 'file' or 'folder' and 'folder' types can contain more 'Files'.
 
 Some key properties to these "Files" are:

    size: a.integer().default(0),
    text: a.string().required(),
    fileReference: a.string(),

 If the file is a text file, then I'd also like to store the file data in this Dynamo document. Otherwise, I just want to store a link to the file data in the string property named 'fileReference". 


## Next, replace hard-coded response with values in Dynamo document representing the root folder

### add the first GET /files functionality to introduce all the dependencies and get it working with Dynamo

I'd like to update the files-handler API and associated lambda function to remove the existing hard-coded JSON response and leverage the new models, "FileFolder" and "File" to handle the following HTTP/REST requests:

`GET {API_URL}/files`: returns same JSON structure as current hard-coded return, but using the provided File and FileFolder models defined in resource.ts file. The return JSON structure for '/files' request is from the 'files' property of FileFolder and returned in same structure as currently:

[
    {
      id: '/Code',
      date: lastUpdatedDate,
      type: 'folder',
    },
    {
      id: '/Music',
      date: lastUpdatedDate,
      type: 'folder',
    },
]

if the 'files' property of FileFolder is empty, I'd like to pre populate with sample data: one sample text file named 'sample.txt' with a size: 8 and a 'text' value of: 'This is a sample text file'. Also, I'd like to pre-populate with  one sample folder named 'sample'.

### GET /info

update the files-handler API and associated lambda function to handle GET /info request:

`GET {API_URL}/info`

It should return a hard coded response with the following JSON document as the body:

{
    "free": 982929299222,
    "total": 1995218165760,
    "used": 1067712249856
}

### /direct?id=%2Fsample.txt

Add a new function to the file-handler lambda to handle the following GET request: `GET /direct?id={id}`.
This finds the associated File entry with the matching id and returns the 'text' value or an empty string
if it's missing.

## update `GET /direct?id={id} to handle non-text files located in S3

I'd like to update the lambda associated with the endpoint: `GET /direct?id={id}. It currently 
returns the value of the 'text' property after finding the file associated with the provided 'id'.

If there is no value in the 'fileReference' property, I'd like to return same text as it currently does.

If there is a value in the 'fileReference' property, I'd like to send a redirect with a direct link
to a pre-signed URL of the file in S3 using the 'fileReference' as the path to the defined S3 
bucket in this project.

This should result in a direct download of the file data.

//npm install @aws-sdk/s3-request-presigner

### now add the handler for creating folders via a POST method

Add a new function to the file-handler lambda to handle the following POST request: `POST /files/{id}`: 

This creates a new File entry of type folder from the 'id' provided in the uri and the 2 properties in the provided body of the POST containing a JSON document with the following structure:
{
  name: "folder name",
	type: "folder"
}

The entry is created using the following standard create from File model specified in the resource.ts file, similar to this:

await client.models.File.create({
  name: "Code",
  type: "folder",
  fileFolderId: parentFolder.data.id,
  createdDate: new Date().toISOString(),
  lastUpdatedDate: new Date().toISOString(),
});

### -- NOT YET -- Handle File Uploads: `POST /upload?id=xx`

Create a new endpoint and handle to process the following POST method: `POST /upload?id=123`.

As part of this we need a new amplify storage definition to create an S3 bucket to store 
files when this endpoint is called.

'id' is provided as a query parameter in the POST uri and associated with the 'File'
associated with the data uploaded. The property, 'fileReference' will contain the the S3
reference associated with file upload


### Rename a file or folder PUT /files/{id}

Create a new lambda function to handle the following PUT request: `PUT /files/{id}`: 
When executed, it finds the 'File' database entry with matching id and updates the model 'name' property.
The body of the 'PUT is a JSON document with the following structure:
 {
    operation: "rename"
    name: "new file name"
    target: "string" // not used
    ids:["string"] // not used here
}

if successful, this returns a successful status code and JSON body including the id and name of the file that was updated: {id: "id", name: "name"}

### DELETE files

Create a new lambda function to handle the following DELETE method request: `DELETE "/files"`:
The body of this DELETE request contains a JSON document containing an array with the name 'ids' which reference the id's of multiple 'File' model entries which should be deleted from Dynamo.
The body of the 'DELETE is a JSON document with the following structure which is same as PUT:

 {
    operation: "ignore" // not used
    name: "ignore" // not used
    target: "string" // not used
    ids:["1", "2"]
}

### GET "/info/{id}"

Create a new lambda function to handle the following GET method request: `GET "/info/{id}"`:
This function returns the size and count of the provided File model. It will first retrieve the File
based on the provided 'id'. If the file type is 'file', then it will return a count of '1' and the value
of the 'size' property of the model in the body of the response as a JSON document. Here is the structure of the JSON
document:

{
    count: "1",
    size: "123456"
}

If the file type is 'folder', then it will recursively retrieve the 'files' 
property of a folder and count the total number of ancestor file entries of the original folder as well
as the sum of all file 'size' properties. This recursive function will only count entries of type 'file' 
and NOT counting entries of type 'folder'.

In the body of the response, it will return a JSON document with the following format:

{
    count: "123",
    size: "123456"
}

### GET "/preview"

### GET "/icons/{size}/{name} --> /icons/big/txt.svg
Create a new, separate lambda function, not using the current files-handler lambda, but a completely new one to return SVG format handle the following GET method request: `GET "/icons/{size}/{name}"`:

-This function always returns an SVG file based on the 'size' value. If the size is 'big',
generate a general text icon, 'txt.svg', to return.

if the size requested is 'small', it returns a generated 'txt.svg' file with a smaller size.

## update "icons/{size}/{name}

update the icons-handler, associated with the `/info/{id}` endpoint, to return a folder icon when the name parameter is 'folder.svg'. If the name path parameter is 'folder.svg', return an new SVG response of a folder that matches the current style of the 'txt.svg', otherwise return the existing txt icon.

## Create UI

Create a REACT UI based app with a file explorer interface that leverages the endpoints of this project.

There are 3 columns in the UI:

The left most column is a tree starting at the root FileFolder model and lists the 'file' property elements
as it's children. Each 'file' element, in turn, has children, so this is probably a recursive traversal to
to render this left most column panel. It starts out 1 level deep.

The middle column displays the children of whatever is selected in the left most column. It is not a tree, but
simply a list of 'file' elements of whatever is selected on the left most column.

The third and last column displays the results from calling the endpoint: GET "/info/{id}", of any 
selected item in the middle column.

Above the 3 columns there's a panel that spans all 3. On the left, there is a search bar to filter
a file or folder based on the name. There is a results panel that is shown as soon as 2 or more 
characters are typed in the search bar. This search results panel overlay's the 3 columns and this 
panel will disappear when the search panel is cleared. There should a button to the right of the 
search panel with an 'x' to clear the search text and restore the 3 columns.

In the same panel as the search field, on the right side, there is also a toggle button to hide or display the last, right most column showing the 'info' of items selected in the middle column.

When you right-click any selected item in column 1 or 2, a context menu pops up allowing you to 
 - rename the item
 - delete the item 
 - create a folder as a child of the item
 - create a file item with a 'name' and textarea associated with the 'text' field of the file item

### UI client tweeks

update the React app in the 'client' folder to handle double-clicking. 

When you double-click a File of type 'folder', it will have that File selected in the 1st 
column tree of the UI. It will also result in listing the elements of the 'files' property in the 2nd column. 

When you double-click a File of type 'file', the result will be a new browser window with the
URL pointing to the direct endpoint: "{API_URL}/direct?id={id}" and showing the raw results of
that endpoint.

### UI Client tweeks

update the React app in the 'client' folder to handle the following 2 scenarios:

double-clicking a File of type 'file' in the 1st column tree will behave same as double clicking a file
in the 2nd column. double-clicking a File of type 'file' should result a new browser window with the
URL pointing to the direct endpoint: "{API_URL}/direct?id={id}" and showing the raw results of
that endpoint.

the tree in the 1st column should automatically expand the tree to whatever File is selected.
So, double-clicking a folder in the 2nd column should trigger the tree to open up and show
the selected folder with it's files elements shown in the 2nd column.

### Add UI Upload capability

Add a file upload capability to the client UI REACT app. Make the 2nd column in the UI a drop zone that accepts
a drop from a file in the local filesystem and pops up an upload dialog with a name field and upload button. 
The name of the dropped file is pre-populated in a name text field, but can be changed before upload 
button is selected.

For this, we need to add a new amplify storage definition to the filesystem-service amplify project. This new storage definition specifies an S3 bucket to store uploaded files in a '/files' folder. 

When the upload is complete, we then also create a new File entry:

The 'name' field of the new File entry matches the name from the upload dialog. 
the 'size' field is based on the size in bytes of the file uploaded.
The 'fileReference' value is populated with the S3 'path' provided in the upload call to Amplify.Storage.uploadData
The 'parentFile' property is based the currently selected File entry:
  - If the currently selected File is of type 'folder', it will serve as the 'parentFile'
  - If the currently selected File is of type 'file', then the 'parentFile' of the new File is
    the same as the 'parentFile' of the selected file.

Once the upload and File creation is complete, the UI is refreshed to show the new file in the 1st column 
tree and the new file is auto-selected in the middle, 2nd column. The selection should also trigger
a refresh to the 3rd 'info' column.

### Add UI upload progress

When a file is dropped to the middle column of this app, an upload dialog appears. I'd like to enhance
this dialog to show upload progress. Amazon S3 documentation shows you can monitor progress
in the following manner:

import { uploadData } from 'aws-amplify/storage';

const monitorUpload = async () => {
  try {
    const result = await uploadData({
      path: "album/2024/1.jpg",
      // Alternatively, path: ({identityId}) => `album/${identityId}/1.jpg`
      data: file,
      options: {
        onProgress: ({ transferredBytes, totalBytes }) => {
          if (totalBytes) {
            console.log(
              `Upload progress ${Math.round(
                (transferredBytes / totalBytes) * 100
              )} %`
            );
          }
        },
      },
    }).result;
    console.log("Path from Response: ", result.path);
  } catch (error) {
    console.log("Error : ", error);
  }
}

I'd like to add a progress bar in the upload dialog showing upload progress.

### Client UI parent select

I'd like to make a change to the client UI REACT app. In the middle column of the UI, the 2nd
column, I'd like to optionally show the first item in the list with " .. " as a the label.
If the currently selected file of type 'folder' has a parent, I'd like to show that ".." and, when
double-clicked, the result is a selectoion of the parent of the currently selected folder.
If the current folder has no parent, then don't show that ".." in the list.

### Login landing page

I'd like to add a login landing page to the client UI React app using the Authenticator component from the Amplify UI library. Once logged in, I'd like to keep the user logged in as long as possible, renewing a login token
every time they return.

As part of this, I'd like to update the root storage folder location, based on the unique user id once they log in.
So this would be the existing S3 bucket already created with this application would then have a
structure of "{bucket}/userID" as the starting point folder for the UI.

### default setup

in the amplify app, I'd like to associate the FileFolder and File models with the current user so every user that logs in gets their own set of results.

in the client UI React app,if the root folder is empty, I'd like to create at least one child file type of 'folder' with the name 'files'

### Top toolbar with logo and profile

I'd like to add a thin header panel at the very top of the client UI React App with:

 - the a branding logo in the top left using the following asset: src/assets/fs-letters-glass-small.png
 - in the top right, generate a new "profile" button image as a head outline, along with the currently logged in email address, both serving as a single drop down button showing a "Sign Out" option.

remove the current Sign Out" button in the panel below this new one
remove the current email address label in this panel below this new one

### customize the login page

I want to customize the React client app login page that Amplify provides. I'd like to
make the sign in button black. I'd like to place a the text "filesystem.io" above the login
box with some spacing from the top of the page.

I'd like to remove the ability to "Create Account" on the sign in page and just have the option to log in


### Upload File changes

In the React client app, I'd like to add an "Action" drop-down to the middle column which mirrors the context menu that pops up when you right-click a file.

I'd also like to add an action to this pop-up menu to "upload file". The result of selecting this action should
leverage the exiting "upload file" dialog with a file browsing option

I'd like the new "Action" button to appear any time a File is selected, even in the tree of the first column

I'd like this new action button to be transparent, with a black border and black text.

I'd like the profile button at the top right to match the other buttons in color, but stay rounded as it is currently. Transparent with a black border and black text.

I'd like to repurpose the "upload file" dialog to include a "browse" button to the right of the file name text.
This button needs a standard browsing icon to reduce the size.

I'd like to update the "action" menu as well as the right-click context menu to include "Upload File"
which results in the "upload file" dialog with no file selected.

#### TASK COMPLETED

I've successfully added "Upload File" to both the Action menu and right-click context menu!

## What Was Implemented

### 1. **Action Menu - Upload File**
- **New handler**: `handleActionUploadFile()` - Opens upload dialog with no file pre-selected
- **Menu location**: Bottom of Action dropdown (after Create File)
- **Smart parent logic**: Uses selected folder or parent folder

### 2. **Context Menu - Upload File**
- **New handler**: `handleUploadFile()` - Opens upload dialog with no file pre-selected
- **Menu location**: Bottom of context menu (after Create File)
- **Same smart logic**: Determines correct parent folder automatically

### 3. **Upload Dialog Behavior**
When opened from these menus:
- **No file selected initially** - Empty upload dialog
- **Browse button ready** - Users click to select file
- **Parent folder determined** - Automatically uses the correct folder
- **Same workflow** - Consistent with drag-and-drop uploads

## Usage

Users can now upload files via:
1. **Drag & drop** - Drag file onto middle column (with pre-selected file)
2. **Action menu** - Click Action → Upload File (no pre-selected file)
3. **Context menu** - Right-click any file/folder → Upload File (no pre-selected file)

All three methods open the same upload dialog with browse functionality!

### Manual steps to create react app

npm create vite@latest my-react-app --template react
npm install
npm run dev

## Next Steps

- Add Dynamo Resources (1 file system document per login)
- Add API via amplify (Gateway/lambda)
- Add S3 integration for actual file storage
