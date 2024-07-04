import type { SortKey } from 'svelte-headless-table/plugins';
import { persisted } from 'svelte-persisted-store';

export const preferences = persisted('oc-settings-tags-preferences', {
	sortBy: <SortKey>{ id: 'tag', order: 'asc' }
});
