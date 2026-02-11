# React File Explorer UI - Complete Guide

This document provides a comprehensive overview of the React-based file explorer UI that has been created for the filesystem service.

## Overview

A fully-featured file explorer interface built with React and TypeScript that provides an intuitive way to navigate, search, and manage your hierarchical file system stored in AWS.

## Architecture

### Frontend (React)
- **Location**: `/client` directory
- **Framework**: React 18 with TypeScript
- **State Management**: React Hooks (useState, useEffect)
- **API Client**: AWS Amplify + Axios

### Backend Integration
- **GraphQL API**: File CRUD operations via AWS Amplify
- **REST API**: File info and operations via API Gateway
- **Authentication**: AWS Cognito (unauthenticated identities enabled)

## Features Implemented

### 1. Three-Column Layout

#### Left Column: File Tree
- Recursive tree structure starting from root FileFolder
- Expandable/collapsible folders with visual indicators (▶)
- Lazy loading of children (loaded on expand)
- Selected item highlighting
- Right-click context menu support

#### Middle Column: List View
- Displays children of selected tree item
- Shows files and folders as a flat list
- Click to select and view info
- Right-click context menu support

#### Right Column: Info Panel
- Shows detailed information for selected middle column item
- Displays: name, type, file count, and total size
- Uses `GET /info/{id}` endpoint for recursive calculation
- Toggle visibility with header button

### 2. Search Functionality
- Search input in header
- Triggers when 2+ characters entered
- Debounced search (300ms delay)
- Searches all files/folders by name
- Results displayed as overlay
- Click result to navigate to item
- Clear button (X) to exit search

### 3. Context Menu Operations
Right-click any item to access:
- **Rename**: Opens modal to rename file/folder
- **Delete**: Deletes item with confirmation dialog
- **Create Folder**: Creates new folder as child
- **Create File**: Creates new file with name and text content

### 4. Modal Dialogs
- Rename: Single input field with pre-filled current name
- Create Folder: Input for folder name
- Create File: Inputs for name and textarea for content
- Form validation (required fields)
- Cancel and submit buttons

## API Integration

### Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| GraphQL API | Query | List files, get file details |
| GraphQL API | Mutation | Create files/folders |
| `/files/{id}` | PUT | Rename files/folders |
| `/files` | DELETE | Delete files/folders |
| `/info/{id}` | GET | Get recursive file count and size |
| `/icons/{size}/{name}` | GET | Fetch file/folder icons |

### Data Flow

```
User Action → Component State Update → API Call → Backend → Response → State Update → UI Render
```

## Component Structure

```
App.tsx (Main Component)
├── Header
│   ├── Search Input
│   ├── Clear Button (X)
│   └── Toggle Info Button
├── Main Content
│   ├── Column 1: Tree View (recursive rendering)
│   ├── Column 2: List View
│   └── Column 3: Info Panel (conditionally rendered)
├── Search Overlay (conditionally rendered)
├── Context Menu (conditionally rendered)
└── Modal Dialog (conditionally rendered)
```

## State Management

### Main State Variables
- `rootFolderId`: ID of the root FileFolder
- `treeData`: Array of TreeNode objects for recursive tree
- `selectedTreeItem`: Currently selected item in tree
- `middleColumnItems`: Files to display in middle column
- `selectedMiddleItem`: Currently selected item in middle column
- `fileInfo`: Info data for selected middle item
- `searchQuery`: Current search text
- `searchResults`: Array of search results
- `showInfoColumn`: Boolean to show/hide info column
- `contextMenu`: Current context menu state and position
- `modal`: Current modal dialog state
- `loading`: Loading state for initial data fetch
- `error`: Error message if data fetch fails

### Tree Structure
```typescript
interface TreeNode {
  file: FileItem;
  children: TreeNode[];
  isExpanded: boolean;
  isLoaded: boolean;
}
```

## Styling

### CSS Architecture
- Single `App.css` file with all styles
- Class-based styling (no CSS-in-JS)
- Responsive layout using flexbox
- CSS variables for consistent theming
- Hover and active states for interactivity

### Key CSS Classes
- `.App`: Main container
- `.header`: Top bar with search and toggle
- `.main-content`: Three-column container
- `.column`: Individual column styles
- `.file-item`: File/folder row
- `.tree-item`: Tree node container
- `.context-menu`: Right-click menu
- `.modal-overlay`: Modal backdrop
- `.search-overlay`: Search results overlay

## Running the Application

### Development Mode
```bash
cd client
npm start
```
Opens at http://localhost:3000

### Production Build
```bash
cd client
npm run build
```
Creates optimized build in `/client/build`

### Prerequisites Checklist
- [ ] Backend service deployed
- [ ] `amplify_outputs.json` in client directory
- [ ] Node.js and npm installed
- [ ] Dependencies installed (`npm install`)

## Usage Examples

### Creating a New Folder
1. Right-click on a parent folder or file
2. Select "Create Folder" from context menu
3. Enter folder name in modal
4. Click "Create"

### Searching for Files
1. Type at least 2 characters in search box
2. View results in overlay
3. Click a result to navigate to it
4. Click X to clear search and return to tree view

### Viewing File Information
1. Select a folder in the left tree
2. Click a file or folder in the middle column
3. View count and size information in the right panel
4. Toggle info panel visibility with header button

## Error Handling

- Failed API calls show browser alerts
- Delete operations require confirmation
- Loading states displayed during data fetch
- Error messages shown if root folder fails to load

## Performance Considerations

- Lazy loading of tree nodes (children loaded on expand)
- Debounced search to reduce API calls
- Efficient state updates to minimize re-renders
- Icon caching by browser

## Future Enhancements (Optional)

- Drag and drop file moving
- File upload functionality
- Text file editing inline
- Keyboard shortcuts
- Bulk operations (multi-select)
- File preview
- Breadcrumb navigation
- Sort and filter options

## Troubleshooting

### Common Issues

**Issue**: App shows "No root folder found"
- **Solution**: Ensure backend is deployed and accessible
- Check `amplify_outputs.json` has correct API URLs

**Issue**: Icons not loading
- **Solution**: Verify icons endpoint is accessible
- Check CORS configuration on backend

**Issue**: Context menu doesn't appear
- **Solution**: Ensure right-click events aren't being blocked
- Check browser console for errors

**Issue**: TypeScript errors in api.ts
- **Solution**: Type conversions use `as unknown as Type` for Amplify compatibility
- Ensure amplify/data/resource.ts is copied to client

## Development Notes

- App uses React functional components with hooks
- All API calls are async/await
- Error boundaries could be added for better error handling
- Icons are fetched from backend (txt.svg for files, folder.svg for folders)
- The recursive tree rendering supports unlimited nesting depth

## File Checklist

Created files:
- ✅ `/client/src/App.tsx` - Main application component
- ✅ `/client/src/App.css` - All styles
- ✅ `/client/src/api.ts` - API integration
- ✅ `/client/src/types.ts` - TypeScript interfaces
- ✅ `/client/README.md` - Client documentation
- ✅ `/client/amplify_outputs.json` - Amplify configuration
- ✅ `/client/amplify/data/resource.ts` - Data schema types

Backend endpoints:
- ✅ `GET /info/{id}` - Implemented in files-handler
- ✅ `GET /icons/{size}/folder.svg` - Implemented in icons-handler
- ✅ All other CRUD endpoints - Existing

## Summary

The React file explorer UI is fully functional and ready to use. It provides a comprehensive interface for managing the hierarchical file system with all requested features including the three-column layout, search, context menus, and info display.

To start using it:
```bash
cd client
npm start
```

The application will open in your browser at http://localhost:3000 and connect to your deployed backend services.
