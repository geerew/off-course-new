<script lang="ts">
	import { Loading } from '$components';
	import { Icons } from '$components/icons';
	import { addToast } from '$lib/stores/addToast';
	import type { FileInfo, FileSystem } from '$lib/types/fileSystem';
	import { AddCourse, ErrorMessage, GetAllCourses, GetFileSystem } from '$lib/utils/api';
	import { createDialog } from '@melt-ui/svelte';
	import { createEventDispatcher } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import BackButton from './internal/AddCoursesBackButton.svelte';
	import AddCoursesRow from './internal/AddCoursesRow.svelte';

	// ----------------------
	// Variables
	// ----------------------

	// Dialog builder
	const {
		elements: { trigger, overlay, content, title, close, portalled },
		states: { open }
	} = createDialog();

	// Dispatcher. This is triggered when the user clicks the add button
	const dispatch = createEventDispatcher();

	// True when the initial loading of the drives/volumes is done
	let loadingDrives = false;

	// True when a path is loading and is used to disable clicking on other paths
	let loadingPath = false;

	// True when the API call errors
	let gotError = false;

	// True when the refresh button is click. By default, this is true
	let refreshing = false;

	// Holds the information about this path, such as files and directories
	let pathInfo: FileSystem;

	// This is bound to the content element and used to reset the scroll position to the top
	// following navigation
	let body: HTMLElement;

	// A sequential list of paths. As the used navigates through the filesystem, the path is added
	// to this array. When the back button is clicked, the last path is popped from the array and
	// the user is taken to that path.
	let paths: string[] = [];

	// An array of the selected courses. This is updated as the user selects/unselects courses
	let selectedCourses: Record<string, string> = {};

	// This is used by the select all button. As courses are manually selected they will be removed
	// from this array and when they are manually unselected they will be added back to this array.
	// When the used clicks the select all button, the courses in this array will be added to the
	// selected courses array
	let selectableCourses: FileInfo[] = [];

	// True when we are required to pull the existing courses from the backend
	let getExistingCourses = true;

	// An array of the existing courses. This is pulled from the backend and used to disable
	// existing courses from being selected/unselected
	let existingCourses: string[] = [];

	// ----------------------
	// Reactive
	// ----------------------

	$: if ($open) {
		// Reset some stuff when the dialog is opened. It is better to do it here as during
		// development this triggers on hot reload
		paths = [];
		selectedCourses = {};
		selectableCourses = [];
		loadingDrives = false;
		loadingPath = false;
		gotError = false;
		refreshing = false;
		getExistingCourses = true;

		// Load the drives when first opened
		(async () => await load('', false, false))();
	}

	$: disableAddButton =
		loadingDrives ||
		loadingPath ||
		refreshing ||
		gotError ||
		Object.keys(selectedCourses).length === 0;

	// ----------------------
	// Functions
	// ----------------------

	// Load will load the drives or information about a path, based upon when path is populated.
	//
	// When path is empty, the drives will be loaded and when path is populated, information about
	// the path will be loaded, such as files and directories.
	//
	// When movingBack is true, the last path will be popped from the paths array. When false, the
	// path will be added to the paths array.
	//
	// When refresh is true, the paths array will not be manipulated
	const load = async (path: string, movingBack: boolean, refresh: boolean) => {
		path ? (loadingPath = true) : (loadingDrives = true);

		// Pull the existing courses so we can identify already added courses. This is used to stop
		// the user for selecting/unselecting existing courses
		if (getExistingCourses) {
			const success = await GetAllCourses({})
				.then((courses) => {
					existingCourses.push(...courses.map((c) => c.path));
					getExistingCourses = false;
					return true;
				})
				.catch((err) => {
					const errMsg = ErrorMessage(err);

					gotError = true;
					loadingDrives = false;
					console.error(errMsg);

					$addToast({
						data: {
							message: errMsg,
							status: 'error'
						}
					});

					return false;
				});

			if (!success) return;
		}

		// Pull the filesystem information for this path. When the path is empty it will load
		// drive information
		await GetFileSystem(path)
			.then((resp) => {
				if (body) body.scrollTop = 0;

				if (resp) {
					const selectedKeys = Object.keys(selectedCourses);
					const allPaths = selectedKeys.concat(existingCourses);

					resp.directories?.forEach((d) => {
						// Mark any existing courses
						if (existingCourses.find((c) => c === d.path)) d.isExistingCourse = true;

						// Mark parents. These are courses that contain a selected course
						if (isParent(d.path, allPaths)) d.isParent = true;

						// Mark selected courses
						if (selectedKeys.find((c) => c === d.path)) d.isSelected = true;
					});

					// Set the selectable courses. This will be a list of directories that are not
					// existing courses, selected courses, or parents.
					selectableCourses = resp?.directories.filter(
						({ isSelected, isExistingCourse, isParent }) =>
							!isExistingCourse && !isSelected && !isParent
					);

					pathInfo = resp;

					// Set the loading state to false
					path ? (loadingPath = false) : (loadingDrives = false);

					// When not part of a refresh, manipulate the paths array
					if (!refresh) {
						if (movingBack) {
							// Pop the last path as we are moving up a directory
							paths.pop();
							paths = [...paths];
						} else if (path) {
							// Add the path to the array as we are moving into a directory
							paths = [...paths, path];
						}
					}
				}
			})
			.catch((err) => {
				const errMsg = ErrorMessage(err);
				if (path) {
					console.error(errMsg);
					loadingPath = false;
					// TODO: Go back to previous path
				} else {
					gotError = true;
					loadingDrives = false;
				}

				$addToast({
					data: {
						message: errMsg,
						status: 'error'
					}
				});
			});
	};

	// Returns turn when a path is the parent of another path, as in, this parent path contains a
	// selected course somewhere in its directory structure
	const isParent = (path: string, paths: string[]) => {
		for (let i = 0; i < paths.length; i++) {
			if (paths[i] === path) return false;
			else if (paths[i].startsWith(path)) return true;
		}

		return false;
	};

	// Select all courses that are currently not selected
	function selectAll() {
		// Loop over all selectable courses and add them to the selected courses array if they
		// are not already selected
		let doReact = false;
		selectableCourses.forEach((course) => {
			if (!(course.path in selectedCourses)) {
				doReact = true;
				selectedCourses[course.path] = course.title;
			}
		});

		// When we added courses to the selected courses array, manually mark these directories as
		// selected
		if (doReact) {
			pathInfo.directories.forEach((d) => {
				if (selectableCourses.find((c) => c.path === d.path)) d.isSelected = true;
			});

			pathInfo = { ...pathInfo };
		}
	}

	// Unselect all courses
	function unselectAll() {
		// Loop over all selectable courses and remove them from the selected courses array
		let doReact = false;
		selectableCourses.forEach((course) => {
			if (course.path in selectedCourses) {
				doReact = true;
				delete selectedCourses[course.path];
			}
		});

		// When we removed courses to the selected courses array, manually mark these directories as
		// unselected
		if (doReact) {
			pathInfo.directories.forEach((d) => {
				if (selectableCourses.find((c) => c.path === d.path)) d.isSelected = false;
			});

			pathInfo = { ...pathInfo };

			// Update the selected courses array to trigger a re-render
			selectedCourses = { ...selectedCourses };
		}
	}
</script>

<button {...$trigger} use:trigger class="bg-primary text-primary-foreground hover:brightness-110">
	<Icons.bookPlus class="h-4 w-4" />
	<span>Add Courses</span>
</button>

<div {...$portalled} use:portalled>
	{#if $open}
		<div
			{...$overlay}
			use:overlay
			transition:fade|local={{ duration: 150 }}
			class="fixed inset-0 bg-black/80 backdrop-blur-sm"
			tabindex="-1"
		/>

		<div
			class="fixed left-0 top-0 z-50 flex h-full w-full flex-row overflow-hidden"
			transition:fly|local={{ x: 200, duration: 400 }}
		>
			<div
				class="bg-background relative ml-auto flex h-full w-[700px] max-w-full flex-col self-end text-sm"
				{...$content}
				use:content
			>
				<!-- Header -->
				<div
					{...$title}
					use:title
					class="bg-background-alt-1 flex h-16 shrink-0 items-center justify-between border-b px-3 text-base font-medium"
				>
					<div class="flex items-center gap-2">
						<Icons.bookPlus class="h-4 w-4" />
						<span>Course Selection</span>
					</div>

					<button
						disabled={loadingDrives || loadingPath}
						class="enabled:hover:bg-accent-1 group rounded-md px-2 py-1.5 text-sm font-semibold duration-200"
						on:click={async () => {
							if (refreshing) return;
							refreshing = true;
							const currentPath = paths[paths.length - 1] ?? '';
							await load(currentPath, false, true);
							refreshing = false;
						}}
					>
						<Icons.refresh
							class="group-hover:text-foreground text-foreground-muted h-5 w-5 duration-200"
						/>
					</button>
				</div>

				<!-- Body -->
				<div
					bind:this={body}
					class="flex min-h-[15rem] grow flex-col overflow-y-scroll"
					tabindex="-1"
				>
					{#if loadingDrives || refreshing}
						<div class="py-10">
							<Loading class="border-primary" />
						</div>
					{:else if gotError}
						<div class="text-error flex w-full justify-center py-10 font-semibold">
							There was an error reading the filesystem.
						</div>
					{:else}
						<div class="flex flex-col">
							<!-- Back button -->
							{#if paths.length > 0}
								{#key paths[paths.length - 1]}
									<BackButton
										bind:loadingPath
										on:click={async () => {
											// When moving back, load the drives if we are only 1
											// level deep, or else load the path before this one
											let backPath = '';
											if (paths.length > 1) backPath = paths[paths.length - 2];
											await load(backPath, true, false);
										}}
									/>
								{/key}
							{/if}

							<!-- Directories -->
							{#each pathInfo.directories as data (data.path)}
								<AddCoursesRow
									{data}
									bind:loadingPath
									on:add={() => {
										selectedCourses[data.path] = data.title;
									}}
									on:remove={() => {
										delete selectedCourses[data.path];
										selectedCourses = { ...selectedCourses };
									}}
									on:click={async () => {
										// Moving into a directory
										await load(data.path, false, false);
									}}
								/>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Footer -->
				<div
					class="bg-background-alt-1 flex h-16 shrink-0 items-center justify-between border-t px-3 py-4 font-medium"
				>
					<div class="flex gap-3">
						<!-- Select all -->
						<button
							disabled={loadingDrives || loadingPath || refreshing}
							class="hover:bg-accent-1 border"
							on:click={selectAll}
						>
							Select All
						</button>

						<!-- Unselect all -->
						<button
							disabled={loadingDrives || loadingPath || refreshing}
							on:click={unselectAll}
							class="hover:bg-accent-1 border"
						>
							Unselect All
						</button>
					</div>

					<div class="flex gap-3">
						<button {...$close} use:close class="hover:bg-accent-1 border">Close</button>

						<button
							{...$close}
							use:close
							disabled={disableAddButton}
							class="bg-primary text-primary-foreground disabled:bg-accent-1 min-w-[6rem] enabled:hover:brightness-110"
							on:click={async () => {
								let sawError = false;
								const keys = Object.keys(selectedCourses);
								for (let i = 0; i < keys.length; i++) {
									await AddCourse({ title: selectedCourses[keys[i]], path: keys[i] }).catch(
										(err) => {
											console.error(err);
											sawError = true;
										}
									);
								}

								!sawError &&
									$addToast({
										data: {
											message: `Course${keys.length > 1 ? 's' : ''} added`,
											status: 'success'
										}
									});

								dispatch('added');
							}}
						>
							Add ({Object.keys(selectedCourses).length})
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<style lang="postcss">
	button {
		@apply inline-flex select-none items-center justify-center gap-2 whitespace-nowrap rounded px-3 py-1.5 text-center text-sm duration-200;
	}
</style>
