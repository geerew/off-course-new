<script lang="ts">
	import { currentTheme, setPreference } from '$lib/utils/theme';
	import { Icons } from './icons';
</script>

<div class="flex items-center px-1">
	<button
		id="theme-btn"
		class="group rounded-md p-1.5 text-sm font-semibold duration-200 hover:bg-accent-1"
		aria-label="auto"
		aria-live="polite"
		on:click={() => {
			const theme = currentTheme();
			setPreference(theme === 'light' ? 'dark' : 'light');
		}}
	>
		<Icons.moon
			class="moon h-5 w-5 text-foreground-muted group-hover:text-foreground duration-200"
		/>
		<Icons.sun class="sun h-5 w-5 text-foreground/80 group-hover:text-foreground duration-200" />
	</button>
</div>
