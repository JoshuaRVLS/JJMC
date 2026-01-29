<script>
	import "../app.css";
	import { page } from "$app/stores";
	import { onMount, onDestroy } from "svelte";
	import { goto } from "$app/navigation";
	import Toast from "$lib/components/Toast.svelte";
	import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
	import InputModal from "$lib/components/InputModal.svelte";

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

	async function checkAuth() {
		try {
			const res = await fetch("/api/auth/status");
			if (res.ok) {
				const status = await res.json();
				const path = window.location.pathname;

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
					pollInterval = setInterval(loadInstances, 5000);
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
		<svg
			class="animate-spin h-8 w-8 text-indigo-500"
			xmlns="http://www.w3.org/2000/svg"
			fill="none"
			viewBox="0 0 24 24"
		>
			<circle
				class="opacity-25"
				cx="12"
				cy="12"
				r="10"
				stroke="currentColor"
				stroke-width="4"
			></circle>
			<path
				class="opacity-75"
				fill="currentColor"
				d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
			></path>
		</svg>
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
							GLOBAL MANAGER
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
								<span
									class="mr-2 group-hover:-translate-x-1 transition-transform"
									>&larr;</span
								>
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
								<svg
									class="w-4 h-4"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
									/></svg
								>
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
								<svg
									class="w-4 h-4"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
									/></svg
								>
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
								<svg
									class="w-4 h-4"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
									/><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
									/></svg
								>
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
								<svg
									class="w-4 h-4"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
									/></svg
								>
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
								<svg
									class="w-4 h-4"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
									/></svg
								>
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
								<svg
									class="w-4 h-4"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
									/><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
									/></svg
								>
								Settings
							</a>
							{#if currentInstanceType !== "spigot"}
								<a
									href="/instances/{currentInstanceId}?tab=mods"
									class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
										'tab',
									) === 'mods'
										? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
										: 'hover:bg-white/5 hover:text-white'}"
								>
									<svg
										class="w-4 h-4"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
										><path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
										/></svg
									>
									Mods
								</a>
							{:else}
								<a
									href="/instances/{currentInstanceId}?tab=plugins"
									class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
										'tab',
									) === 'plugins'
										? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
										: 'hover:bg-white/5 hover:text-white'}"
								>
									<svg
										class="w-4 h-4"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
										><path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
										/></svg
									>
									Plugins
								</a>
							{/if}
							<a
								href="/instances/{currentInstanceId}?tab=configs"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'configs'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<svg
									class="w-4 h-4"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
									/><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
									/></svg
								>
								Configs
							</a>
							<a
								href="/instances/{currentInstanceId}?tab=type"
								class="flex items-center gap-3 px-3 py-2 rounded-lg transition-all duration-200 {$page.url.searchParams.get(
									'tab',
								) === 'type'
									? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20'
									: 'hover:bg-white/5 hover:text-white'}"
							>
								<svg
									class="w-4 h-4"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
									/><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"
									/></svg
								>
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
									<svg
										class="w-4 h-4"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
										><path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"
										/></svg
									>
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
							<svg
								class="w-5 h-5"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
								><path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
								></path></svg
							>
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
				/>
				<div
					class="absolute bottom-[-20%] left-[-10%] w-[600px] h-[600px] bg-cyan-900/10 rounded-full blur-3xl opacity-30"
				/>
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
