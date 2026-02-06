<script>
    import { goto } from "$app/navigation";
    import { addToast } from "$lib/stores/toast";
    import { onMount } from "svelte";
    import { Eye, EyeOff, Lock, ArrowRight, Loader2, AlertCircle } from "lucide-svelte";
    import confetti from "canvas-confetti";

    let password = "";
    let loading = false;
    let showPassword = false;
    let shake = false;
    let errorMsg = "";
    let waitTime = "";

    async function login() {
        if (!password) return;
        loading = true;
        errorMsg = "";
        
        try {
            const res = await fetch("/api/auth/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ password }),
            });

            const data = await res.json();

            if (res.ok) {
                confetti({
                    particleCount: 100,
                    spread: 70,
                    origin: { y: 0.6 }
                });
                setTimeout(() => {
                   window.location.href = "/";
                }, 800);
            } else {
                errorMsg = data.error || "Login failed";
                if (data.wait) {
                    waitTime = data.wait;
                }
                if (data.remaining !== undefined) {
                    errorMsg += ` (${data.remaining} attempts remaining)`;
                }
                
                triggerShake();
                addToast(errorMsg, "error");
            }
        } catch (e) {
            errorMsg = "Network error occurred";
            triggerShake();
            addToast("Network error", "error");
        } finally {
            loading = false;
        }
    }

    function triggerShake() {
        shake = true;
        setTimeout(() => shake = false, 500);
    }

    onMount(async () => {
        const res = await fetch("/api/auth/status");
        if (res.ok) {
            const status = await res.json();
            if (status.authenticated) {
                goto("/");
            }
        }
    });
</script>

<div class="min-h-screen flex items-center justify-center bg-[#050510] p-4 relative overflow-hidden font-sans">
    <!-- Animated background elements -->
    <div class="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none">
        <div class="absolute top-[-20%] right-[-10%] w-[800px] h-[800px] bg-indigo-500/10 rounded-full blur-[120px] animate-pulse"></div>
        <div class="absolute bottom-[-20%] left-[-10%] w-[600px] h-[600px] bg-cyan-500/10 rounded-full blur-[100px] opacity-70"></div>
    </div>

    <div class="w-full max-w-md relative z-10 perspective-1000">
        <div 
            class="bg-gray-900/40 backdrop-blur-xl rounded-2xl border border-white/10 p-8 shadow-2xl relative overflow-hidden transition-transform duration-300"
            class:shake={shake}
        >
            <!-- Top accent line -->
            <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-indigo-500 to-transparent opacity-50"></div>

            <div class="text-center mb-8">
                <div class="w-16 h-16 bg-indigo-500/10 rounded-2xl flex items-center justify-center mx-auto mb-6 transform rotate-3 border border-indigo-500/20 shadow-lg shadow-indigo-500/5 group">
                    <Lock class="w-8 h-8 text-indigo-400 group-hover:scale-110 transition-transform duration-300" />
                </div>
                <h1 class="text-3xl font-bold text-white mb-2 tracking-tight">Welcome Back</h1>
                <p class="text-gray-400 text-sm">Enter your administrative password to access the panel</p>
            </div>

            {#if errorMsg}
                <div class="mb-6 p-3 rounded-lg bg-red-500/10 border border-red-500/20 flex items-start gap-3 text-red-400 text-sm animate-in fade-in slide-in-from-top-2">
                    <AlertCircle class="w-5 h-5 shrink-0" />
                    <p>{errorMsg}</p>
                </div>
            {/if}

            <form on:submit|preventDefault={login} class="space-y-6">
                <div class="space-y-2">
                    <label for="password" class="block text-xs font-medium text-gray-500 uppercase tracking-widest pl-1">
                        Password
                    </label>
                    <div class="relative group">
                        <input
                            type={showPassword ? "text" : "password"}
                            id="password"
                            bind:value={password}
                            class="w-full bg-black/20 border border-white/10 rounded-xl py-3.5 pl-4 pr-12 text-white placeholder-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500/50 focus:bg-black/40 transition-all duration-300"
                            placeholder="••••••••"
                        />
                        <button 
                            type="button"
                            class="absolute right-3 top-1/2 -translate-y-1/2 p-1.5 text-gray-500 hover:text-gray-300 transition-colors rounded-lg hover:bg-white/5"
                            on:click={() => showPassword = !showPassword}
                        >
                            {#if showPassword}
                                <EyeOff class="w-4 h-4" />
                            {:else}
                                <Eye class="w-4 h-4" />
                            {/if}
                        </button>
                    </div>
                </div>

                <button
                    type="submit"
                    disabled={loading || !password}
                    class="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-medium py-3.5 px-4 rounded-xl transition-all duration-300 transform hover:translate-y-[-1px] active:translate-y-[1px] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0 shadow-lg shadow-indigo-600/20 hover:shadow-indigo-600/40 flex items-center justify-center gap-2 group"
                >
                    {#if loading}
                        <Loader2 class="w-5 h-5 animate-spin" />
                        <span>Authenticating...</span>
                    {:else}
                        <span>Access Panel</span>
                        <ArrowRight class="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                    {/if}
                </button>
            </form>
            
            <div class="mt-8 text-center">
                <p class="text-[10px] text-gray-600 font-mono">
                    JJMC Server Manager • Secure Access
                </p>
            </div>
        </div>
    </div>
</div>

<style>
    .shake {
        animation: shake 0.5s cubic-bezier(.36,.07,.19,.97) both;
    }

    @keyframes shake {
        10%, 90% { transform: translate3d(-1px, 0, 0); }
        20%, 80% { transform: translate3d(2px, 0, 0); }
        30%, 50%, 70% { transform: translate3d(-4px, 0, 0); }
        40%, 60% { transform: translate3d(4px, 0, 0); }
    }

    .perspective-1000 {
        perspective: 1000px;
    }
</style>
