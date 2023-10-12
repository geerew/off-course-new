<script lang="ts">
	import Toast from '$components/toaster/Toast.svelte';
	import { addToast } from '$lib/stores/addToast';
	import type { ToastData } from '$lib/types/general';
	import { createToaster } from '@melt-ui/svelte';
	import { flip } from 'svelte/animate';

	// ----------------------
	// Variables
	// ----------------------
	const {
		elements,
		helpers: { addToast: addToastHelper },
		states: { toasts },
		actions: { portal }
	} = createToaster<ToastData>({ closeDelay: 3000 });

	// Externalize the addToast function
	addToast.set(addToastHelper);
</script>

<div
	class="fixed left-1/2 top-0 z-[100] m-4 flex -translate-x-1/2 transform flex-col items-end gap-2"
	use:portal
>
	{#each $toasts as toast (toast.id)}
		<div animate:flip={{ duration: 500 }}>
			<Toast {elements} {toast} />
		</div>
	{/each}
</div>
