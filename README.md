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

## 

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


## Next, replace hard-coded response with values in Dynamo document representing the root folder

add a filesystem folder model with nested file structure to your AWS Amplify data schema. Let me first examine your current data schema:

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


##


Create UI
npm create vite@latest my-react-app --template react
npm install
npm run dev

## Next Steps

- Add Dynamo Resources (1 file system document per login)
- Add API via amplify (Gateway/lambda)
- Add S3 integration for actual file storage
