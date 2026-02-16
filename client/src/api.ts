import axios from 'axios';
import { Amplify } from 'aws-amplify';
import { generateClient } from 'aws-amplify/data';
import { uploadData } from 'aws-amplify/storage';
import { fetchAuthSession } from 'aws-amplify/auth';
import type { Schema } from '../../amplify/data/resource';
import amplifyOutputs from './amplify_outputs.json';
import { FileItem, FileInfo } from './types';

// Configure Amplify with token refresh
Amplify.configure(amplifyOutputs, {
  ssr: false,
});

const client = generateClient<Schema>();
const API_BASE_URL = amplifyOutputs.custom.FilesApiUrl;

// Helper to get current authenticated user ID
export const getCurrentUserId = async (): Promise<string | null> => {
  try {
    const session = await fetchAuthSession();
    return session.identityId || session.userSub || null;
  } catch (error) {
    console.log('No authenticated user:', error);
    return null;
  }
};

// Helper to get file icon URL
export const getFileIconUrl = (type: 'file' | 'folder', size: 'small' | 'big' = 'small'): string => {
  const iconName = type === 'folder' ? 'folder.svg' : 'txt.svg';
  return `${API_BASE_URL}icons/${size}/${iconName}`;
};

// Helper to get direct URL for a file
export const getDirectUrl = (fileId: string): string => {
  return `${API_BASE_URL}direct?id=${fileId}`;
};

// Fetch root FileFolder and all files, creating the root folder if it doesn't exist
export const fetchRootFolder = async (): Promise<{ rootFolderId: string; rootFiles: FileItem[] }> => {
  try {
    const { data: fileFolders } = await client.models.FileFolder.list();

    let rootFolder;

    // Create root folder if it doesn't exist
    if (!fileFolders || fileFolders.length === 0) {
      console.log('No root folder found, creating one...');
      const now = new Date().toISOString();
      const { data: newRootFolder } = await client.models.FileFolder.create({
        name: 'root',
        createdDate: now,
        lastUpdatedDate: now,
      });

      if (!newRootFolder) {
        throw new Error('Failed to create root folder');
      }

      rootFolder = newRootFolder;
      console.log('Root folder created:', rootFolder);

      // Return with empty files array since folder was just created
      return {
        rootFolderId: rootFolder.id,
        rootFiles: []
      };
    }

    rootFolder = fileFolders[0];
    const { data: files } = await client.models.File.list({
      filter: {
        fileFolderId: { eq: rootFolder.id },
        parentFileId: { attributeExists: false }
      }
    });

    let rootFiles = files as unknown as FileItem[];

    // If root folder is empty, create a default 'files' folder
    if (!rootFiles || rootFiles.length === 0) {
      console.log('Root folder is empty, creating default "files" folder...');
      const now = new Date().toISOString();
      const { data: filesFolder } = await client.models.File.create({
        name: 'files',
        type: 'folder',
        size: 0,
        fileFolderId: rootFolder.id,
        createdDate: now,
        lastUpdatedDate: now,
      });

      if (filesFolder) {
        rootFiles = [filesFolder as unknown as FileItem];
        console.log('Default "files" folder created');
      }
    }

    return {
      rootFolderId: rootFolder.id,
      rootFiles: rootFiles
    };
  } catch (error) {
    console.error('Error fetching root folder:', error);
    throw error;
  }
};

// Fetch children of a file/folder
export const fetchChildren = async (parentId: string): Promise<FileItem[]> => {
  try {
    const { data: files } = await client.models.File.list({
      filter: {
        parentFileId: { eq: parentId }
      }
    });

    return files as unknown as FileItem[];
  } catch (error) {
    console.error('Error fetching children:', error);
    throw error;
  }
};

// Get file info (count and size)
export const fetchFileInfo = async (fileId: string): Promise<FileInfo> => {
  try {
    const response = await axios.get(`${API_BASE_URL}info/${fileId}`);
    return response.data;
  } catch (error) {
    console.error('Error fetching file info:', error);
    throw error;
  }
};

// Rename a file
export const renameFile = async (fileId: string, newName: string): Promise<void> => {
  try {
    await axios.put(`${API_BASE_URL}files/${fileId}`, {
      operation: 'rename',
      name: newName
    });
  } catch (error) {
    console.error('Error renaming file:', error);
    throw error;
  }
};

// Delete files
export const deleteFiles = async (fileIds: string[]): Promise<void> => {
  try {
    await axios.delete(`${API_BASE_URL}files`, {
      data: { ids: fileIds }
    });
  } catch (error) {
    console.error('Error deleting files:', error);
    throw error;
  }
};

// Create a folder
export const createFolder = async (parentId: string, name: string): Promise<FileItem> => {
  try {
    const now = new Date().toISOString();
    const { data: newFolder } = await client.models.File.create({
      name,
      type: 'folder',
      size: 0,
      parentFileId: parentId,
      fileFolderId: parentId, // Assuming parent is the root folder ID
      createdDate: now,
      lastUpdatedDate: now
    });

    return newFolder as unknown as FileItem;
  } catch (error) {
    console.error('Error creating folder:', error);
    throw error;
  }
};

// Create a file
export const createFile = async (
  parentId: string,
  name: string,
  text: string,
  rootFolderId: string
): Promise<FileItem> => {
  try {
    const now = new Date().toISOString();
    const { data: newFile } = await client.models.File.create({
      name,
      type: 'file',
      size: text.length,
      text,
      parentFileId: parentId,
      fileFolderId: rootFolderId,
      createdDate: now,
      lastUpdatedDate: now
    });

    return newFile as unknown as FileItem;
  } catch (error) {
    console.error('Error creating file:', error);
    throw error;
  }
};

// Search files by name
export const searchFiles = async (query: string): Promise<FileItem[]> => {
  try {
    const { data: files } = await client.models.File.list({
      filter: {
        name: { contains: query }
      }
    });

    return files as unknown as FileItem[];
  } catch (error) {
    console.error('Error searching files:', error);
    throw error;
  }
};

// Upload a file to S3 and create File entry
export const uploadFile = async (
  file: File,
  fileName: string,
  parentFileId: string | null,
  rootFolderId: string,
  onProgress?: (progress: number) => void
): Promise<FileItem> => {
  console.log('uploading file:', fileName, 'to parentFileId:', parentFileId, 'with rootFolderId:', rootFolderId);
  try {
    // Get current user ID for user-specific storage path
    const userId = await getCurrentUserId();

    // Upload file to S3 in user-specific folder
    const s3Path = userId
      ? `files/${userId}/${Date.now()}_${fileName}`
      : `files/${Date.now()}_${fileName}`;

    const result = await uploadData({
      path: s3Path,
      data: file,
      options: {
        contentType: file.type,
        onProgress: ({ transferredBytes, totalBytes }) => {
          if (totalBytes && onProgress) {
            const percentComplete = Math.round((transferredBytes / totalBytes) * 100);
            onProgress(percentComplete);
          }
        }
      }
    }).result;

    console.log('File uploaded to S3 at path:', s3Path, 'with result:', result);
    // Create File entry in database
    const now = new Date().toISOString();
    const { data: newFile } = await client.models.File.create({
      name: fileName,
      type: 'file',
      size: file.size,
      fileReference: s3Path,
      parentFileId: parentFileId || undefined,
      fileFolderId: rootFolderId,
      createdDate: now,
      lastUpdatedDate: now
    });

    return newFile as unknown as FileItem;
  } catch (error) {
    console.error('Error uploading file:', error);
    throw error;
  }
};
