import type { APIGatewayProxyHandler } from 'aws-lambda';
import { Amplify } from 'aws-amplify';
import { generateClient } from 'aws-amplify/data';
import type { Schema } from '../../data/resource';

// Configure Amplify
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

const client = generateClient<Schema>();

export const handler: APIGatewayProxyHandler = async (event) => {
  console.log('event', event);

  // Handle PUT /files/{id} endpoint (rename operation)
  if (event.httpMethod === 'PUT' && (event.resource === '/files/{id}' || event.path?.match(/\/files\/[^/]+$/))) {
    try {
      // Get the file id from path parameters
      const fileId = event.pathParameters?.id;

      if (!fileId) {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Missing id parameter in path' }),
        };
      }

      // Parse the request body
      if (!event.body) {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Missing request body' }),
        };
      }

      const body = JSON.parse(event.body);
      const { operation, name } = body;

      if (operation !== 'rename') {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Only rename operation is supported' }),
        };
      }

      if (!name) {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Missing required field: name' }),
        };
      }

      // Update the File entry
      const now = new Date().toISOString();
      const { data: updatedFile } = await client.models.File.update({
        id: fileId,
        name,
        lastUpdatedDate: now,
      });

      if (!updatedFile) {
        return {
          statusCode: 404,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'File not found' }),
        };
      }

      return {
        statusCode: 200,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          id: updatedFile.id,
          name: updatedFile.name,
        }),
      };
    } catch (error) {
      console.error('Error updating file:', error);
      return {
        statusCode: 500,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          error: 'Internal server error',
          message: error instanceof Error ? error.message : 'Unknown error',
        }),
      };
    }
  }

  // Handle POST /files/{id} endpoint
  if (event.httpMethod === 'POST' && (event.resource === '/files/{id}' || event.path?.match(/\/files\/[^/]+$/))) {
    try {
      // Get the parent folder id from path parameters
      const parentFolderId = event.pathParameters?.id;

      if (!parentFolderId) {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Missing id parameter in path' }),
        };
      }

      // Parse the request body
      if (!event.body) {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Missing request body' }),
        };
      }

      const body = JSON.parse(event.body);
      const { name, type } = body;

      if (!name || !type) {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Missing required fields: name and type' }),
        };
      }

      // Create the new File entry
      const now = new Date().toISOString();
      const { data: newFile } = await client.models.File.create({
        name,
        type,
        size: 0,
        fileFolderId: parentFolderId,
        createdDate: now,
        lastUpdatedDate: now,
      });

      return {
        statusCode: 201,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newFile),
      };
    } catch (error) {
      console.error('Error creating file:', error);
      return {
        statusCode: 500,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          error: 'Internal server error',
          message: error instanceof Error ? error.message : 'Unknown error',
        }),
      };
    }
  }

  // Handle /info endpoint
  if (event.path === '/dev/info' || event.resource === '/info') {
    return {
      statusCode: 200,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': '*',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        free: 982929299222,
        total: 1995218165760,
        used: 1067712249856,
      }),
    };
  }

  // Handle /direct endpoint
  if (event.path === '/dev/direct' || event.resource === '/direct') {
    try {
      // Get the id from query parameters
      const id = event.queryStringParameters?.id;

      if (!id) {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Missing id parameter' }),
        };
      }

      // Query the File by id
      const { data: file } = await client.models.File.get({ id });

      if (!file) {
        return {
          statusCode: 404,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'text/plain',
          },
          body: 'empty',
        };
      }

      // Return the text value or empty string
      return {
        statusCode: 200,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'text/plain',
        },
        body: file.text || 'empty',
      };
    } catch (error) {
      console.error('Error fetching file:', error);
      return {
        statusCode: 500,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'text/plain',
        },
        body: '',
      };
    }
  }

  // Handle DELETE /files endpoint
  if (event.httpMethod === 'DELETE' && (event.resource === '/files' || event.path === '/dev/files')) {
    try {
      // Parse the request body
      if (!event.body) {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Missing request body' }),
        };
      }

      const body = JSON.parse(event.body);
      const { ids } = body;

      if (!ids || !Array.isArray(ids) || ids.length === 0) {
        return {
          statusCode: 400,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Headers': '*',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ error: 'Missing or invalid ids array' }),
        };
      }

      // Delete each file by id
      const deleteResults = await Promise.allSettled(
        ids.map((id) => client.models.File.delete({ id }))
      );

      // Count successful deletions
      const successCount = deleteResults.filter(
        (result) => result.status === 'fulfilled' && result.value.data !== null
      ).length;

      return {
        statusCode: 200,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          message: `Successfully deleted ${successCount} of ${ids.length} files`,
          deletedCount: successCount,
          totalRequested: ids.length,
        }),
      };
    } catch (error) {
      console.error('Error deleting files:', error);
      return {
        statusCode: 500,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          error: 'Internal server error',
          message: error instanceof Error ? error.message : 'Unknown error',
        }),
      };
    }
  }

  // Handle GET /files endpoint
  try {
    // Query for the root FileFolder
    const { data: fileFolders } = await client.models.FileFolder.list();

    let fileFolder = fileFolders?.[0];

    // If no FileFolder exists, create one with sample data
    if (!fileFolder) {
      console.log('No FileFolder found, creating one with sample data');

      const now = new Date().toISOString();

      // Create the root FileFolder
      const { data: newFileFolder } = await client.models.FileFolder.create({
        name: 'root',
        createdDate: now,
        lastUpdatedDate: now,
      });

      if (newFileFolder) {
        fileFolder = newFileFolder;
      } else {
        throw new Error('Failed to create or retrieve FileFolder');
      }
    }

    // Get all files for this FileFolder
    let { data: files } = await client.models.File.list({
      filter: {
        fileFolderId: {
          eq: fileFolder.id,
        },
        // Only get root-level files (no parent)
        parentFileId: {
          attributeExists: false,
        },
      },
    });

    // If no files exist, create sample data
    if (!files || files.length === 0) {
      console.log('No files found, creating sample data');

      const now = new Date().toISOString();

      // Create sample text file
      await client.models.File.create({
        name: 'sample.txt',
        type: 'file',
        size: 8,
        text: 'This is a sample text file',
        createdDate: now,
        lastUpdatedDate: now,
        fileFolderId: fileFolder.id,
      });

      // Create sample folder
      await client.models.File.create({
        name: 'sample',
        type: 'folder',
        size: 0,
        createdDate: now,
        lastUpdatedDate: now,
        fileFolderId: fileFolder.id,
      });

      // Re-query to get the newly created files
      const { data: newFiles } = await client.models.File.list({
        filter: {
          fileFolderId: {
            eq: fileFolder.id,
          },
          parentFileId: {
            attributeExists: false,
          },
        },
      });
      files = newFiles
    }

    // Format the files to match the expected response structure
    const formattedFiles = files.map((file) => {
      const result: any = {
        id: `/${file.name}`,
        //id: `${file.id}`,
        date: new Date(file.lastUpdatedDate),
        type: file.type,
      };

      if (file.type === 'file' && file.size !== null && file.size !== undefined) {
        result.size = file.size;
      }

      return result;
    });

    return {
      statusCode: 200,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': '*',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(formattedFiles),
    };
  } catch (error) {
    console.error('Error:', error);

    return {
      statusCode: 500,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': '*',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        error: 'Internal server error',
        message: error instanceof Error ? error.message : 'Unknown error',
      }),
    };
  }
};
