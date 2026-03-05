Updated `fetchRootFolder` in `client/src/api.ts` to work with the new `Member` model. The function now:

1. Gets the current authenticated user's ID via `getCurrentUserId()`
2. Queries `client.models.Member.list()` filtered by `userId` to find the associated `Member` record
3. Calls `member.fileFolder()` to traverse the `hasOne` relationship and retrieve the member's `FileFolder`
4. Uses that `FileFolder` as the `rootFolder` going forward (fetching its files, creating the default "files" folder if empty, etc.)

The old approach of blindly listing all `FileFolder` records and picking the first one is replaced with a user-scoped lookup through the `Member` model.




