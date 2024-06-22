import { persisted } from 'svelte-persisted-store';

export const preferences = persisted('oc-video-preferences', {
	autoplay: false,
	autoloadNext: true,
	playbackRate: 1,
	volume: 1,
	muted: false
});
