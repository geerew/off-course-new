<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Check, ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { MediaRemoteControl } from 'vidstack';
	import { getCtx } from './context';
	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();

	// The settings menu element
	let menuEl: HTMLDivElement;

	let open = false;

	// The current section of the settings menu
	let section: 'top' | 'speed' = 'top';

	// The current playback rate the the available playback rates
	let playbackRate = 1;
	let playbackRates = [0.5, 0.75, 1, 1.25, 1.5, 1.75, 2, 4];

	// Get the video ctx
	const ctx = getCtx();

	// ----------------------
	// Reactive
	// ----------------------

	// Toggle the idle tracking of controls
	function controls(open: boolean) {
		if (open) {
			// Update the video ctx to mark settings as open
			ctx.set({ ...$ctx, settingsOpen: true });
			remote.pauseControls();
		} else {
			// Update the video ctx to mark settings as closed and resume idle tracking (if required)
			ctx.set({ ...$ctx, settingsOpen: false });
			if (!$ctx.controlsOpen && !$ctx.settingsOpen) remote.resumeControls();
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
		const playbackRateUnsub = player.subscribe(({ playbackRate: playerPlaybackRate }) => {
			playbackRate = playerPlaybackRate;
		});

		// Hide the menu when the controls are hidden
		const controlsUnsub = player.subscribe(({ controlsVisible }) => {
			if (!controlsVisible) open = false;
		});

		// Unsubscribe
		return () => {
			playbackRateUnsub();
			controlsUnsub();
		};
	});
</script>

<div bind:this={menuEl} class="inline-flex">
	<DropdownMenu.Root
		bind:open
		closeOnItemClick={false}
		onOpenChange={(o) => {
			if (!o) section = 'top';
		}}
	>
		<DropdownMenu.Trigger asChild let:builder>
			<Button
				builders={[builder]}
				variant="ghost"
				class="hover:text-secondary inline-flex h-auto cursor-pointer items-center bg-black px-0 py-0 text-white hover:bg-black"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 24 24"
					fill="currentColor"
					class="size-[18px]"
				>
					<path
						fill-rule="evenodd"
						d="M11.078 2.25c-.917 0-1.699.663-1.85 1.567L9.05 4.889c-.02.12-.115.26-.297.348a7.493 7.493 0 0 0-.986.57c-.166.115-.334.126-.45.083L6.3 5.508a1.875 1.875 0 0 0-2.282.819l-.922 1.597a1.875 1.875 0 0 0 .432 2.385l.84.692c.095.078.17.229.154.43a7.598 7.598 0 0 0 0 1.139c.015.2-.059.352-.153.43l-.841.692a1.875 1.875 0 0 0-.432 2.385l.922 1.597a1.875 1.875 0 0 0 2.282.818l1.019-.382c.115-.043.283-.031.45.082.312.214.641.405.985.57.182.088.277.228.297.35l.178 1.071c.151.904.933 1.567 1.85 1.567h1.844c.916 0 1.699-.663 1.85-1.567l.178-1.072c.02-.12.114-.26.297-.349.344-.165.673-.356.985-.57.167-.114.335-.125.45-.082l1.02.382a1.875 1.875 0 0 0 2.28-.819l.923-1.597a1.875 1.875 0 0 0-.432-2.385l-.84-.692c-.095-.078-.17-.229-.154-.43a7.614 7.614 0 0 0 0-1.139c-.016-.2.059-.352.153-.43l.84-.692c.708-.582.891-1.59.433-2.385l-.922-1.597a1.875 1.875 0 0 0-2.282-.818l-1.02.382c-.114.043-.282.031-.449-.083a7.49 7.49 0 0 0-.985-.57c-.183-.087-.277-.227-.297-.348l-.179-1.072a1.875 1.875 0 0 0-1.85-1.567h-1.843ZM12 15.75a3.75 3.75 0 1 0 0-7.5 3.75 3.75 0 0 0 0 7.5Z"
						clip-rule="evenodd"
					/>
				</svg>
			</Button>
		</DropdownMenu.Trigger>

		{#if section === 'top'}
			<DropdownMenu.Content class="w-44" side="top" align="end">
				<DropdownMenu.Item
					class="flex w-full cursor-pointer items-center justify-between"
					on:click={() => {
						section = 'speed';
					}}
				>
					<span class="font-semibold leading-3">Speed</span>
					<div class="flex items-center gap-1.5">
						<span class="text-muted-foreground">
							{playbackRate === 1 ? 'Normal' : playbackRate + 'x'}
						</span>
						<ChevronRight class="size-4" />
					</div>
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		{:else if section === 'speed'}
			<DropdownMenu.Content class="w-32" side="top" align="end">
				<DropdownMenu.Item
					class="flex w-full cursor-pointer items-center gap-2.5"
					on:click={() => {
						section = 'top';
					}}
				>
					<ChevronLeft class="size-4" />
					<span>Speed</span>
				</DropdownMenu.Item>

				<DropdownMenu.Separator />

				<DropdownMenu.Group>
					{#each playbackRates as rate}
						<DropdownMenu.Item
							class="flex w-full cursor-pointer items-center justify-between"
							on:click={() => {
								remote.changePlaybackRate(rate);
								section = 'top';
							}}
						>
							<span>
								{rate === 1 ? 'Normal' : rate + 'x'}
							</span>

							{#if rate === playbackRate}
								<Check class="text-secondary size-4" />
							{/if}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Group>
			</DropdownMenu.Content>
		{/if}
	</DropdownMenu.Root>
</div>
