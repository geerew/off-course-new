import { persisted } from 'svelte-persisted-store';

export const preferences = persisted('oc-video-preferences', {
	autoplay: false,
	autoplayNext: true
});
