<script lang="ts">
	import { dev } from '$app/environment';

	// Fonts
	import '@fontsource-variable/inter';
	import '@fontsource/ubuntu-mono';
	// Css
	import '../styles/app.postcss';

	import { TailwindIndicator } from '$components';
	import { Header } from '$components/header';
	import { Toaster } from '$components/toaster';
</script>

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
