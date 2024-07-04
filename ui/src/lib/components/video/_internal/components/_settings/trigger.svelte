<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import type { Builder } from 'bits-ui';

	// ----------------------
	// Exports
	// ----------------------
	export let builder: Builder;
</script>

<Button
	builders={[builder]}
	variant="ghost"
	class="hover:text-secondary group relative inline-flex h-auto cursor-pointer items-center px-0 py-0 text-white hover:bg-transparent"
>
	<Icons.Settings
		weight="fill"
		class="group-hover:fill-secondary size-6 fill-white duration-200 group-data-[state=open]:rotate-90"
	/>
</Button>
