<script lang="ts">
	import { Checkbox } from '$components/ui/checkbox';
	import { createEventDispatcher } from 'svelte';
	import { type Writable } from 'svelte/store';

	export let value: string;
	export let checked: Writable<boolean>;

	const dispatch = createEventDispatcher();
</script>

<Checkbox
	on:click={() => {
		dispatch('click', value);
	}}
	class="border-alt-1 data-[state=checked]:border-primary"
	bind:checked={$checked}
/>
