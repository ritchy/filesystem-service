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

`POST /upload`: 'id' is provided as a query parameter in the POST uri. The body of the request is a multipart form.

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

This function always returns an SVG file based on the 'size' value. If the size is 'big',
generate a general text icon, 'txt.svg', to return.

if the size requested is 'small', it returns a generated 'txt.svg' file with a smaller size.


## Create UI

Create UI
npm create vite@latest my-react-app --template react
npm install
npm run dev

## Next Steps

- Add Dynamo Resources (1 file system document per login)
- Add API via amplify (Gateway/lambda)
- Add S3 integration for actual file storage
