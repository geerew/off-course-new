<script lang="ts">
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import { cn } from '$lib/utils';
	import { Button, DropdownMenu, Separator } from 'bits-ui';
	import { Logo } from '.';
	import { RightChevron } from './icons';

	const menu = [
		{
			label: 'Courses',
			href: '/courses',
			matcher: '/courses/'
		}
	];

	function logout() {
		auth.logout();
	}
</script>

<header>
	<div class="flex items-center justify-between py-6" aria-label="Global">
		<!-- Logo -->
		<div class="flex flex-1">
			<a href="/" class="-m-1.5 p-1.5">
				<Logo size="small" />
			</a>
		</div>

		<!-- Menu -->
		<nav class="flex gap-x-12">
			{#each menu as item}
				<a
					href={item.href}
					class={cn(
						'text-foreground-alt-1 hover:text-foreground relative rounded-lg px-2.5 py-1.5 leading-6 font-semibold duration-200',
						page.url.pathname === item.matcher &&
							'after:bg-background-primary after:absolute after:-bottom-0.5 after:left-0 after:h-0.5 after:w-full'
					)}
					aria-current={page.url.pathname === item.matcher}
				>
					{item.label}
				</a>
			{/each}
		</nav>

		{#if auth.user !== null}
			<!-- Profile -->
			<div class="flex flex-1 justify-end">
				<DropdownMenu.Root>
					<DropdownMenu.Trigger
						class="bg-background-primary-alt-1 hover:bg-background-primary text-foreground-alt-3 relative flex size-10 cursor-pointer items-center justify-center rounded-full font-semibold"
					>
						{auth.userLetter}
					</DropdownMenu.Trigger>

					<DropdownMenu.Portal>
						<DropdownMenu.Content
							sideOffset={8}
							align={'end'}
							class="bg-background-alt-1 text-foreground-alt-1 border-background-alt-3 data-[state=open]:animate-in [state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 z-30 w-54 rounded-lg border p-2.5"
						>
							<div class="flex flex-col select-none">
								<!-- Name -->
								<div class="flex flex-row items-center gap-3 p-1.5">
									<span
										class="bg-background-primary text-foreground-alt-3 relative flex size-10 items-center justify-center rounded-full font-semibold"
									>
										{auth.userLetter}
									</span>
									<span class="text-base font-semibold tracking-wide">
										{auth.user.username}
									</span>
								</div>

								<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

								<div class="flex flex-col gap-2">
									<!-- Profile link -->
									<DropdownMenu.Item>
										<Button.Root
											href="/profile"
											class="hover:bg-background-alt-3 hover:text-foreground flex cursor-pointer flex-row items-center justify-between rounded-lg p-1.5 duration-200"
										>
											<div class="flex flex-row items-center gap-3">
												<svg
													xmlns="http://www.w3.org/2000/svg"
													fill="none"
													viewBox="0 0 24 24"
													stroke-width="1.5"
													stroke="currentColor"
													class="size-5"
												>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z"
													/>
												</svg>

												<span>Profile</span>
											</div>

											<RightChevron class="size-4" />
										</Button.Root>
									</DropdownMenu.Item>

									<!-- Admin link -->
									{#if auth.user.role === 'admin'}
										<DropdownMenu.Item>
											<Button.Root
												href="/admin"
												class="hover:bg-background-alt-3 hover:text-foreground flex cursor-pointer flex-row items-center justify-between rounded-lg p-1.5 duration-200"
											>
												<div class="flex flex-row items-center gap-3">
													<svg
														xmlns="http://www.w3.org/2000/svg"
														fill="none"
														viewBox="0 0 24 24"
														stroke-width="1.5"
														stroke="currentColor"
														class="size-5"
													>
														<path
															stroke-linecap="round"
															stroke-linejoin="round"
															d="M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z"
														/>
													</svg>

													<span>Admin</span>
												</div>

												<RightChevron class="size-4" />
											</Button.Root>
										</DropdownMenu.Item>
									{/if}

									<!-- Logout link-->
									<DropdownMenu.Item>
										<Button.Root
											onclick={logout}
											class="hover:bg-background-error hover:text-foreground flex w-full cursor-pointer flex-row items-center justify-between rounded-lg p-1.5 duration-200"
										>
											<div class="flex flex-row items-center gap-3">
												<svg
													xmlns="http://www.w3.org/2000/svg"
													fill="none"
													viewBox="0 0 24 24"
													stroke-width="1.5"
													stroke="currentColor"
													class="size-5"
												>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														d="M8.25 9V5.25A2.25 2.25 0 0 1 10.5 3h6a2.25 2.25 0 0 1 2.25 2.25v13.5A2.25 2.25 0 0 1 16.5 21h-6a2.25 2.25 0 0 1-2.25-2.25V15m-3 0-3-3m0 0 3-3m-3 3H15"
													/>
												</svg>

												<span>Logout</span>
											</div>

											<RightChevron class="size-4" />
										</Button.Root>
									</DropdownMenu.Item>
								</div>
							</div>
						</DropdownMenu.Content>
					</DropdownMenu.Portal>
				</DropdownMenu.Root>
			</div>
		{/if}
	</div>
</header>
