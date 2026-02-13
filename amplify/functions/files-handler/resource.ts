import { defineFunction } from '@aws-amplify/backend';

export const filesHandler = defineFunction({
  name: 'files-handler',
  environment: {
    // Environment variables will be populated by backend.ts
    AMPLIFY_DATA_DEFAULT_NAME: 'filesystemData',
  },
});
