import { defineStorage } from '@aws-amplify/backend';

export const storage = defineStorage({
  name: 'filesystemStorage',
  access: (allow) => ({
    // User-specific storage paths - each user gets their own folder
    'files/{entity_id}/*': [
      allow.authenticated.to(['read', 'write', 'delete']),
      allow.guest.to(['read', 'write', 'delete'])
    ],
    // Legacy path for backward compatibility
    'files/*': [
      allow.guest.to(['read', 'write', 'delete'])
    ]
  })
});
