<script lang="ts">
	import '@fontsource/source-sans-pro/400.css';
	import '@fontsource/source-sans-pro/700.css';
	import '../app.css';

	import { Toaster } from 'svelte-sonner';

	let { children } = $props();
</script>

<Toaster theme="dark" richColors />

<main>
	<div class="h-screen w-screen px-5">
		{@render children()}
	</div>
</main>
