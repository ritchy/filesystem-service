import { defineFunction } from '@aws-amplify/backend';

export const iconsHandler = defineFunction({
  name: 'icons-handler',
  resourceGroupName: 'auth',
});
