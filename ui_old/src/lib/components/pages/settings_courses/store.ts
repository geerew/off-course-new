import type { SortKey } from 'svelte-headless-table/plugins';
import { persisted } from 'svelte-persisted-store';

export const preferences = persisted('oc-settings-courses-preferences', {
	sortBy: <SortKey>{ id: 'createdAt', order: 'desc' },
	hiddenColumns: ['updatedAt']
});
