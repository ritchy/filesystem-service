import { type ClientSchema, a, defineData } from '@aws-amplify/backend';

/*== STEP 1 ===============================================================
The section below creates a Todo database table with a "content" field. Try
adding a new "isDone" field as a boolean. The authorization rule below
specifies that any unauthenticated user can "create", "read", "update", 
and "delete" any "Todo" records.
=========================================================================*/
const schema = a.schema({

  Todo: a
    .model({
      content: a.string(),
    })
    .authorization((allow) => [allow.guest()]),

  // Enum for file type
  FileType: a.enum(['file', 'folder']),

  // Root filesystem folder container
  FileFolder: a
    .model({
      id: a.id().required(),
      name: a.string().required(),
      createdDate: a.datetime().required(),
      lastUpdatedDate: a.datetime().required(),
      files: a.hasMany('File', 'fileFolderId'),
    })
    .authorization((allow) => [allow.guest()]),

  // File model - can represent both files and folders
  File: a
    .model({
      id: a.id().required(),
      name: a.string().required(),
      type: a.ref('FileType').required(),
      size: a.integer().default(0),
      text: a.string(), // Store text content for text files
      fileReference: a.string(), // Store S3 reference for binary files
      createdDate: a.datetime().required(),
      lastUpdatedDate: a.datetime().required(),
      
      // Relationship to parent FileFolder (root level)
      fileFolderId: a.id(),
      fileFolder: a.belongsTo('FileFolder', 'fileFolderId'),
      
      // Self-referential relationship for nested structure
      // If this file is of type 'folder', it can have children
      parentFileId: a.id(),
      parentFile: a.belongsTo('File', 'parentFileId'),
      childFiles: a.hasMany('File', 'parentFileId'),
    })
    .authorization((allow) => [allow.guest()]),

});

export type Schema = ClientSchema<typeof schema>;

export const data = defineData({
  schema,
  authorizationModes: {
    defaultAuthorizationMode: 'identityPool',
  },
});

/*== STEP 2 ===============================================================
Go to your frontend source code. From your client-side code, generate a
Data client to make CRUDL requests to your table. (THIS SNIPPET WILL ONLY
WORK IN THE FRONTEND CODE FILE.)

Using JavaScript or Next.js React Server Components, Middleware, Server
Actions or Pages Router? Review how to generate Data clients for those use
cases: https://docs.amplify.aws/gen2/build-a-backend/data/connect-to-API/
=========================================================================*/

/*
"use client"
import { generateClient } from "aws-amplify/data";
import type { Schema } from "@/amplify/data/resource";

const client = generateClient<Schema>() // use this Data client for CRUDL requests
*/

/*== STEP 3 ===============================================================
Fetch records from the database and use them in your frontend component.
(THIS SNIPPET WILL ONLY WORK IN THE FRONTEND CODE FILE.)
=========================================================================*/

/* For example, in a React component, you can use this snippet in your
  function's RETURN statement */
// const { data: todos } = await client.models.Todo.list()

// return <ul>{todos.map(todo => <li key={todo.id}>{todo.content}</li>)}</ul>
