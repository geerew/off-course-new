<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { AlertDialog, Separator } from 'bits-ui';

	let deleteDialogOpen = $state(false);

	function deleteAccount() {
		auth.delete().finally(() => {
			deleteDialogOpen = false;
		});
	}
</script>

{#if auth.user !== null}
	<div class="mx-auto flex max-w-2xl flex-col place-content-center items-start gap-5">
		<div class="flex flex-col gap-3">
			<div class="text-foreground-alt-2 text-sm uppercase">Username</div>
			<span class="text-background-primary text-2xl">{auth.user.username}</span>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<div class="flex flex-col gap-3">
			<div class="text-foreground-alt-2 text-sm uppercase">Display Name</div>
			<span class="text-background-primary text-2xl">{auth.user.displayName}</span>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<div class="flex flex-col gap-3">
			<div class="text-foreground-alt-2 text-sm uppercase">Delete Account</div>
			<AlertDialog.Root bind:open={deleteDialogOpen}>
				<AlertDialog.Trigger
					class="bg-background-error hover:bg-background-error-alt-1 text-foreground-alt-1 hover:text-foreground w-36 cursor-pointer rounded-md py-2 duration-200 select-none"
				>
					Delete Account
				</AlertDialog.Trigger>

				<AlertDialog.Portal>
					<AlertDialog.Overlay
						class="bg-background/70 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50"
					/>

					<!--  -->
					<AlertDialog.Content
						interactOutsideBehavior="close"
						class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:zoom-in-95  data-[state=closed]:zoom-out-95 data-[state=open]:slide-in-from-top-5 bg-background-alt-1 data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-top-5 fixed top-20 left-1/2 z-50 w-full max-w-[26rem] min-w-[20rem] -translate-x-1/2 overflow-hidden rounded-lg data-[state=closed]:duration-200 data-[state=open]:duration-200 md:max-w-lg"
					>
						<div class="flex flex-col gap-2.5 p-5">
							<div class="flex items-center justify-center">
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke-width="1.5"
									stroke="currentColor"
									class="text-foreground-error size-14"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
									/>
								</svg>
							</div>

							<AlertDialog.Description
								class="text-foreground-alt-1 flex flex-col gap-2 text-center"
							>
								<span class="text-lg">Are you sure you want to delete your account?</span>
								<span class="text-foreground-alt-2">All associated data will be deleted</span>
							</AlertDialog.Description>
						</div>

						<div
							class="bg-background-alt-2 border-background-alt-3 flex w-full items-center justify-end gap-2 border-t px-5 py-2.5"
						>
							<AlertDialog.Cancel
								class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none"
							>
								Cancel
							</AlertDialog.Cancel>

							<AlertDialog.Action
								onclick={deleteAccount}
								class="bg-background-error hover:bg-background-error-alt-1 text-foreground-alt-1 hover:text-foreground w-24 cursor-pointer rounded-md py-2 duration-200 select-none"
							>
								Delete
							</AlertDialog.Action>
						</div>
					</AlertDialog.Content>
				</AlertDialog.Portal>
			</AlertDialog.Root>
		</div>
	</div>
{/if}
