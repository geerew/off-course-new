<script lang="ts">
	import { Logo } from '$lib/components';
	import { FormRegister } from '$lib/components/form';
</script>

<div class="flex h-full min-h-110 flex-col items-center justify-center gap-8">
	<Logo />

	<div class="flex min-w-sm flex-col gap-5">
		<div class="mb-2.5 flex flex-col gap-2">
			<div class="text-foreground-alt-1 text-center text-lg">Create administrator account</div>
		</div>

		<FormRegister endpoint="/api/auth/bootstrap" />
	</div>
</div>
