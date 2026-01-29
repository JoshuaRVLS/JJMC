<script>
    import { goto } from "$app/navigation";
    import { addToast } from "$lib/stores/toast";
    import { onMount } from "svelte";

    let password = "";
    let confirmPassword = "";
    let loading = false;

    async function setup() {
        if (!password || !confirmPassword) {
            addToast("Please fill in all fields", "error");
            return;
        }

        if (password.length < 8) {
            addToast("Password must be at least 8 characters", "error");
            return;
        }

        if (password !== confirmPassword) {
            addToast("Passwords do not match", "error");
            return;
        }

        loading = true;
        try {
            const res = await fetch("/api/auth/setup", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ password }),
            });

            if (res.ok) {
                addToast("Setup complete!", "success");
                window.location.href = "/";
            } else {
                const data = await res.json();
                addToast(data.error || "Setup failed", "error");
            }
        } catch (e) {
            addToast("Network error", "error");
        } finally {
            loading = false;
        }
    }

    onMount(async () => {
        const res = await fetch("/api/auth/status");
        if (res.ok) {
            const status = await res.json();
            if (status.isSetup) {
                goto("/login");
            }
        }
    });
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-950 p-4">
    <div
        class="w-full max-w-md bg-gray-900 rounded-lg shadow-xl border border-gray-800 p-8"
    >
        <div class="mb-8 text-center">
            <h1 class="text-3xl font-bold text-white mb-2">Welcome</h1>
            <p class="text-gray-400">
                Create an administrator password to secure your server manager
            </p>
        </div>

        <form on:submit|preventDefault={setup} class="space-y-6">
            <div>
                <label
                    for="password"
                    class="block text-sm font-medium text-gray-400 mb-2"
                    >Password</label
                >
                <input
                    type="password"
                    id="password"
                    bind:value={password}
                    class="w-full bg-gray-800 border border-gray-700 rounded-md py-3 px-4 text-white focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all"
                    placeholder="Min. 8 characters"
                    autofocus
                />
            </div>

            <div>
                <label
                    for="confirm"
                    class="block text-sm font-medium text-gray-400 mb-2"
                    >Confirm Password</label
                >
                <input
                    type="password"
                    id="confirm"
                    bind:value={confirmPassword}
                    class="w-full bg-gray-800 border border-gray-700 rounded-md py-3 px-4 text-white focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all"
                    placeholder="Repeat password"
                />
            </div>

            <button
                type="submit"
                disabled={loading}
                class="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-semibold py-3 px-4 rounded-md transition-all flex items-center justify-center disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {#if loading}
                    <svg
                        class="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
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
                    Finish Setup
                {:else}
                    Create Account
                {/if}
            </button>
        </form>
    </div>
</div>
