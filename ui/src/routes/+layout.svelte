<script lang="ts">
	import { dev } from '$app/environment';

	// Theme
	import { ModeWatcher } from 'mode-watcher';

	// Fonts
	import '@fontsource-variable/inter';
	import '@fontsource/ubuntu-mono';
	// Css
	import '../app.pcss';

	import { TailwindIndicator } from '$components';
	import { Header } from '$components/header';
	import { Toaster } from '$components/toaster';
</script>

<ModeWatcher />

<Toaster />

<div class="flex min-h-screen flex-col">
	<Header />
	<div class="flex flex-1">
		<slot />
	</div>
	{#if dev}
		<TailwindIndicator />
	{/if}
</div>
