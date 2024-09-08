<script lang="ts">
	import * as Drawer from '$lib/components/ui/drawer';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { onMount } from 'svelte';
	import theme from 'tailwindcss/defaultTheme';
	import { MediaRemoteControl } from 'vidstack';
	import { isSettingsMenuOpen, preferences } from '../store';
	import Playback from './_settings/playback.svelte';
	import Trigger from './_settings/trigger.svelte';

	// ----------------------
	// Exports
	// ----------------------

	export let isMobile: boolean;

	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();

	// The settings menu element
	let menuEl: HTMLDivElement;

	// Whether the menu is open
	let open = false;

	// The current section of the menu
	let section: 'top' | 'playback' = 'top';

	// The breakpoint for md
	const mdPx = +theme.screens.md.replace('px', '');

	// ----------------------
	// Reactive
	// ----------------------

	// Toggle the idle tracking of controls
	function controls(open: boolean) {
		if (open) {
			// Update the video ctx to mark settings as open
			remote.pauseControls();
			isSettingsMenuOpen.set(true);
		} else {
			// Update the video ctx to mark settings as closed and resume idle tracking
			remote.resumeControls();
			isSettingsMenuOpen.set(false);
		}
	}

	// ----------------------
	// Reactive
	// ----------------------
	$: controls(open);

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		// Find the player
		const player = remote.getPlayer(menuEl);
		if (!player) return;

		// Set the playback rate
		const playbackRateUnsub = player.subscribe(({ playbackRate }) => {
			preferences.set({ ...$preferences, playbackRate: playbackRate });
		});

		// Unsubscribe
		return () => {
			playbackRateUnsub();
			remote.resumeControls();
		};
	});
</script>

<!-- Close the settings menu if the window is resized to sm- -->
<svelte:window
	on:resize={() => {
		if (window.innerWidth < mdPx) open = false;
	}}
/>

<div bind:this={menuEl} class="inline-flex">
	{#if isMobile}
		<Drawer.Root
			bind:open
			portal={null}
			onOpenChange={(o) => {
				if (!o) section = 'top';
			}}
		>
			<Drawer.Trigger asChild let:builder>
				<Trigger {builder} />
			</Drawer.Trigger>

			<Drawer.Content class="mx-auto max-h-[70%] min-h-28 max-w-sm">
				<div class="mx-auto mb-2 mt-4 h-2 w-[100px] shrink-0 rounded-full bg-muted"></div>
				<div class="flex h-full w-full flex-col overflow-y-auto px-2.5 pt-5" data-vaul-no-drag="">
					<Playback
						show={section === 'playback'}
						on:close={() => {
							section = 'top';
						}}
						on:open={() => {
							section = 'playback';
						}}
					/>
				</div>
			</Drawer.Content>
		</Drawer.Root>
	{:else}
		<DropdownMenu.Root
			bind:open
			portal={null}
			closeOnItemClick={false}
			typeahead={false}
			preventScroll={false}
			onOpenChange={(o) => {
				if (!o) section = 'top';
			}}
		>
			<DropdownMenu.Trigger asChild let:builder>
				<Trigger {builder} />
			</DropdownMenu.Trigger>

			<DropdownMenu.Content class="min-w-56 p-3 text-sm font-light" side="top" align="end">
				<Playback
					show={section === 'playback'}
					on:close={() => {
						section = 'top';
					}}
					on:open={() => {
						section = 'playback';
					}}
				/>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	{/if}
</div>
