<script lang="ts">
	import { page } from '$app/stores';
	import { settings } from '$lib/config/navs';
	import { cn } from '$lib/utils';

	// ----------------------
	// Variables
	// ----------------------
	let backgroundStyle = '';

	// ----------------------
	// Actions
	// ----------------------

	// Nav action that floats a background behind the nav items
	function floatingBackground(node: HTMLElement) {
		let slide = false;

		function handleMouseOver(e: MouseEvent) {
			if (!e.target) return;

			const nodeX = node.getBoundingClientRect().x;
			const { width: targetWidth, x: targetX } = (e.target as HTMLElement).getBoundingClientRect();

			backgroundStyle = `width: ${targetWidth}px; `;
			backgroundStyle += `transform: translateX(${targetX - nodeX}px); `;

			if (slide) {
				backgroundStyle += `transition-duration: 200ms; `;
				backgroundStyle += `transition-property: width,opacity,transform; `;
			}

			slide = true;
		}

		function handleMouseLeave() {
			slide = false;
		}

		node.addEventListener('mouseover', handleMouseOver);
		node.addEventListener('mouseleave', handleMouseLeave);

		return {
			destroy() {
				node.removeEventListener('mouseover', handleMouseOver);
				node.removeEventListener('mouseleave', handleMouseLeave);
			}
		};
	}
</script>

<nav class="group relative flex select-none" use:floatingBackground>
	<div
		class="background bg-muted pointer-events-none absolute left-0 top-0 h-8 rounded-md opacity-0 transition-opacity duration-200"
		style={backgroundStyle}
	/>
	{#each settings as settingsItem}
		<a
			class={cn(
				'relative inline-flex items-center justify-center whitespace-nowrap rounded px-3 pb-4 pt-2 text-center text-sm duration-200',
				$page.url.pathname.startsWith(settingsItem.href)
					? 'text-foreground after:border-foreground after:bg-foreground after:absolute after:bottom-0 after:left-0 after:h-0.5 after:w-full after:border-b after:transition-all after:duration-200 after:ease-in-out after:content-[""]'
					: 'text-muted-foreground hover:text-foreground'
			)}
			href={settingsItem.href}
		>
			{settingsItem.title}
		</a>
	{/each}
</nav>

<style>
	nav:hover > .background {
		opacity: 1;
	}
</style>
