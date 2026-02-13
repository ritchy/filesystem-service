import { defineFunction } from '@aws-amplify/backend';

export const filesHandler = defineFunction({
  name: 'files-handler',
  resourceGroupName: 'data',
});
