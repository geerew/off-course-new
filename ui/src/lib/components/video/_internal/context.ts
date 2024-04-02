import { getContext, setContext } from 'svelte';
import { writable, type Writable } from 'svelte/store';

type Props = {
	controlsOpen: boolean;
	settingsOpen: boolean;
	ended: boolean;
	buffering: boolean;
	draggingTime: boolean;
};
type Context = Writable<Props>;

export function setCtx() {
	const props = writable<Props>({
		controlsOpen: false,
		settingsOpen: false,
		ended: false,
		buffering: true,
		draggingTime: false
	});
	setContext('ctx', props);
}

export function getCtx() {
	return getContext<Context>('ctx');
}
