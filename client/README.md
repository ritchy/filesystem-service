# Filesystem Explorer - React UI

A React-based file explorer interface that integrates with the AWS Amplify backend for managing a hierarchical file system.

## Features

### Three-Column Layout
1. **Left Column (File Tree)**: Recursive tree view starting from the root FileFolder
   - Expandable/collapsible folder nodes
   - Visual indicators for files and folders
   - Click to select and view children

2. **Middle Column (List View)**: Lists children of the selected item from the left column
   - Shows files and folders in the selected directory
   - Click to select and view info
   - Right-click for context menu operations

3. **Right Column (Info Panel)**: Displays detailed information about selected items
   - File/folder name
   - Type
   - Count of descendant files
   - Total size in bytes
   - Toggle visibility with button in header

### Search Functionality
- Search bar in the header
- Real-time search (debounced)
- Shows results overlay when 2+ characters are typed
- Results display all matching files and folders
- Click result to navigate to that item
- Clear button (X) to exit search and restore columns

### Context Menu Operations
Right-click on any file or folder in columns 1 or 2 to:
- **Rename**: Change the name of the item
- **Delete**: Remove the item (with confirmation)
- **Create Folder**: Add a new folder as a child
- **Create File**: Add a new file with name and text content

### Backend Integration
- Connects to AWS Amplify API Gateway endpoints
- Uses GraphQL for file operations
- REST API for file info retrieval
- Automatic icon loading from backend

## Getting Started

### Prerequisites
- Node.js and npm installed
- Backend service deployed and running
- `amplify_outputs.json` file in the client directory

### Installation

```bash
cd client
npm install
```

### Running the Application

```bash
npm start
```

Opens the app in development mode at [http://localhost:3000](http://localhost:3000).

### Building for Production

```bash
npm run build
```

Builds the app for production to the `build` folder.

## API Endpoints Used

- `GET /files` - List root level files
- `POST /files/{id}` - Create new file/folder
- `PUT /files/{id}` - Rename file/folder
- `DELETE /files` - Delete files
- `GET /info/{id}` - Get file/folder info (count and size)
- `GET /icons/{size}/{name}` - Get file/folder icons
- GraphQL API for file operations

## Tech Stack

- React 18
- TypeScript
- AWS Amplify
- Axios for REST calls
- CSS for styling

## Project Structure

```
client/
├── src/
│   ├── App.tsx          # Main component with all UI logic
│   ├── App.css          # Styles
│   ├── api.ts           # API integration functions
│   ├── types.ts         # TypeScript interfaces
│   └── index.tsx        # Entry point
├── amplify/
│   └── data/
│       └── resource.ts  # Amplify data schema types
└── amplify_outputs.json # Amplify configuration
```

## Usage Tips

1. **Navigate**: Click folders in the left tree to expand and view contents
2. **View Details**: Select items in the middle column to see info on the right
3. **Search**: Type at least 2 characters to search across all files and folders
4. **Manage Files**: Right-click items to rename, delete, or create new items
5. **Toggle Info**: Use the "Hide Info" / "Show Info" button to maximize space

## Notes

- The tree starts 1 level deep and expands on demand
- Search performs a case-sensitive contains match on file names
- File info shows recursive counts for folders (only counting files, not subfolders)
- Icons are fetched from the backend icon service
