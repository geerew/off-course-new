import { getContext, setContext } from 'svelte';
import { writable, type Writable } from 'svelte/store';

type Props = {
	settingsOpen: boolean;
	autoplay: boolean;
	autoplayNext: boolean;
};
type Context = Writable<Props>;

export function setCtx() {
	const props = writable<Props>({
		settingsOpen: false,
		autoplay: false,
		autoplayNext: true
	});
	setContext('ctx', props);
}

export function getCtx() {
	return getContext<Context>('ctx');
}
