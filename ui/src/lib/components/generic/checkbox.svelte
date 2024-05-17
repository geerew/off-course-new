<script lang="ts">
	import { Checkbox } from '$components/ui/checkbox';
	import { type Writable } from 'svelte/store';

	export let checked: Writable<boolean>;
</script>

<Checkbox
	on:click
	class="border-alt-1 data-[state=checked]:border-primary"
	bind:checked={$checked}
/>
