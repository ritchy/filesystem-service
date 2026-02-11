import React, { useState, useEffect, useRef } from 'react';
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
} from './api';
import { FileItem, FileInfo, ContextMenuPosition, ModalData } from './types';

interface TreeNode {
  file: FileItem;
  children: TreeNode[];
  isExpanded: boolean;
  isLoaded: boolean;
}

function App() {
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

  const modalNameRef = useRef<HTMLInputElement>(null);
  const modalTextRef = useRef<HTMLTextAreaElement>(null);

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

  // Modal submit
  const handleModalSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!modal) return;

    try {
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
      }

      await loadRootData();
      setModal(null);
    } catch (err) {
      alert('Operation failed');
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
      <div className="header">
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

            <div className="column">
              <div className="column-header">
                {selectedTreeItem ? selectedTreeItem.name : 'Select a folder'}
              </div>
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
        </div>
      )}

      {modal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <div className="modal-title">
              {modal.type === 'rename' && `Rename "${modal.file?.name}"`}
              {modal.type === 'createFolder' && 'Create New Folder'}
              {modal.type === 'createFile' && 'Create New File'}
            </div>
            <form className="modal-form" onSubmit={handleModalSubmit}>
              <div className="form-group">
                <label className="form-label">Name</label>
                <input
                  ref={modalNameRef}
                  type="text"
                  className="form-input"
                  defaultValue={modal.type === 'rename' ? modal.file?.name : ''}
                  required
                  autoFocus
                />
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
              <div className="modal-buttons">
                <button type="button" className="btn btn-secondary" onClick={() => setModal(null)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  {modal.type === 'rename' ? 'Rename' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;
