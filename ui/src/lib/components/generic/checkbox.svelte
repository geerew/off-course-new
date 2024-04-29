<script lang="ts">
	import { Checkbox } from '$components/ui/checkbox';
	import { type Writable } from 'svelte/store';

	export let selected: Writable<Record<string, string>>;
	export let id: string;

	$: checked = ($selected[id] && true) || false;
</script>

<Checkbox on:click class="border-alt-1" bind:checked />
