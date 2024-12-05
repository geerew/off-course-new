import type { LogLevel } from '$lib/types/models';
import { persisted } from 'svelte-persisted-store';

export const preferences = persisted('oc-settings-logs-preferences', {
	messages: <string[]>[],
	levels: <LogLevel[]>[],
	types: <string[]>[]
});
