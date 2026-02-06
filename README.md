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

I'd like to update the files-handler API and associated lambda function to remove the existing hard-coded JSON response and leverage the new models, "FileFolder" and "File" to handle the following HTTP/REST requests:

`GET {API_URL}/files`: returns same JSON structure as current hard-coded return, but using the provided File and FileFolder models defined in resource.ts file. The return JSON structure is as follows:
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
]

`POST /files/{id}`: Creates a new File entry of type folder from the 'id' provided in the uri and the 2 properties in the provided body containing a JSON document with the following structure:
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



`PUT /files/{id}`: Finds the 'File' database entry with matching id and updates the model 'name' property.
The body of the 'PUT is a JSON document with the following structure:
 {
    operation: "rename"
    name: "new file name"
    target: "string" // not used
    ids:["string"] // not used here
}

`DELETE "/files"`: The body of this DELETE request contains an array of 'id's which reference 'File' model
entries which should be deleted from Dynamo. The body of the 'DELETE is a JSON document with the following structure which is same as PUT:

 {
    operation: "ignore" // not used
    name: "ignore" // not used
    target: "string" // not used
    ids:["1", "2"]
}


## Create UI

Create UI
npm create vite@latest my-react-app --template react
npm install
npm run dev

## Next Steps

- Add Dynamo Resources (1 file system document per login)
- Add API via amplify (Gateway/lambda)
- Add S3 integration for actual file storage
