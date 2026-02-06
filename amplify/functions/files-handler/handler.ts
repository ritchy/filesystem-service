import type { APIGatewayProxyHandler } from 'aws-lambda';
import { Amplify } from 'aws-amplify';
import { generateClient } from 'aws-amplify/data';
import type { Schema } from '../../data/resource';
import outputs from "../../../amplify_outputs.json";

// Configure Amplify
Amplify.configure(outputs);

const client = generateClient<Schema>();

export const handler: APIGatewayProxyHandler = async (event) => {
  console.log('event', event);

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
        // Create sample text file
        await client.models.File.create({
          name: 'sample.txt',
          type: 'file',
          size: 0,
          text: '',
          createdDate: now,
          lastUpdatedDate: now,
          fileFolderId: newFileFolder.id,
        });

        // Create sample folder
        await client.models.File.create({
          name: 'sample',
          type: 'folder',
          size: 0,
          createdDate: now,
          lastUpdatedDate: now,
          fileFolderId: newFileFolder.id,
        });

        fileFolder = newFileFolder;
      }
    }

    if (!fileFolder) {
      throw new Error('Failed to create or retrieve FileFolder');
    }

    // Get all files for this FileFolder
    const { data: files } = await client.models.File.list({
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

      // Format and return the response
      const formattedFiles = (newFiles || []).map((file) => {
        const result: any = {
          id: `/${file.name}`,
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
    }

    // Format the files to match the expected response structure
    const formattedFiles = files.map((file) => {
      const result: any = {
        id: `/${file.name}`,
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
