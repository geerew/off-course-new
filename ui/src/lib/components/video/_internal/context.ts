import { getContext, setContext } from 'svelte';
import { writable, type Writable } from 'svelte/store';

type Props = {
	controls: boolean;
	settings: boolean;
};
type Context = Writable<Props>;

export function setCtx() {
	const props = writable<Props>({ controls: false, settings: false });
	setContext('ctx', props);
}

export function getCtx() {
	return getContext<Context>('ctx');
}
