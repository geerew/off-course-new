<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Check, ChevronLeft, ChevronRight, Settings } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { MediaRemoteControl } from 'vidstack';

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
		const constrolsUnsub = player.subscribe(({ controlsVisible }) => {
			if (!controlsVisible) open = false;
		});

		// Unsubscribe
		return () => {
			playbackRateUnsub();
			constrolsUnsub();
		};
	});
</script>

<div bind:this={menuEl}>
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
				class="ring-media-focus hover:bg-secondary inline-flex h-auto cursor-pointer items-center justify-center rounded-md p-1.5 outline-none ring-inset data-[focus]:ring-4"
			>
				<Settings class="size-[18px] text-white" />
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
				<DropdownMenu.Arrow />
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
				<DropdownMenu.Arrow />
			</DropdownMenu.Content>
		{/if}
	</DropdownMenu.Root>
</div>
