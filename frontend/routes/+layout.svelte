<script>
	import "../app.css";
	import { page } from "$app/stores";
	import { onMount, onDestroy } from "svelte";
	import { goto } from "$app/navigation";
	import Toast from "$lib/components/Toast.svelte";
	import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
	import InputModal from "$lib/components/InputModal.svelte";
	import {
		Terminal,
		Folder,
		Settings2,
		Scroll,
		Shield,
		Settings,
		Puzzle,
		FileCog,
		Server,
		LayoutGrid,
		LogOut,
		ArrowLeft,
		Loader2,
		Archive,
	} from "lucide-svelte";

	/**
	 * @typedef {Object} Instance
	 * @property {string} id
	 * @property {string} name
	 * @property {string} type
	 * @property {string} status
	 */

	/** @type {Instance[]} */
	let instances = [];
	/** @type {ReturnType<typeof setInterval> | undefined} */
	let pollInterval;

	async function loadInstances() {
		try {
			const res = await fetch("/api/instances");
			if (res.ok) {
				instances = await res.json();
			}
		} catch (e) {
			console.error("Failed to load instances", e);
		}
	}

	// Auth State
	let isAuthChecked = false;

	// Launch ID for auto-reload
	let storedLaunchId = null;

	async function checkAuth() {
		try {
			// Add timestamp to prevent caching
			const res = await fetch("/api/auth/status?t=" + Date.now());
			if (res.ok) {
				const status = await res.json();
				const path = window.location.pathname;

				// Check Launch ID
				if (storedLaunchId === null) {
					storedLaunchId = status.launchId;
				} else if (storedLaunchId !== status.launchId) {
					console.log("New version detected, reloading...");
					window.location.reload();
					return;
				}

				if (!status.isSetup) {
					if (path !== "/setup") {
						await goto("/setup");
					}
				} else if (!status.authenticated) {
					// Check if we are already on a public page to avoid loop
					if (path !== "/login" && path !== "/setup") {
						await goto("/login");
					}
				} else {
					// Authenticated
					if (path === "/login" || path === "/setup") {
						await goto("/");
					}
					// Only load instances if authenticated
					loadInstances();

					// Setup polling if not already set
					if (!pollInterval) {
						pollInterval = setInterval(() => {
							loadInstances();
							checkAuth(); // Poll auth/status for version check
						}, 5000);
					}
				}
			} else {
				console.error("Auth status failed", res.status);
			}
		} catch (e) {
			console.error("Auth check failed", e);
		} finally {
			isAuthChecked = true;
		}
	}

	onMount(() => {
		console.log("JJMC Frontend Loaded - Force Refresh Check");

		// Force unregister any service workers
		if ("serviceWorker" in navigator) {
			navigator.serviceWorker
				.getRegistrations()
				.then(function (registrations) {
					for (let registration of registrations) {
						console.log(
							"Unregistering Service Worker:",
							registration,
						);
						registration.unregister();
					}
				});
		}

		checkAuth();
	});

	onDestroy(() => {
		if (pollInterval) clearInterval(pollInterval);
	});

	$: instanceMatch = $page.url.pathname.match(/^\/instances\/([^/]+)/);
	$: currentInstanceId =
		instanceMatch && instanceMatch[1] !== "create"
			? instanceMatch[1]
			: null;

	// Find name and type if we have the list
	$: currentInstanceObj = instances.find((i) => i.id === currentInstanceId);
	$: currentInstanceName = currentInstanceObj?.name || currentInstanceId;
	$: currentInstanceType = currentInstanceObj?.type;

	$: isPublicPage =
		$page.url.pathname === "/login" || $page.url.pathname === "/setup";
</script>

{#if !isAuthChecked}
	<div class="h-screen bg-gray-950 flex items-center justify-center">
		<!-- Loading Spinner -->
		<!-- Loading Spinner -->
		<Loader2 class="animate-spin h-8 w-8 text-indigo-500" />
	</div>
{:else}
	<div
		class="h-screen bg-gray-950 text-gray-300 flex font-sans text-sm overflow-hidden selection:bg-indigo-500/30 selection:text-indigo-200"
	>
		<!-- Sidebar -->
		{#if !isPublicPage}
			<aside
				class="w-64 bg-gray-900/50 backdrop-blur-xl flex flex-col border-r border-white/5"
			>
				<div class="p-6">
					{#if currentInstanceId}
						<h1
							class="font-bold text-white tracking-tight text-lg break-words truncate"
						>
							{currentInstanceName}
						</h1>
						<div
							class="text-[10px] font-bold uppercase tracking-widest text-indigo-400 mt-1"
						>
							{currentInstanceType === "spigot"
								? "Spigot Server"
								: "Instance Manager"}
						</div>
					{:else}
						<h1
							class="text-xl text-white tracking-tighter"
							style="font-family: 'Press Start 2P', cursive; text-shadow: 2px 2px 0px #4f46e5;"
						>
							JJMC
						</h1>
						<div class="text-[10px] font-bold text-gray-500 mt-1">
							Minecraft Server Manager
						</div>
					{/if}
				</div>

				<nav class="flex-1 px-3 space-y-1 overflow-y-auto">
					{#if currentInstanceId}
						<!-- Instance Context Menu -->
						<div class="mb-4">
							<a
								href="/instances"
								class="flex items-center px-3 py-2 text-xs font-medium text-gray-500 hover:text-white transition-colors group"
							>
								<ArrowLeft
									class="w-4 h-4 mr-2 group-hover:-translate-x-1 transition-transform"
								/>
								Back to Dashboard
							</a>
						</div>

						<div class="space-y-1">
							<div
								class="px-3 text-[10px] font-bold text-gray-600 uppercase tracking-widest mb-2"
							>
								Control
							</div>
							<a
								href="/instances/{currentInstanceId}"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page
									.url.pathname ===
									`/instances/${currentInstanceId}` &&
								(!$page.url.searchParams.get('tab') ||
									$page.url.searchParams.get('tab') ===
										'console')
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Terminal class="w-4 h-4" />
								Console
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=files"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'files'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Folder class="w-4 h-4" />
								Files
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=properties"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'properties'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Settings2 class="w-4 h-4" />
								Properties
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=whitelist"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'whitelist'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Scroll class="w-4 h-4" />
								Whitelist
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=ops"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'ops'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Shield class="w-4 h-4" />
								Ops
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=settings"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'settings'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Settings class="w-4 h-4" />
								Settings
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=mods"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'mods'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Puzzle class="w-4 h-4" />
								Mods
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=plugins"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'plugins'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Puzzle class="w-4 h-4" />
								Plugins
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=configs"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'configs'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<FileCog class="w-4 h-4" />
								Configs
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=backups"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'backups'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Archive class="w-4 h-4" />
								Backups
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=type"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'type'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<Server class="w-4 h-4" />
								Server Type
							</a>
						</div>
					{:else}
						<!-- Global Context Menu -->
						<div class="space-y-6">
							<div>
								<a
									href="/instances"
									class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page
										.url.pathname === '/instances'
										? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
										: 'hover:bg-white/5 hover:text-white'}"
								>
									<LayoutGrid class="w-4 h-4" />
									All Instances
								</a>
							</div>
						</div>
					{/if}
				</nav>

				<div class="p-6 border-t border-white/5">
					<div class="flex items-center justify-between gap-3">
						<div class="flex items-center gap-3"></div>
						<button
							on:click={async () => {
								await fetch("/api/auth/logout", {
									method: "POST",
								});
								window.location.href = "/login";
							}}
							class="text-gray-500 hover:text-white transition-colors"
							title="Logout"
						>
							<LogOut class="w-5 h-5" />
						</button>
					</div>
				</div>
			</aside>
		{/if}

		<!-- Main Content -->
		<main class="flex-1 flex flex-col min-h-0 relative">
			<!-- Main Background Glow -->
			<div
				class="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none"
			>
				<div
					class="absolute top-[-20%] right-[-10%] w-[800px] h-[800px] bg-indigo-900/10 rounded-full blur-3xl opacity-50"
				></div>
				<div
					class="absolute bottom-[-20%] left-[-10%] w-[600px] h-[600px] bg-cyan-900/10 rounded-full blur-3xl opacity-30"
				></div>
			</div>

			<!-- Content Slot -->
			<div class="relative z-10 flex-1 flex flex-col min-h-0">
				<slot />
			</div>
		</main>

		<!-- Global Overlays -->
		<Toast />
		<ConfirmDialog />
		<InputModal />
	</div>
{/if}
