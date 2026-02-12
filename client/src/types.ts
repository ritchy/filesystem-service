export interface FileItem {
  id: string;
  name: string;
  type: 'file' | 'folder';
  size?: number;
  text?: string;
  createdDate: string;
  lastUpdatedDate: string;
  fileFolderId?: string;
  parentFileId?: string;
  childFiles?: FileItem[];
}

export interface FileInfo {
  count: string;
  size: string;
}

export interface ContextMenuPosition {
  x: number;
  y: number;
}

export interface ModalData {
  type: 'rename' | 'createFolder' | 'createFile' | 'upload';
  file?: FileItem;
  parentId?: string;
  uploadFile?: File;
}
