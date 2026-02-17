import React, { useState, useEffect, useRef } from 'react';
import { Authenticator } from '@aws-amplify/ui-react';
import '@aws-amplify/ui-react/styles.css';
import './App.css';
import {
  fetchRootFolder,
  fetchChildren,
  fetchFileInfo,
  renameFile,
  deleteFiles,
  createFolder,
  createFile,
  searchFiles,
  getFileIconUrl,
  getDirectUrl,
  uploadFile,
  getCurrentUserId,
} from './api';
import { FileItem, FileInfo, ContextMenuPosition, ModalData } from './types';

interface TreeNode {
  file: FileItem;
  children: TreeNode[];
  isExpanded: boolean;
  isLoaded: boolean;
}

function FileSystemApp({ signOut, user }: { signOut?: () => void; user?: any }) {
  const [rootFolderId, setRootFolderId] = useState<string>('');
  const [treeData, setTreeData] = useState<TreeNode[]>([]);
  const [selectedTreeItem, setSelectedTreeItem] = useState<FileItem | null>(null);
  const [middleColumnItems, setMiddleColumnItems] = useState<FileItem[]>([]);
  const [selectedMiddleItem, setSelectedMiddleItem] = useState<FileItem | null>(null);
  const [fileInfo, setFileInfo] = useState<FileInfo | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<FileItem[]>([]);
  const [showInfoColumn, setShowInfoColumn] = useState(true);
  const [contextMenu, setContextMenu] = useState<{ position: ContextMenuPosition; file: FileItem; source: 'tree' | 'middle' } | null>(null);
  const [modal, setModal] = useState<ModalData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [uploadingFile, setUploadingFile] = useState<FileItem | null>(null);
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [isUploading, setIsUploading] = useState(false);
  const [userId, setUserId] = useState<string | null>(null);
  const [showProfileMenu, setShowProfileMenu] = useState(false);
  const [showActionMenu, setShowActionMenu] = useState(false);

  const modalNameRef = useRef<HTMLInputElement>(null);
  const modalTextRef = useRef<HTMLTextAreaElement>(null);
  const profileMenuRef = useRef<HTMLDivElement>(null);
  const actionMenuRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Get user ID on mount
  useEffect(() => {
    const fetchUserId = async () => {
      const id = await getCurrentUserId();
      setUserId(id);
      console.log('User ID for storage:', id);
    };
    fetchUserId();
  }, []);

  // Load root folder and initial data
  useEffect(() => {
    loadRootData();
  }, []);

  const loadRootData = async () => {
    try {
      setLoading(true);
      setError(null);
      const { rootFolderId: id, rootFiles } = await fetchRootFolder();
      setRootFolderId(id);
      const nodes: TreeNode[] = rootFiles.map(file => ({
        file,
        children: [],
        isExpanded: false,
        isLoaded: false,
      }));
      setTreeData(nodes);
      setLoading(false);
    } catch (err) {
      setError('Failed to load filesystem data');
      setLoading(false);
    }
  };

  // Load children for a tree node
  const loadChildren = async (node: TreeNode): Promise<TreeNode[]> => {
    if (node.file.type === 'file') return [];
    const children = await fetchChildren(node.file.id);
    return children.map(file => ({
      file,
      children: [],
      isExpanded: false,
      isLoaded: false,
    }));
  };

  // Toggle tree node expansion
  const toggleTreeNode = async (path: number[]) => {
    const newTreeData = [...treeData];
    let current: TreeNode[] = newTreeData;
    let node: TreeNode | null = null;

    for (let i = 0; i < path.length; i++) {
      node = current[path[i]];
      if (i < path.length - 1) {
        current = node.children;
      }
    }

    if (!node) return;

    if (!node.isExpanded && !node.isLoaded) {
      const children = await loadChildren(node);
      node.children = children;
      node.isLoaded = true;
    }

    node.isExpanded = !node.isExpanded;
    setTreeData(newTreeData);
  };

  // Handle tree item selection
  const handleTreeSelect = async (file: FileItem) => {
    setSelectedTreeItem(file);
    setSelectedMiddleItem(null);
    setFileInfo(null);

    if (file.type === 'folder') {
      try {
        const children = await fetchChildren(file.id);
        setMiddleColumnItems(children);
      } catch (err) {
        console.error('Failed to load children:', err);
      }
    } else {
      setMiddleColumnItems([]);
    }
  };

  // Handle middle column item selection
  const handleMiddleSelect = async (file: FileItem) => {
    setSelectedMiddleItem(file);
    if (showInfoColumn) {
      try {
        const info = await fetchFileInfo(file.id);
        setFileInfo(info);
      } catch (err) {
        console.error('Failed to load file info:', err);
        setFileInfo(null);
      }
    }
  };

  // Handle double-click on parent directory (..)
  const handleParentNavigate = async () => {
    if (!selectedTreeItem || !selectedTreeItem.parentFileId) return;

    // Find the parent file
    const findParentFile = async (parentId: string): Promise<FileItem | null> => {
      // Search in tree data
      const searchInTree = (nodes: TreeNode[]): FileItem | null => {
        for (const node of nodes) {
          if (node.file.id === parentId) {
            return node.file;
          }
          if (node.children.length > 0) {
            const found = searchInTree(node.children);
            if (found) return found;
          }
        }
        return null;
      };

      let parentFile = searchInTree(treeData);

      // If not found in tree, fetch it
      if (!parentFile) {
        try {
          const children = await fetchChildren(parentId);
          // We need to get the parent file itself, so let's search root level
          const { rootFiles } = await fetchRootFolder();
          parentFile = rootFiles.find(f => f.id === parentId) || null;
        } catch (err) {
          console.error('Failed to find parent:', err);
        }
      }

      return parentFile;
    };

    const parentFile = await findParentFile(selectedTreeItem.parentFileId);
    if (parentFile) {
      await handleTreeSelect(parentFile);
      await expandTreeToFile(parentFile.id);
    }
  };

  // Handle double-click on middle column items
  const handleMiddleDoubleClick = async (file: FileItem) => {
    if (file.type === 'folder') {
      // Select the folder in the tree and show its children
      await handleTreeSelect(file);
      // Expand the tree to show this folder
      await expandTreeToFile(file.id);
    } else if (file.type === 'file') {
      // Open file content in new window
      const directUrl = getDirectUrl(file.id);
      window.open(directUrl, '_blank');
    }
  };

  // Handle double-click on tree items
  const handleTreeDoubleClick = async (file: FileItem, path: number[]) => {
    if (file.type === 'file') {
      // Open file content in new window
      const directUrl = getDirectUrl(file.id);
      window.open(directUrl, '_blank');
    } else if (file.type === 'folder') {
      // Toggle expansion for folders
      await toggleTreeNode(path);
    }
  };

  // Find and expand tree path to a specific file
  const expandTreeToFile = async (fileId: string) => {
    // First, we need to find the path to this file
    const path = await findFilePathInTree(fileId);
    if (path) {
      // Expand all parent nodes
      await expandPath(path);
    }
  };

  // Find the path to a file in the tree
  const findFilePathInTree = async (fileId: string, nodes: TreeNode[] = treeData, currentPath: number[] = []): Promise<number[] | null> => {
    for (let i = 0; i < nodes.length; i++) {
      const node = nodes[i];
      if (node.file.id === fileId) {
        return [...currentPath, i];
      }

      // Load children if not loaded and it's a folder
      if (node.file.type === 'folder' && !node.isLoaded) {
        const children = await loadChildren(node);
        node.children = children;
        node.isLoaded = true;
      }

      // Search in children
      if (node.children.length > 0) {
        const childPath = await findFilePathInTree(fileId, node.children, [...currentPath, i]);
        if (childPath) {
          return childPath;
        }
      }
    }
    return null;
  };

  // Expand all nodes along a path
  const expandPath = async (path: number[]) => {
    const newTreeData = [...treeData];
    let current: TreeNode[] = newTreeData;

    for (let i = 0; i < path.length - 1; i++) {
      const node = current[path[i]];
      if (!node.isLoaded) {
        const children = await loadChildren(node);
        node.children = children;
        node.isLoaded = true;
      }
      node.isExpanded = true;
      current = node.children;
    }

    setTreeData(newTreeData);
  };

  // Handle search
  useEffect(() => {
    const performSearch = async () => {
      if (searchQuery.length >= 2) {
        try {
          const results = await searchFiles(searchQuery);
          setSearchResults(results);
        } catch (err) {
          console.error('Search failed:', err);
        }
      } else {
        setSearchResults([]);
      }
    };

    const timeoutId = setTimeout(performSearch, 300);
    return () => clearTimeout(timeoutId);
  }, [searchQuery]);

  // Handle context menu
  const handleContextMenu = (e: React.MouseEvent, file: FileItem, source: 'tree' | 'middle') => {
    e.preventDefault();
    setContextMenu({
      position: { x: e.clientX, y: e.clientY },
      file,
      source,
    });
  };

  // Close context menu
  useEffect(() => {
    const handleClick = () => setContextMenu(null);
    document.addEventListener('click', handleClick);
    return () => document.removeEventListener('click', handleClick);
  }, []);

  // Close profile menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (profileMenuRef.current && !profileMenuRef.current.contains(event.target as Node)) {
        setShowProfileMenu(false);
      }
    };

    if (showProfileMenu) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [showProfileMenu]);

  // Close action menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (actionMenuRef.current && !actionMenuRef.current.contains(event.target as Node)) {
        setShowActionMenu(false);
      }
    };

    if (showActionMenu) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [showActionMenu]);

  // Action menu handlers (mirror context menu actions)
  const handleActionRename = () => {
    const targetFile = selectedMiddleItem || selectedTreeItem;
    if (!targetFile) return;
    setModal({
      type: 'rename',
      file: targetFile,
    });
    setShowActionMenu(false);
  };

  const handleActionDelete = async () => {
    const targetFile = selectedMiddleItem || selectedTreeItem;
    if (!targetFile) return;
    if (window.confirm(`Are you sure you want to delete "${targetFile.name}"?`)) {
      try {
        await deleteFiles([targetFile.id]);
        await loadRootData();
        if (selectedTreeItem?.id === targetFile.id) {
          setSelectedTreeItem(null);
          setMiddleColumnItems([]);
        }
        if (selectedMiddleItem?.id === targetFile.id) {
          setSelectedMiddleItem(null);
          setFileInfo(null);
        }
      } catch (err) {
        alert('Failed to delete item');
      }
    }
    setShowActionMenu(false);
  };

  const handleActionCreateFolder = () => {
    const targetFile = selectedMiddleItem || selectedTreeItem;
    const parentId = targetFile?.type === 'folder'
      ? targetFile.id
      : targetFile?.parentFileId || rootFolderId;
    setModal({
      type: 'createFolder',
      parentId,
    });
    setShowActionMenu(false);
  };

  const handleActionCreateFile = () => {
    const targetFile = selectedMiddleItem || selectedTreeItem;
    const parentId = targetFile?.type === 'folder'
      ? targetFile.id
      : targetFile?.parentFileId || rootFolderId;
    setModal({
      type: 'createFile',
      parentId,
    });
    setShowActionMenu(false);
  };

  const handleActionUploadFile = () => {
    const targetFile = selectedMiddleItem || selectedTreeItem;
    const parentId = targetFile?.type === 'folder'
      ? targetFile.id
      : targetFile?.parentFileId || rootFolderId;
    setModal({
      type: 'upload',
      parentId,
    });
    setShowActionMenu(false);
  };

  // Context menu actions
  const handleRename = () => {
    if (!contextMenu) return;
    setModal({
      type: 'rename',
      file: contextMenu.file,
    });
    setContextMenu(null);
  };

  const handleDelete = async () => {
    if (!contextMenu) return;
    if (window.confirm(`Are you sure you want to delete "${contextMenu.file.name}"?`)) {
      try {
        await deleteFiles([contextMenu.file.id]);
        await loadRootData();
        if (selectedTreeItem?.id === contextMenu.file.id) {
          setSelectedTreeItem(null);
          setMiddleColumnItems([]);
        }
        if (selectedMiddleItem?.id === contextMenu.file.id) {
          setSelectedMiddleItem(null);
          setFileInfo(null);
        }
      } catch (err) {
        alert('Failed to delete item');
      }
    }
    setContextMenu(null);
  };

  const handleCreateFolder = () => {
    if (!contextMenu) return;
    setModal({
      type: 'createFolder',
      parentId: contextMenu.file.type === 'folder' ? contextMenu.file.id : contextMenu.file.parentFileId || rootFolderId,
    });
    setContextMenu(null);
  };

  const handleCreateFile = () => {
    if (!contextMenu) return;
    setModal({
      type: 'createFile',
      parentId: contextMenu.file.type === 'folder' ? contextMenu.file.id : contextMenu.file.parentFileId || rootFolderId,
    });
    setContextMenu(null);
  };

  const handleUploadFile = () => {
    if (!contextMenu) return;
    setModal({
      type: 'upload',
      parentId: contextMenu.file.type === 'folder' ? contextMenu.file.id : contextMenu.file.parentFileId || rootFolderId,
    });
    setContextMenu(null);
  };

  // Handle file drop
  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);

    if (!e.dataTransfer.files || e.dataTransfer.files.length === 0) return;

    const droppedFile = e.dataTransfer.files[0];

    // Open upload modal with pre-populated filename
    setModal({
      type: 'upload',
      uploadFile: droppedFile,
      parentId: selectedTreeItem?.type === 'folder'
        ? selectedTreeItem.id
        : selectedTreeItem?.parentFileId || rootFolderId,
    });
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  // Handle browse button click
  const handleBrowseClick = () => {
    fileInputRef.current?.click();
  };

  // Handle file selection from browse dialog
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const selectedFile = e.target.files[0];
      setModal({
        ...modal!,
        type: 'upload',
        uploadFile: selectedFile,
      });
      if (modalNameRef.current) {
        modalNameRef.current.value = selectedFile.name;
      }
    }
  };

  // Modal submit
  const handleModalSubmit = async (e: React.FormEvent) => {
    console.log('handleModalSubmit');
    e.preventDefault();
    if (!modal) return;

    try {
      console.log('initiating upload');
      if (modal.type === 'rename' && modal.file) {
        const newName = modalNameRef.current?.value || '';
        await renameFile(modal.file.id, newName);
      } else if (modal.type === 'createFolder') {
        const name = modalNameRef.current?.value || '';
        await createFolder(modal.parentId || rootFolderId, name);
      } else if (modal.type === 'createFile') {
        const name = modalNameRef.current?.value || '';
        const text = modalTextRef.current?.value || '';
        await createFile(modal.parentId || rootFolderId, name, text, rootFolderId);
      } else if (modal.type === 'upload' && modal.uploadFile) {
        const fileName = modalNameRef.current?.value || '';
        setUploadingFile(null);
        setIsUploading(true);
        setUploadProgress(0);

        // Determine parent file ID based on selected item
        const parentFileId = selectedTreeItem?.type === 'folder'
          ? selectedTreeItem.id
          : selectedTreeItem?.parentFileId || null;

        // Upload file with progress tracking
        const newFile = await uploadFile(
          modal.uploadFile,
          fileName,
          parentFileId,
          rootFolderId,
          (progress) => {
            setUploadProgress(progress);
          }
        );
        setUploadingFile(newFile);
        setIsUploading(false);

        // Reload data to refresh tree
        await loadRootData();

        // Refresh middle column if we have a selected tree item
        if (selectedTreeItem) {
          const children = await fetchChildren(
            selectedTreeItem.type === 'folder'
              ? selectedTreeItem.id
              : selectedTreeItem.parentFileId || rootFolderId
          );
          setMiddleColumnItems(children);

          // Auto-select the new file in middle column
          setSelectedMiddleItem(newFile);

          // Load file info for the new file
          if (showInfoColumn) {
            const info = await fetchFileInfo(newFile.id);
            setFileInfo(info);
          }
        }
      }

      await loadRootData();
      setModal(null);
    } catch (err) {
      alert('Operation failed');
      console.error(err);
    }
  };

  // Render tree recursively
  const renderTreeNode = (node: TreeNode, path: number[], level: number = 0): React.ReactNode => {
    const hasChildren = node.file.type === 'folder';
    const isSelected = selectedTreeItem?.id === node.file.id;

    return (
      <div key={node.file.id} className="tree-item">
        <div
          className={`file-item ${isSelected ? 'selected' : ''}`}
          style={{ paddingLeft: `${level * 24 + 12}px` }}
          onClick={() => handleTreeSelect(node.file)}
          onDoubleClick={() => handleTreeDoubleClick(node.file, path)}
          onContextMenu={(e) => handleContextMenu(e, node.file, 'tree')}
        >
          {hasChildren && (
            <span
              className={`expand-icon ${node.isExpanded ? 'expanded' : ''}`}
              onClick={(e) => {
                e.stopPropagation();
                toggleTreeNode(path);
              }}
            >
              ▶
            </span>
          )}
          {!hasChildren && <span style={{ width: '16px', display: 'inline-block' }} />}
          <img src={getFileIconUrl(node.file.type)} alt={node.file.type} className="file-icon" />
          <span className="file-name">{node.file.name}</span>
        </div>
        {node.isExpanded && node.children.length > 0 && (
          <div className="tree-children">
            {node.children.map((child, idx) =>
              renderTreeNode(child, [...path, idx], level + 1)
            )}
          </div>
        )}
      </div>
    );
  };

  if (loading) {
    return <div className="App"><div className="loading">Loading...</div></div>;
  }

  if (error) {
    return <div className="App"><div className="error">{error}</div></div>;
  }

  return (
    <div className="App">
      {/* Brand Header */}
      <div className="brand-header">
        <div className="brand-logo">
          filesystem.io
        </div>
        <div className="profile-section" ref={profileMenuRef}>
          <button
            className="profile-button"
            onClick={() => setShowProfileMenu(!showProfileMenu)}
          >
            <svg className="profile-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
            <span className="profile-email">
              {user?.signInDetails?.loginId || user?.username || 'User'}
            </span>
          </button>
          {showProfileMenu && (
            <div className="profile-dropdown">
              <div className="profile-menu-item" onClick={() => {
                setShowProfileMenu(false);
                signOut && signOut();
              }}>
                Sign Out
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Search Header */}
      <div className="search-header">
        <div className="search-container">
          <input
            type="text"
            className="search-input"
            placeholder="Search files and folders..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          {searchQuery && (
            <button className="clear-button" onClick={() => setSearchQuery('')}>
              ✕
            </button>
          )}
        </div>
        <button className="toggle-button" onClick={() => setShowInfoColumn(!showInfoColumn)}>
          {showInfoColumn ? 'Hide Info' : 'Show Info'}
        </button>
      </div>

      <div className="main-content">
        {searchQuery.length >= 2 ? (
          <div className="search-overlay">
            <div className="search-results">
              <h3>Search Results ({searchResults.length})</h3>
              {searchResults.map((file) => (
                <div
                  key={file.id}
                  className="search-result-item"
                  onClick={() => {
                    setSearchQuery('');
                    handleTreeSelect(file);
                  }}
                >
                  <img src={getFileIconUrl(file.type)} alt={file.type} className="file-icon" />
                  <span className="file-name">{file.name}</span>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <>
            <div className="column">
              <div className="column-header">File Tree</div>
              {treeData.map((node, idx) => renderTreeNode(node, [idx]))}
            </div>

            <div
              className={`column drop-zone ${isDragging ? 'dragging' : ''}`}
              onDrop={handleDrop}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
            >
              <div className="column-header-with-actions">
                <span className="column-header-title">
                  {selectedTreeItem ? selectedTreeItem.name : 'Select a folder'}
                </span>
                {(selectedMiddleItem || selectedTreeItem) && (
                  <div className="action-menu-container" ref={actionMenuRef}>
                    <button
                      className="action-button"
                      onClick={() => setShowActionMenu(!showActionMenu)}
                    >
                      Action ▼
                    </button>
                    {showActionMenu && (
                      <div className="action-dropdown">
                        <div className="context-menu-item" onClick={handleActionRename}>
                          Rename
                        </div>
                        <div className="context-menu-item" onClick={handleActionDelete}>
                          Delete
                        </div>
                        <div className="context-menu-separator" />
                        <div className="context-menu-item" onClick={handleActionCreateFolder}>
                          Create Folder
                        </div>
                        <div className="context-menu-item" onClick={handleActionCreateFile}>
                          Create File
                        </div>
                        <div className="context-menu-item" onClick={handleActionUploadFile}>
                          Upload File
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
              {isDragging && (
                <div className="drop-overlay">
                  <div className="drop-message">
                    Drop file here to upload
                  </div>
                </div>
              )}
              {selectedTreeItem && selectedTreeItem.type === 'folder' && selectedTreeItem.parentFileId && (
                <div
                  key="parent-nav"
                  className="file-item parent-nav"
                  onDoubleClick={handleParentNavigate}
                >
                  <img src={getFileIconUrl('folder')} alt="folder" className="file-icon" />
                  <span className="file-name">..</span>
                </div>
              )}
              {middleColumnItems.map((file) => (
                <div
                  key={file.id}
                  className={`file-item ${selectedMiddleItem?.id === file.id ? 'selected' : ''}`}
                  onClick={() => handleMiddleSelect(file)}
                  onDoubleClick={() => handleMiddleDoubleClick(file)}
                  onContextMenu={(e) => handleContextMenu(e, file, 'middle')}
                >
                  <img src={getFileIconUrl(file.type)} alt={file.type} className="file-icon" />
                  <span className="file-name">{file.name}</span>
                </div>
              ))}
            </div>

            {showInfoColumn && (
              <div className="column">
                <div className="column-header">Info</div>
                {selectedMiddleItem && fileInfo && (
                  <div className="info-panel">
                    <div className="info-item">
                      <div className="info-label">Name</div>
                      <div className="info-value">{selectedMiddleItem.name}</div>
                    </div>
                    <div className="info-item">
                      <div className="info-label">Type</div>
                      <div className="info-value">{selectedMiddleItem.type}</div>
                    </div>
                    <div className="info-item">
                      <div className="info-label">Count</div>
                      <div className="info-value">{fileInfo.count} files</div>
                    </div>
                    <div className="info-item">
                      <div className="info-label">Size</div>
                      <div className="info-value">{parseInt(fileInfo.size).toLocaleString()} bytes</div>
                    </div>
                  </div>
                )}
              </div>
            )}
          </>
        )}
      </div>

      {contextMenu && (
        <div
          className="context-menu"
          style={{ left: contextMenu.position.x, top: contextMenu.position.y }}
        >
          <div className="context-menu-item" onClick={handleRename}>
            Rename
          </div>
          <div className="context-menu-item" onClick={handleDelete}>
            Delete
          </div>
          <div className="context-menu-separator" />
          <div className="context-menu-item" onClick={handleCreateFolder}>
            Create Folder
          </div>
          <div className="context-menu-item" onClick={handleCreateFile}>
            Create File
          </div>
          <div className="context-menu-item" onClick={handleUploadFile}>
            Upload File
          </div>
        </div>
      )}

      {modal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <div className="modal-title">
              {modal.type === 'rename' && `Rename "${modal.file?.name}"`}
              {modal.type === 'createFolder' && 'Create New Folder'}
              {modal.type === 'createFile' && 'Create New File'}
              {modal.type === 'upload' && 'Upload File'}
            </div>
            <form className="modal-form" onSubmit={handleModalSubmit}>
              <div className="form-group">
                <label className="form-label">Name</label>
                <div className="input-with-button">
                  <input
                    ref={modalNameRef}
                    type="text"
                    className="form-input"
                    defaultValue={
                      modal.type === 'rename' ? modal.file?.name :
                        modal.type === 'upload' ? modal.uploadFile?.name : ''
                    }
                    required
                    autoFocus
                  />
                  {modal.type === 'upload' && (
                    <>
                      <button
                        type="button"
                        className="browse-button"
                        onClick={handleBrowseClick}
                        title="Browse for file"
                      >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className="browse-icon">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                        </svg>
                      </button>
                      <input
                        ref={fileInputRef}
                        type="file"
                        style={{ display: 'none' }}
                        onChange={handleFileChange}
                      />
                    </>
                  )}
                </div>
              </div>
              {modal.type === 'createFile' && (
                <div className="form-group">
                  <label className="form-label">Content</label>
                  <textarea
                    ref={modalTextRef}
                    className="form-textarea"
                    placeholder="Enter file content..."
                  />
                </div>
              )}
              {modal.type === 'upload' && modal.uploadFile && (
                <>
                  <div className="form-group">
                    <label className="form-label">File Info</label>
                    <div className="info-value">
                      Size: {modal.uploadFile.size.toLocaleString()} bytes
                    </div>
                    <div className="info-value">
                      Type: {modal.uploadFile.type || 'unknown'}
                    </div>
                  </div>
                  {isUploading && (
                    <div className="form-group">
                      <label className="form-label">Upload Progress</label>
                      <div className="progress-bar-container">
                        <div
                          className="progress-bar-fill"
                          style={{ width: `${uploadProgress}%` }}
                        />
                      </div>
                      <div className="progress-text">{uploadProgress}%</div>
                    </div>
                  )}
                </>
              )}
              <div className="modal-buttons">
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => setModal(null)}
                  disabled={isUploading}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={isUploading}
                >
                  {modal.type === 'rename' ? 'Rename' : modal.type === 'upload' ? (isUploading ? 'Uploading...' : 'Upload') : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

// Main App component wrapped with Authenticator
function App() {
  const components = {
    Header() {
      return (
        <div style={{
          textAlign: 'center',
          marginTop: '60px',
          marginBottom: '40px'
        }}>
          <h1 style={{
            fontSize: '48px',
            fontWeight: '300',
            color: '#333',
            margin: '0',
            letterSpacing: '2px'
          }}>
            filesystem.io
          </h1>
        </div>
      );
    },
  };

  return (
    <Authenticator
      components={components}
      hideSignUp={true}
    >
      {({ signOut, user }) => (
        <FileSystemApp signOut={signOut} user={user} />
      )}
    </Authenticator>
  );
}

export default App;
