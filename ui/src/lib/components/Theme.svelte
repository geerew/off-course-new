<script lang="ts">
	import { currentTheme, setPreference } from '$lib/utils/theme';
	import { Icons } from './icons';
</script>

<div class="flex items-center">
	<button
		id="theme-btn"
		class="hover:bg-accent-1 group rounded-md p-1.5 text-sm font-semibold duration-200"
		aria-label="auto"
		aria-live="polite"
		on:click={() => {
			const theme = currentTheme();
			setPreference(theme === 'light' ? 'dark' : 'light');
		}}
	>
		<Icons.moon
			class="moon text-foreground-muted group-hover:text-foreground h-5 w-5 duration-200"
		/>
		<Icons.sun class="sun text-foreground/80 group-hover:text-foreground h-5 w-5 duration-200" />
	</button>
</div>
