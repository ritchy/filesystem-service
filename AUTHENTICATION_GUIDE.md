# Authentication Implementation Guide

## Overview

This guide documents the authentication implementation for the filesystem.io application using AWS Amplify Authenticator and user-specific storage paths.

## Features Implemented

### 1. Login Landing Page
- **Component**: AWS Amplify UI Authenticator
- **Location**: `client/src/App.tsx`
- **Functionality**: 
  - Displays a login/signup form before users can access the file system
  - Handles email-based authentication
  - Provides built-in UI components for login, signup, and password reset

### 2. Persistent Login with Token Refresh
- **Configuration**: Amplify automatically handles token refresh
- **Implementation**: 
  - Amplify SDK automatically refreshes access tokens before they expire
  - Refresh tokens are stored securely in browser storage
  - Users remain logged in across browser sessions until they explicitly sign out
- **Token Lifecycle**:
  - Access tokens expire after 1 hour by default
  - Refresh tokens are valid for 30 days by default
  - Tokens are automatically refreshed when needed

### 3. User-Specific Storage Paths
- **Structure**: `files/{userId}/*`
- **Implementation**:
  - Each authenticated user gets a unique folder in S3
  - User ID is obtained from AWS Cognito Identity Pool (identityId) or user pool (userSub)
  - Files uploaded by a user are stored in their specific folder: `files/{userId}/{timestamp}_{filename}`

### 4. Per-User Data Isolation
- **Owner-Based Authorization**: FileFolder and File models use owner-based authorization
- **Automatic Filtering**: AWS Amplify automatically filters data to show only records owned by the current user
- **Data Segregation**:
  - Each user gets their own FileFolder (root folder)
  - Each user's files are isolated from other users
  - Users cannot see or access other users' files or folders

## Files Modified

### 1. `amplify/auth/resource.ts`
- Configured email-based authentication
- Set up required user attributes

### 2. `amplify/storage/resource.ts`
- Updated access rules to support user-specific paths: `files/{entity_id}/*`
- Configured permissions for authenticated and guest users
- **Note**: Legacy `files/*` path was removed to avoid path conflict errors (AWS Amplify doesn't allow overlapping storage paths)

### 3. `client/src/api.ts`
- Added `getCurrentUserId()` function to retrieve authenticated user's ID
- Updated `uploadFile()` function to use user-specific storage paths
- Configured Amplify with token refresh support
- Imported `fetchAuthSession` from aws-amplify/auth

### 4. `client/src/App.tsx`
- Wrapped application with `<Authenticator>` component
- Split into two components:
  - `FileSystemApp`: Main application logic
  - `App`: Wrapper component with Authenticator
- Added user information display in header (email/username)
- Added "Sign Out" button
- Added userId state tracking

### 5. `amplify/data/resource.ts`
- Updated FileFolder model with owner-based authorization (`allow.owner()`)
- Updated File model with owner-based authorization (`allow.owner()`)
- Changed default authorization mode from 'identityPool' to 'userPool'
- Each user now gets their own isolated FileFolder and File records
- AWS Amplify automatically filters queries to return only the current user's data

### 6. `client/src/api.ts`
- Updated `fetchRootFolder()` to automatically create root folder if missing
- Root folder is created per-user with proper timestamps

### 7. `client/package.json`
- Added `@aws-amplify/ui-react` dependency for Authenticator component

## Usage

### For Users

1. **First Time Access**:
   - Navigate to the application
   - You'll see a login screen
   - Click "Create Account" to sign up with email and password
   - Verify your email address
   - Log in with your credentials

2. **Subsequent Visits**:
   - If you remain logged in, you'll go directly to the file system
   - If session expired, you'll see the login screen
   - Log in to access your files

3. **File Uploads**:
   - All files you upload are stored in your personal folder
   - Other users cannot access your files
   - Your files persist across sessions

4. **Signing Out**:
   - Click the "Sign Out" button in the header to log out
   - You'll be redirected to the login screen

### For Developers

1. **Testing Authentication**:
   ```bash
   cd client
   npm start
   ```
   - Create a test account
   - Upload a file
   - Check the console for user ID logs
   - Verify files are stored in user-specific S3 paths

2. **Customizing the Authenticator**:
   - Modify the `<Authenticator>` component in `App.tsx`
   - Add custom components or styling
   - See: https://ui.docs.amplify.aws/react/connected-components/authenticator

3. **Accessing User Information**:
   ```typescript
   import { fetchAuthSession } from 'aws-amplify/auth';
   
   const session = await fetchAuthSession();
   const userId = session.identityId || session.userSub;
   const userEmail = session.tokens?.idToken?.payload.email;
   ```

## Storage Structure

### S3 Bucket Organization
```
filesystemStorage/
├── files/
│   ├── {userId1}/
│   │   ├── {timestamp1}_{filename1}
│   │   └── {timestamp2}_{filename2}
│   ├── {userId2}/
│   │   └── {timestamp3}_{filename3}
│   └── (legacy files without user folders)
```

### Database Structure
- File metadata is stored in DynamoDB via the `File` model
- Each file record includes:
  - `fileReference`: S3 path including user folder
  - `fileFolderId`: Root folder reference
  - User ownership is implicit through the storage path

## Security Considerations

1. **Authentication Required**:
   - Users must be authenticated to access the application
   - All API calls are made with authenticated user credentials

2. **Storage Access Control**:
   - S3 access rules enforce user-specific folder access
   - Pattern: `files/{entity_id}/*` ensures users can only access their own files
   - AWS Amplify automatically validates the entity_id matches the authenticated user

3. **Token Security**:
   - Tokens are stored securely in browser storage
   - Access tokens expire after 1 hour
   - Refresh tokens are automatically rotated

4. **Future Enhancements**:
   - Consider implementing file sharing between users
   - Add user groups for shared folders
   - Implement file encryption at rest

## Troubleshooting

### Issue: User ID not found
- **Cause**: Authentication session not established
- **Solution**: Ensure user is logged in and Amplify is configured correctly

### Issue: Files not uploading to user folder
- **Cause**: `getCurrentUserId()` returns null
- **Solution**: Check authentication state and wait for session to be established

### Issue: Access denied on S3
- **Cause**: Storage access rules not deployed
- **Solution**: Redeploy Amplify backend with updated storage configuration

### Issue: Login screen not appearing
- **Cause**: Authenticator component not rendering
- **Solution**: Check that `@aws-amplify/ui-react` is installed and imported correctly

## Deployment

1. **Update Backend**:
   ```bash
   # From project root
   npx ampx sandbox
   ```

2. **Deploy to Production**:
   ```bash
   # Deploy Amplify backend changes
   npx ampx pipeline-deploy --branch main
   ```

3. **Verify Changes**:
   - Test authentication flow
   - Upload a file and verify S3 path
   - Check CloudWatch logs for any errors

## Additional Resources

- [AWS Amplify Authentication Docs](https://docs.amplify.aws/react/build-a-backend/auth/)
- [Amplify UI Authenticator](https://ui.docs.amplify.aws/react/connected-components/authenticator)
- [AWS Amplify Storage](https://docs.amplify.aws/react/build-a-backend/storage/)
- [Token Refresh Configuration](https://docs.amplify.aws/react/build-a-backend/auth/concepts/tokens/)
